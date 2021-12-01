// SPDX-License-Identifier: MIT

pragma solidity ^0.6.6;

import "@openzeppelin/contracts/math/SafeMath.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/SafeERC20.sol";
import "@openzeppelin/contracts/utils/Address.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

pragma experimental ABIEncoderV2;

// This is being used purely to avoid stack too deep errors
struct LogicCallArgs {
	// Transfers out to the logic contract
	uint256[] transferAmounts;
	address[] transferTokenContracts;
	// The fees (transferred to msg.sender)
	uint256[] feeAmounts;
	address[] feeTokenContracts;
	// The arbitrary logic call
	address logicContractAddress;
	bytes payload;
	// Invalidation metadata
	uint256 timeOut;
	bytes32 invalidationId;
	uint256 invalidationNonce;
}

interface IWETH {
    function deposit() external payable;
    function withdraw(uint) external;
}

contract Hub2 is ReentrancyGuard {
	using SafeMath for uint256;
	using SafeMath for uint;
	using SafeERC20 for IERC20;

	// These are updated often
	bytes32 public state_lastValsetCheckpoint;
	mapping(address => uint256) public state_lastBatchNonces;
	mapping(bytes32 => uint256) public state_invalidationMapping;
	uint256 public state_lastValsetNonce = 0;
	// event nonce zero is reserved by the Cosmos module as a special
	// value indicating that no events have yet been submitted
	uint256 public state_lastEventNonce = 1;

	// These are set once at initialization
	bytes32 public state_gravityId;
	uint256 public state_powerThreshold;

	address public wethAddress;

	address public guardian;

	// TransactionBatchExecutedEvent and TransferToChain both include the field _eventNonce.
	// This is incremented every time one of these events is emitted. It is checked by the
	// Cosmos module to ensure that all events are received in order, and that none are lost.
	//
	// ValsetUpdatedEvent does not include the field _eventNonce because it is never submitted to the Cosmos
	// module. It is purely for the use of relayers to allow them to successfully submit batches.
	event TransactionBatchExecutedEvent(
		uint256 indexed _batchNonce,
		address indexed _token,
		uint256 _eventNonce
	);
	event TransferToChainEvent(
		address indexed _tokenContract,
		address indexed _sender,
		bytes32 indexed _destinationChain,
		bytes32 _destination,
		uint256 _amount,
		uint256 _fee,
		uint256 _eventNonce
	);
	event ValsetUpdatedEvent(
		uint256 indexed _newValsetNonce,
		uint256 _eventNonce,
		address[] _validators,
		uint256[] _powers
	);
	event LogicCallEvent(
		bytes32 _invalidationId,
		uint256 _invalidationNonce,
		uint256 _eventNonce,
		bytes _returnData
	);

	receive() external payable {
        assert(msg.sender == wethAddress); // only accept ETH via fallback from the WETH contract
    }

	function lastBatchNonce(address _erc20Address) public view returns (uint256) {
		return state_lastBatchNonces[_erc20Address];
	}

	function lastLogicCallNonce(bytes32 _invalidation_id) public view returns (uint256) {
		return state_invalidationMapping[_invalidation_id];
	}

	// Utility function to verify geth style signatures
	function verifySig(
		address _signer,
		bytes32 _theHash,
		uint8 _v,
		bytes32 _r,
		bytes32 _s
	) private pure returns (bool) {
		bytes32 messageDigest =
			keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", _theHash));
		return _signer == ecrecover(messageDigest, _v, _r, _s);
	}

	// Make a new checkpoint from the supplied validator set
	// A checkpoint is a hash of all relevant information about the valset. This is stored by the contract,
	// instead of storing the information directly. This saves on storage and gas.
	// The format of the checkpoint is:
	// h(gravityId, "checkpoint", valsetNonce, validators[], powers[])
	// Where h is the keccak256 hash function.
	// The validator powers must be decreasing or equal. This is important for checking the signatures on the
	// next valset, since it allows the caller to stop verifying signatures once a quorum of signatures have been verified.
	function makeCheckpoint(
		address[] memory _validators,
		uint256[] memory _powers,
		uint256 _valsetNonce,
		bytes32 _gravityId
	) private pure returns (bytes32) {
		// bytes32 encoding of the string "checkpoint"
		bytes32 methodName = 0x636865636b706f696e7400000000000000000000000000000000000000000000;

		bytes32 checkpoint =
			keccak256(abi.encode(_gravityId, methodName, _valsetNonce, _validators, _powers));

		return checkpoint;
	}

	function checkValidatorSignatures(
		// The current validator set and their powers
		address[] memory _currentValidators,
		uint256[] memory _currentPowers,
		// The current validator's signatures
		uint8[] memory _v,
		bytes32[] memory _r,
		bytes32[] memory _s,
		// This is what we are checking they have signed
		bytes32 _theHash,
		uint256 _powerThreshold
	) private pure {
		uint256 cumulativePower = 0;

		for (uint256 i = 0; i < _currentValidators.length; i++) {
			// If v is set to 0, this signifies that it was not possible to get a signature from this validator and we skip evaluation
			// (In a valid signature, it is either 27 or 28)
			if (_v[i] != 0) {
				// Check that the current validator has signed off on the hash
				require(
					verifySig(_currentValidators[i], _theHash, _v[i], _r[i], _s[i]),
					"Validator signature does not match."
				);

				// Sum up cumulative power
				cumulativePower = cumulativePower + _currentPowers[i];

				// Break early to avoid wasting gas
				if (cumulativePower > _powerThreshold) {
					break;
				}
			}
		}

		// Check that there was enough power
		require(
			cumulativePower > _powerThreshold,
			"Submitted validator set signatures do not have enough power."
		);
		// Success
	}

	// This updates the valset by checking that the validators in the current valset have signed off on the
	// new valset. The signatures supplied are the signatures of the current valset over the checkpoint hash
	// generated from the new valset.
	// Anyone can call this function, but they must supply valid signatures of state_powerThreshold of the current valset over
	// the new valset.
	function updateValset(
		// The new version of the validator set
		address[] memory _newValidators,
		uint256[] memory _newPowers,
		uint256 _newValsetNonce,
		// The current validators that approve the change
		address[] memory _currentValidators,
		uint256[] memory _currentPowers,
		uint256 _currentValsetNonce,
		// These are arrays of the parts of the current validator's signatures
		uint8[] memory _v,
		bytes32[] memory _r,
		bytes32[] memory _s
	) public nonReentrant {
		// CHECKS

		// Check that the valset nonce is greater than the old one
		require(
			_newValsetNonce > _currentValsetNonce,
			"New valset nonce must be greater than the current nonce"
		);

		// Check that new validators and powers set is well-formed
		require(_newValidators.length == _newPowers.length, "Malformed new validator set");

		// Check that current validators, powers, and signatures (v,r,s) set is well-formed
		require(
			_currentValidators.length == _currentPowers.length &&
				_currentValidators.length == _v.length &&
				_currentValidators.length == _r.length &&
				_currentValidators.length == _s.length,
			"Malformed current validator set"
		);

		// Check that the supplied current validator set matches the saved checkpoint
		require(
			makeCheckpoint(
				_currentValidators,
				_currentPowers,
				_currentValsetNonce,
				state_gravityId
			) == state_lastValsetCheckpoint,
			"Supplied current validators and powers do not match checkpoint."
		);

		// Check that enough current validators have signed off on the new validator set
		bytes32 newCheckpoint =
			makeCheckpoint(_newValidators, _newPowers, _newValsetNonce, state_gravityId);

		checkValidatorSignatures(
			_currentValidators,
			_currentPowers,
			_v,
			_r,
			_s,
			newCheckpoint,
			state_powerThreshold
		);

		// ACTIONS

		// Stored to be used next time to validate that the valset
		// supplied by the caller is correct.
		state_lastValsetCheckpoint = newCheckpoint;

		// Store new nonce
		state_lastValsetNonce = _newValsetNonce;

		// LOGS
		state_lastEventNonce = state_lastEventNonce.add(1);
		emit ValsetUpdatedEvent(_newValsetNonce, state_lastEventNonce, _newValidators, _newPowers);
	}

	// submitBatch processes a batch of Cosmos -> Ethereum transactions by sending the tokens in the transactions
	// to the destination addresses. It is approved by the current Cosmos validator set.
	// Anyone can call this function, but they must supply valid signatures of state_powerThreshold of the current valset over
	// the batch.
	function submitBatch(
		// The validators that approve the batch
		address[] memory _currentValidators,
		uint256[] memory _currentPowers,
		uint256 _currentValsetNonce,
		// These are arrays of the parts of the validators signatures
		uint8[] memory _v,
		bytes32[] memory _r,
		bytes32[] memory _s,
		// The batch of transactions
		uint256[] memory _amounts,
		address payable[] memory _destinations,
		uint256[] memory _fees,
		uint256 _batchNonce,
		address _tokenContract,
		// a block height beyond which this batch is not valid
		// used to provide a fee-free timeout
		uint256 _batchTimeout
	) public nonReentrant {
		// CHECKS scoped to reduce stack depth
		{
			// Check that the batch nonce is higher than the last nonce for this token
			require(
				state_lastBatchNonces[_tokenContract] < _batchNonce,
				"New batch nonce must be greater than the current nonce"
			);

			// Check that the block height is less than the timeout height
			require(
				block.number < _batchTimeout,
				"Batch timeout must be greater than the current block height"
			);

			// Check that current validators, powers, and signatures (v,r,s) set is well-formed
			require(
				_currentValidators.length == _currentPowers.length &&
					_currentValidators.length == _v.length &&
					_currentValidators.length == _r.length &&
					_currentValidators.length == _s.length,
				"Malformed current validator set"
			);

			// Check that the supplied current validator set matches the saved checkpoint
			require(
				makeCheckpoint(
					_currentValidators,
					_currentPowers,
					_currentValsetNonce,
					state_gravityId
				) == state_lastValsetCheckpoint,
				"Supplied current validators and powers do not match checkpoint."
			);

			// Check that the transaction batch is well-formed
			require(
				_amounts.length == _destinations.length && _amounts.length == _fees.length,
				"Malformed batch of transactions"
			);

			// Check that enough current validators have signed off on the transaction batch and valset
			checkValidatorSignatures(
				_currentValidators,
				_currentPowers,
				_v,
				_r,
				_s,
				// Get hash of the transaction batch and checkpoint
				keccak256(
					abi.encode(
						state_gravityId,
						// bytes32 encoding of "transactionBatch"
						0x7472616e73616374696f6e426174636800000000000000000000000000000000,
						_amounts,
						_destinations,
						_fees,
						_batchNonce,
						_tokenContract,
						_batchTimeout
					)
				),
				state_powerThreshold
			);

			// ACTIONS

			// Store batch nonce
			state_lastBatchNonces[_tokenContract] = _batchNonce;

			// Send transaction amounts to destinations
			if (_tokenContract == wethAddress) {
				for (uint256 i = 0; i < _amounts.length; i++) {
					IWETH(wethAddress).withdraw(_amounts[i]);
					TransferHelper.safeTransferETH(_destinations[i], _amounts[i]);
				}
			} else {
				for (uint256 i = 0; i < _amounts.length; i++) {
					IERC20(_tokenContract).safeTransfer(_destinations[i], _amounts[i]);
				}
			}
		}

		// LOGS scoped to reduce stack depth
		{
			state_lastEventNonce = state_lastEventNonce.add(1);
			emit TransactionBatchExecutedEvent(_batchNonce, _tokenContract, state_lastEventNonce);
		}
	}

	// This makes calls to contracts that execute arbitrary logic
	// First, it gives the logic contract some tokens
	// Then, it gives msg.senders tokens for fees
	// Then, it calls an arbitrary function on the logic contract
	// invalidationId and invalidationNonce are used for replay prevention.
	// They can be used to implement a per-token nonce by setting the token
	// address as the invalidationId and incrementing the nonce each call.
	// They can be used for nonce-free replay prevention by using a different invalidationId
	// for each call.
	function submitLogicCall(
		// The validators that approve the call
		address[] memory _currentValidators,
		uint256[] memory _currentPowers,
		uint256 _currentValsetNonce,
		// These are arrays of the parts of the validators signatures
		uint8[] memory _v,
		bytes32[] memory _r,
		bytes32[] memory _s,
		LogicCallArgs memory _args
	) public nonReentrant {
		// CHECKS scoped to reduce stack depth
		{
			// Check that the call has not timed out
			require(block.number < _args.timeOut, "Timed out");

			// Check that the invalidation nonce is higher than the last nonce for this invalidation Id
			require(
				state_invalidationMapping[_args.invalidationId] < _args.invalidationNonce,
				"New invalidation nonce must be greater than the current nonce"
			);

			// Check that current validators, powers, and signatures (v,r,s) set is well-formed
			require(
				_currentValidators.length == _currentPowers.length &&
					_currentValidators.length == _v.length &&
					_currentValidators.length == _r.length &&
					_currentValidators.length == _s.length,
				"Malformed current validator set"
			);

			// Check that the supplied current validator set matches the saved checkpoint
			require(
				makeCheckpoint(
					_currentValidators,
					_currentPowers,
					_currentValsetNonce,
					state_gravityId
				) == state_lastValsetCheckpoint,
				"Supplied current validators and powers do not match checkpoint."
			);

			// Check that the token transfer list is well-formed
			require(
				_args.transferAmounts.length == _args.transferTokenContracts.length,
				"Malformed list of token transfers"
			);

			// Check that the fee list is well-formed
			require(
				_args.feeAmounts.length == _args.feeTokenContracts.length,
				"Malformed list of fees"
			);
		}

		bytes32 argsHash =
			keccak256(
				abi.encode(
					state_gravityId,
					// bytes32 encoding of "logicCall"
					0x6c6f67696343616c6c0000000000000000000000000000000000000000000000,
					_args.transferAmounts,
					_args.transferTokenContracts,
					_args.feeAmounts,
					_args.feeTokenContracts,
					_args.logicContractAddress,
					_args.payload,
					_args.timeOut,
					_args.invalidationId,
					_args.invalidationNonce
				)
			);

		{
			// Check that enough current validators have signed off on the transaction batch and valset
			checkValidatorSignatures(
				_currentValidators,
				_currentPowers,
				_v,
				_r,
				_s,
				// Get hash of the transaction batch and checkpoint
				argsHash,
				state_powerThreshold
			);
		}

		// ACTIONS

		// Update invaldiation nonce
		state_invalidationMapping[_args.invalidationId] = _args.invalidationNonce;

		// Send tokens to the logic contract
		for (uint256 i = 0; i < _args.transferAmounts.length; i++) {
			IERC20(_args.transferTokenContracts[i]).safeTransfer(
				_args.logicContractAddress,
				_args.transferAmounts[i]
			);
		}

		// Make call to logic contract
		bytes memory returnData = Address.functionCall(_args.logicContractAddress, _args.payload);

		// Send fees to msg.sender
		for (uint256 i = 0; i < _args.feeAmounts.length; i++) {
			IERC20(_args.feeTokenContracts[i]).safeTransfer(msg.sender, _args.feeAmounts[i]);
		}

		// LOGS scoped to reduce stack depth
		{
			state_lastEventNonce = state_lastEventNonce.add(1);
			emit LogicCallEvent(
				_args.invalidationId,
				_args.invalidationNonce,
				state_lastEventNonce,
				returnData
			);
		}
	}

	function transferToChain(
		address _tokenContract,
		bytes32 _destinationChain,
		bytes32 _destination,
		uint256 _amount,
		uint256 _fee
	) public nonReentrant {
		IERC20(_tokenContract).safeTransferFrom(msg.sender, address(this), _amount);
		state_lastEventNonce = state_lastEventNonce.add(1);
		emit TransferToChainEvent(
			_tokenContract,
			msg.sender,
			_destinationChain,
			_destination,
			_amount,
			_fee,
			state_lastEventNonce
		);
	}

	function transferETHToChain(
		bytes32 _destinationChain,
		bytes32 _destination,
		uint256 _fee
	) public nonReentrant payable {
		IWETH(wethAddress).deposit{value: msg.value}();
		state_lastEventNonce = state_lastEventNonce.add(1);
		emit TransferToChainEvent(
			wethAddress,
			msg.sender,
			_destinationChain,
			_destination,
			msg.value,
			_fee,
			state_lastEventNonce
		);
	}

	function changeGuardian(address _guardian) public {
		require(msg.sender == guardian, "permission denied");

		guardian = _guardian;
	}

	function panicHalt(address[] memory _tokenContracts, address _safeAddress) public {
		require(msg.sender == guardian, "permission denied");

		for (uint256 i = 0; i < _tokenContracts.length; i++) {
			IERC20 token = IERC20(_tokenContracts[i]);
			token.safeTransfer(_safeAddress, token.balanceOf(address(this)));
		}
	}

	constructor(
		// A unique identifier for this gravity instance to use in signatures
		bytes32 _gravityId,
		// How much voting power is needed to approve operations
		uint256 _powerThreshold,
		// The validator set
		address[] memory _validators,
		uint256[] memory _powers,
		address _wethAddress,
		address _guardian
	) public {
		// CHECKS

		// Check that validators, powers, and signatures (v,r,s) set is well-formed
		require(_validators.length == _powers.length, "Malformed current validator set");

		// Check cumulative power to ensure the contract has sufficient power to actually
		// pass a vote
		uint256 cumulativePower = 0;
		for (uint256 i = 0; i < _powers.length; i++) {
			cumulativePower = cumulativePower + _powers[i];
			if (cumulativePower > _powerThreshold) {
				break;
			}
		}
		require(
			cumulativePower > _powerThreshold,
			"Submitted validator set signatures do not have enough power."
		);

		bytes32 newCheckpoint = makeCheckpoint(_validators, _powers, 0, _gravityId);

		// ACTIONS

		state_gravityId = _gravityId;
		state_powerThreshold = _powerThreshold;
		state_lastValsetCheckpoint = newCheckpoint;

		wethAddress = _wethAddress;
		guardian = _guardian;

		// LOGS

		emit ValsetUpdatedEvent(state_lastValsetNonce, state_lastEventNonce, _validators, _powers);
	}
}

library TransferHelper {
    function safeApprove(address token, address to, uint value) internal {
        // bytes4(keccak256(bytes('approve(address,uint256)')));
        (bool success, bytes memory data) = token.call(abi.encodeWithSelector(0x095ea7b3, to, value));
        require(success && (data.length == 0 || abi.decode(data, (bool))), 'TransferHelper: APPROVE_FAILED');
    }

    function safeTransfer(address token, address to, uint value) internal {
        // bytes4(keccak256(bytes('transfer(address,uint256)')));
        (bool success, bytes memory data) = token.call(abi.encodeWithSelector(0xa9059cbb, to, value));
        require(success && (data.length == 0 || abi.decode(data, (bool))), 'TransferHelper: TRANSFER_FAILED');
    }

    function safeTransferFrom(address token, address from, address to, uint value) internal {
        // bytes4(keccak256(bytes('transferFrom(address,address,uint256)')));
        (bool success, bytes memory data) = token.call(abi.encodeWithSelector(0x23b872dd, from, to, value));
        require(success && (data.length == 0 || abi.decode(data, (bool))), 'TransferHelper: TRANSFER_FROM_FAILED');
    }

    function safeTransferETH(address to, uint value) internal {
        (bool success,) = to.call{value:value}(new bytes(0));
        require(success, 'TransferHelper: ETH_TRANSFER_FAILED');
    }
}
