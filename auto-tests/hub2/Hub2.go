// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package hub2

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// LogicCallArgs is an auto generated low-level Go binding around an user-defined struct.
type LogicCallArgs struct {
	TransferAmounts        []*big.Int
	TransferTokenContracts []common.Address
	FeeAmounts             []*big.Int
	FeeTokenContracts      []common.Address
	LogicContractAddress   common.Address
	Payload                []byte
	TimeOut                *big.Int
	InvalidationId         [32]byte
	InvalidationNonce      *big.Int
}

// Hub2MetaData contains all meta data concerning the Hub2 contract.
var Hub2MetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_gravityId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_powerThreshold\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_powers\",\"type\":\"uint256[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"_invalidationId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_invalidationNonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"_returnData\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_eventNonce\",\"type\":\"uint256\"}],\"name\":\"LogicCallEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_tokenContract\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"_destination\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_eventNonce\",\"type\":\"uint256\"}],\"name\":\"SendToHubEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_batchNonce\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_eventNonce\",\"type\":\"uint256\"}],\"name\":\"TransactionBatchExecutedEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_newValsetNonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_eventNonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"_powers\",\"type\":\"uint256[]\"}],\"name\":\"ValsetUpdatedEvent\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_erc20Address\",\"type\":\"address\"}],\"name\":\"lastBatchNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_invalidation_id\",\"type\":\"bytes32\"}],\"name\":\"lastLogicCallNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_tokenContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"_destination\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"sendToHub\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"state_gravityId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"state_invalidationMapping\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"state_lastBatchNonces\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"state_lastEventNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"state_lastValsetCheckpoint\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"state_lastValsetNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"state_powerThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_currentValidators\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_currentPowers\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"_currentValsetNonce\",\"type\":\"uint256\"},{\"internalType\":\"uint8[]\",\"name\":\"_v\",\"type\":\"uint8[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_s\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"addresspayable[]\",\"name\":\"_destinations\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_fees\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"_batchNonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_tokenContract\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_batchTimeout\",\"type\":\"uint256\"}],\"name\":\"submitBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_currentValidators\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_currentPowers\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"_currentValsetNonce\",\"type\":\"uint256\"},{\"internalType\":\"uint8[]\",\"name\":\"_v\",\"type\":\"uint8[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_s\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"uint256[]\",\"name\":\"transferAmounts\",\"type\":\"uint256[]\"},{\"internalType\":\"address[]\",\"name\":\"transferTokenContracts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"feeAmounts\",\"type\":\"uint256[]\"},{\"internalType\":\"address[]\",\"name\":\"feeTokenContracts\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"logicContractAddress\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"timeOut\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"invalidationId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"invalidationNonce\",\"type\":\"uint256\"}],\"internalType\":\"structLogicCallArgs\",\"name\":\"_args\",\"type\":\"tuple\"}],\"name\":\"submitLogicCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_currentValidators\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_currentPowers\",\"type\":\"uint256[]\"},{\"internalType\":\"uint8[]\",\"name\":\"_v\",\"type\":\"uint8[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_s\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32\",\"name\":\"_theHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_powerThreshold\",\"type\":\"uint256\"}],\"name\":\"testCheckValidatorSignatures\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_validators\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_powers\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"_valsetNonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_gravityId\",\"type\":\"bytes32\"}],\"name\":\"testMakeCheckpoint\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_newValidators\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_newPowers\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"_newValsetNonce\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"_currentValidators\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_currentPowers\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"_currentValsetNonce\",\"type\":\"uint256\"},{\"internalType\":\"uint8[]\",\"name\":\"_v\",\"type\":\"uint8[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"_s\",\"type\":\"bytes32[]\"}],\"name\":\"updateValset\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"wethAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6080604052600060045560016005553480156200001b57600080fd5b5060405162002663380380620026638339810160408190526200003e9162000250565b60016000558051825114620000705760405162461bcd60e51b815260040162000067906200044e565b60405180910390fd5b6000805b8251811015620000af578281815181106200008b57fe5b60200260200101518201915084821115620000a657620000af565b60010162000074565b50838111620000d25760405162461bcd60e51b81526004016200006790620003f1565b6000620000e2848483896200016a565b600687905560078690556001819055600880546001600160a01b03191673c02aaa39b223fe8d0a0e5c4f27ead9083c756cc217905560045460055460405192935090917fb119f1f36224601586b5037da909ecf37e83864dddea5d32ad4e32ac1d97e62b9162000156918890889062000485565b60405180910390a250505050505062000505565b6040516000906918da1958dadc1bda5b9d60b21b90829062000199908590849088908b908b90602001620003aa565b60408051808303601f190181529190528051602090910120979650505050505050565b80516001600160a01b0381168114620001d457600080fd5b92915050565b600082601f830112620001eb578081fd5b815162000202620001fc82620004e5565b620004be565b8181529150602080830190848101818402860182018710156200022457600080fd5b60005b84811015620002455781518452928201929082019060010162000227565b505050505092915050565b6000806000806080858703121562000266578384fd5b845160208087015160408801519296509450906001600160401b03808211156200028e578485fd5b818801915088601f830112620002a2578485fd5b8151620002b3620001fc82620004e5565b81815284810190848601868402860187018d1015620002d0578889fd5b8895505b83861015620002fe57620002e98d82620001bc565b835260019590950194918601918601620002d4565b5060608b0151909750945050508083111562000318578384fd5b50506200032887828801620001da565b91505092959194509250565b6000815180845260208085019450808401835b838110156200036e5781516001600160a01b03168752958201959082019060010162000347565b509495945050505050565b6000815180845260208085019450808401835b838110156200036e578151875295820195908201906001016200038c565b600086825285602083015284604083015260a06060830152620003d160a083018562000334565b8281036080840152620003e5818562000379565b98975050505050505050565b6020808252603c908201527f5375626d69747465642076616c696461746f7220736574207369676e6174757260408201527f657320646f206e6f74206861766520656e6f75676820706f7765722e00000000606082015260800190565b6020808252601f908201527f4d616c666f726d65642063757272656e742076616c696461746f722073657400604082015260600190565b600084825260606020830152620004a0606083018562000334565b8281036040840152620004b4818562000379565b9695505050505050565b6040518181016001600160401b0381118282101715620004dd57600080fd5b604052919050565b60006001600160401b03821115620004fb578081fd5b5060209081020190565b61214e80620005156000396000f3fe608060405234801561001057600080fd5b50600436106101005760003560e01c8063c227c30b11610097578063e08bf6ea11610066578063e08bf6ea146101e2578063e3cb9f62146101f5578063e5a2b5d214610208578063f2b533071461021057610100565b8063c227c30b14610196578063c9d194d5146101a9578063db7c4e57146101bc578063df97174b146101cf57610100565b80637dfb6f86116100d35780637dfb6f861461016057806383b435db14610173578063b56561fe14610186578063bdda81d41461018e57610100565b8063011b2174146101055780630c246c821461012e5780634f0e0ef31461014357806373b2054714610158575b600080fd5b6101186101133660046112db565b610218565b60405161012591906119c6565b60405180910390f35b61014161013c3660046116cc565b610233565b005b61014b61054b565b6040516101259190611975565b61011861055a565b61011861016e366004611854565b610560565b61014161018136600461152f565b610572565b6101186108c3565b6101186108c9565b6101416101a43660046117c4565b6108cf565b6101186101b7366004611854565b6108e2565b6101416101ca36600461132b565b6108f4565b6101186101dd3660046112db565b61090c565b6101416101f03660046112f7565b61091e565b610141610203366004611409565b6109bf565b610118610b29565b610118610b2f565b6001600160a01b031660009081526002602052604090205490565b6002600054141561025f5760405162461bcd60e51b815260040161025690612015565b60405180910390fd5b600260005560c081015143106102875760405162461bcd60e51b815260040161025690611d68565b61010081015160e0820151600090815260036020526040902054106102be5760405162461bcd60e51b815260040161025690611cd4565b855187511480156102d0575083518751145b80156102dd575082518751145b80156102ea575081518751145b6103065760405162461bcd60e51b815260040161025690611fde565b600154610317888888600654610b35565b146103345760405162461bcd60e51b815260040161025690611bea565b6020810151518151511461035a5760405162461bcd60e51b815260040161025690611e2b565b806060015151816040015151146103835760405162461bcd60e51b815260040161025690611bba565b6000600654681b1bd9da58d0d85b1b60ba1b836000015184602001518560400151866060015187608001518860a001518960c001518a60e001518b61010001516040516020016103dd9b9a99989796959493929190611a12565b604051602081830303815290604052805190602001209050610406888887878786600754610b87565b61010082015160e08301516000908152600360205260408120919091555b8251518110156104865761047e83608001518460000151838151811061044657fe5b60200260200101518560200151848151811061045e57fe5b60200260200101516001600160a01b0316610c7e9092919063ffffffff16565b600101610424565b50606061049b83608001518460a00151610cd9565b905060005b8360400151518110156104e1576104d933856040015183815181106104c157fe5b60200260200101518660600151848151811061045e57fe5b6001016104a0565b506005546104f0906001610d24565b600581905560e08401516101008501516040517f7c2bb24f8e1b3725cb613d7f11ef97d9745cc97a0e40f730621c052d684077a193610533939291869190611b59565b60405180910390a15050600160005550505050505050565b6008546001600160a01b031681565b60055481565b60036020526000908152604090205481565b600260005414156105955760405162461bcd60e51b815260040161025690612015565b600260008181556001600160a01b0384168152602091909152604090205483116105d15760405162461bcd60e51b815260040161025690611c47565b8043106105f05760405162461bcd60e51b815260040161025690611dce565b8a518c51148015610602575088518c51145b801561060f575087518c51145b801561061c575086518c51145b6106385760405162461bcd60e51b815260040161025690611fde565b6001546106498d8d8d600654610b35565b146106665760405162461bcd60e51b815260040161025690611bea565b84518651148015610678575083518651145b6106945760405162461bcd60e51b815260040161025690611c9d565b6106ee8c8c8b8b8b6006546f0e8e4c2dce6c2c6e8d2dedc84c2e8c6d60831b8d8d8d8d8d8d6040516020016106d0989796959493929190611ab8565b60405160208183030381529060405280519060200120600754610b87565b6001600160a01b03808316600081815260026020526040902085905560085490911614156108005760005b86518110156107fa5760085487516001600160a01b0390911690632e1a7d4d9089908490811061074557fe5b60200260200101516040518263ffffffff1660e01b815260040161076991906119c6565b600060405180830381600087803b15801561078357600080fd5b505af1158015610797573d6000803e3d6000fd5b505050508581815181106107a757fe5b60200260200101516001600160a01b03166108fc8883815181106107c757fe5b60200260200101519081150290604051600060405180830381858888f193505050506107f257600080fd5b600101610719565b5061085a565b60005b86518110156108585761085086828151811061081b57fe5b602002602001015188838151811061082f57fe5b6020026020010151856001600160a01b0316610c7e9092919063ffffffff16565b600101610803565b505b600554610868906001610d24565b60058190556040516001600160a01b0384169185917f02c7e81975f8edb86e2a0c038b7b86a49c744236abf0f6177ff5afc6986ab708916108a8916119c6565b60405180910390a35050600160005550505050505050505050565b60045481565b60065481565b6108db84848484610b35565b5050505050565b60009081526003602052604090205490565b61090387878787878787610b87565b50505050505050565b60026020526000908152604090205481565b600260005414156109415760405162461bcd60e51b815260040161025690612015565b600260005561095b6001600160a01b038416333084610d49565b600554610969906001610d24565b6005819055604051839133916001600160a01b038716917f8a90c92ae4d3ad99d0e0af0871264ba019979b4532daba15508810d224e8bbf6916109ad918791612081565b60405180910390a45050600160005550565b600260005414156109e25760405162461bcd60e51b815260040161025690612015565b6002600055838711610a065760405162461bcd60e51b815260040161025690611f00565b8751895114610a275760405162461bcd60e51b815260040161025690611e6c565b84518651148015610a39575082518651145b8015610a46575081518651145b8015610a53575080518651145b610a6f5760405162461bcd60e51b815260040161025690611fde565b600154610a80878787600654610b35565b14610a9d5760405162461bcd60e51b815260040161025690611bea565b6000610aad8a8a8a600654610b35565b9050610ac0878786868686600754610b87565b60018181556004899055600554610ad691610d24565b600581905560405189917fb119f1f36224601586b5037da909ecf37e83864dddea5d32ad4e32ac1d97e62b91610b1091908e908e9061204c565b60405180910390a2505060016000555050505050505050565b60075481565b60015481565b6040516000906918da1958dadc1bda5b9d60b21b908290610b62908590849088908b908b906020016119cf565b60408051601f198184030181529190528051602090910120925050505b949350505050565b6000805b8851811015610c5457868181518110610ba057fe5b602002602001015160ff16600014610c4c57610c0b898281518110610bc157fe5b602002602001015185898481518110610bd657fe5b6020026020010151898581518110610bea57fe5b6020026020010151898681518110610bfe57fe5b6020026020010151610d70565b610c275760405162461bcd60e51b815260040161025690611d8b565b878181518110610c3357fe5b60200260200101518201915082821115610c4c57610c54565b600101610b8b565b50818111610c745760405162461bcd60e51b815260040161025690611ea3565b5050505050505050565b610cd48363a9059cbb60e01b8484604051602401610c9d9291906119ad565b60408051601f198184030181529190526020810180516001600160e01b03166001600160e01b031990931692909217909152610e0b565b505050565b6060610d1b83836040518060400160405280601e81526020017f416464726573733a206c6f772d6c6576656c2063616c6c206661696c65640000815250610e9a565b90505b92915050565b600082820183811015610d1b5760405162461bcd60e51b815260040161025690611d31565b610d6a846323b872dd60e01b858585604051602401610c9d93929190611989565b50505050565b60008085604051602001610d849190611944565b60405160208183030381529060405280519060200120905060018186868660405160008152602001604052604051610dbf9493929190611b89565b6020604051602081039080840390855afa158015610de1573d6000803e3d6000fd5b505050602060405103516001600160a01b0316876001600160a01b03161491505095945050505050565b6060610e60826040518060400160405280602081526020017f5361666545524332303a206c6f772d6c6576656c2063616c6c206661696c6564815250856001600160a01b0316610e9a9092919063ffffffff16565b805190915015610cd45780806020019051810190610e7e9190611834565b610cd45760405162461bcd60e51b815260040161025690611f94565b6060610b7f84846000856060610eaf85610f68565b610ecb5760405162461bcd60e51b815260040161025690611f5d565b60006060866001600160a01b03168587604051610ee89190611928565b60006040518083038185875af1925050503d8060008114610f25576040519150601f19603f3d011682016040523d82523d6000602084013e610f2a565b606091505b50915091508115610f3e579150610b7f9050565b805115610f4e5780518082602001fd5b8360405162461bcd60e51b81526004016102569190611ba7565b6000813f7fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470818114801590610b7f575050151592915050565b8035610d1e81612100565b600082601f830112610fbc578081fd5b8135610fcf610fca826120b5565b61208f565b818152915060208083019084810181840286018201871015610ff057600080fd5b60005b8481101561101857813561100681612100565b84529282019290820190600101610ff3565b505050505092915050565b600082601f830112611033578081fd5b8135611041610fca826120b5565b81815291506020808301908481018184028601820187101561106257600080fd5b60005b8481101561101857813561107881612100565b84529282019290820190600101611065565b600082601f83011261109a578081fd5b81356110a8610fca826120b5565b8181529150602080830190848101818402860182018710156110c957600080fd5b60005b84811015611018578135845292820192908201906001016110cc565b600082601f8301126110f8578081fd5b8135611106610fca826120b5565b81815291506020808301908481018184028601820187101561112757600080fd5b6000805b8581101561115557823560ff81168114611143578283fd5b8552938301939183019160010161112b565b50505050505092915050565b600082601f830112611171578081fd5b81356001600160401b03811115611186578182fd5b611199601f8201601f191660200161208f565b91508082528360208285010111156111b057600080fd5b8060208401602084013760009082016020015292915050565b60006101208083850312156111dc578182fd5b6111e58161208f565b91505081356001600160401b03808211156111ff57600080fd5b61120b8583860161108a565b8352602084013591508082111561122157600080fd5b61122d85838601610fac565b6020840152604084013591508082111561124657600080fd5b6112528583860161108a565b6040840152606084013591508082111561126b57600080fd5b61127785838601610fac565b60608401526112898560808601610fa1565b608084015260a08401359150808211156112a257600080fd5b506112af84828501611161565b60a08301525060c082013560c082015260e082013560e082015261010080830135818301525092915050565b6000602082840312156112ec578081fd5b8135610d1b81612100565b60008060006060848603121561130b578182fd5b833561131681612100565b95602085013595506040909401359392505050565b600080600080600080600060e0888a031215611345578283fd5b87356001600160401b038082111561135b578485fd5b6113678b838c01610fac565b985060208a013591508082111561137c578485fd5b6113888b838c0161108a565b975060408a013591508082111561139d578485fd5b6113a98b838c016110e8565b965060608a01359150808211156113be578485fd5b6113ca8b838c0161108a565b955060808a01359150808211156113df578485fd5b506113ec8a828b0161108a565b93505060a0880135915060c0880135905092959891949750929550565b60008060008060008060008060006101208a8c031215611427578182fd5b89356001600160401b038082111561143d578384fd5b6114498d838e01610fac565b9a5060208c013591508082111561145e578384fd5b61146a8d838e0161108a565b995060408c0135985060608c0135915080821115611486578384fd5b6114928d838e01610fac565b975060808c01359150808211156114a7578384fd5b6114b38d838e0161108a565b965060a08c0135955060c08c01359150808211156114cf578384fd5b6114db8d838e016110e8565b945060e08c01359150808211156114f0578384fd5b6114fc8d838e0161108a565b93506101008c0135915080821115611512578283fd5b5061151f8c828d0161108a565b9150509295985092959850929598565b6000806000806000806000806000806000806101808d8f031215611551578586fd5b6001600160401b038d351115611565578586fd5b6115728e8e358f01610fac565b9b506001600160401b0360208e0135111561158b578586fd5b61159b8e60208f01358f0161108a565b9a5060408d013599506001600160401b0360608e013511156115bb578586fd5b6115cb8e60608f01358f016110e8565b98506001600160401b0360808e013511156115e4578586fd5b6115f48e60808f01358f0161108a565b97506001600160401b0360a08e0135111561160d578586fd5b61161d8e60a08f01358f0161108a565b96506001600160401b0360c08e01351115611636578586fd5b6116468e60c08f01358f0161108a565b95506001600160401b0360e08e0135111561165f578283fd5b61166f8e60e08f01358f01611023565b94506001600160401b036101008e01351115611689578283fd5b61169a8e6101008f01358f0161108a565b93506101208d013592506116b28e6101408f01610fa1565b91506101608d013590509295989b509295989b509295989b565b600080600080600080600060e0888a0312156116e6578081fd5b87356001600160401b03808211156116fc578283fd5b6117088b838c01610fac565b985060208a013591508082111561171d578283fd5b6117298b838c0161108a565b975060408a0135965060608a0135915080821115611745578283fd5b6117518b838c016110e8565b955060808a0135915080821115611766578283fd5b6117728b838c0161108a565b945060a08a0135915080821115611787578283fd5b6117938b838c0161108a565b935060c08a01359150808211156117a8578283fd5b506117b58a828b016111c9565b91505092959891949750929550565b600080600080608085870312156117d9578182fd5b84356001600160401b03808211156117ef578384fd5b6117fb88838901610fac565b95506020870135915080821115611810578384fd5b5061181d8782880161108a565b949794965050505060408301359260600135919050565b600060208284031215611845578081fd5b81518015158114610d1b578182fd5b600060208284031215611865578081fd5b5035919050565b6001600160a01b0316815260200190565b6001600160a01b03169052565b6000815180845260208085019450808401835b838110156118c25781516001600160a01b03168752958201959082019060010161189d565b509495945050505050565b6000815180845260208085019450808401835b838110156118c2578151875295820195908201906001016118e0565b600081518084526119148160208601602086016120d4565b601f01601f19169290920160200192915050565b6000825161193a8184602087016120d4565b9190910192915050565b7f19457468657265756d205369676e6564204d6573736167653a0a3332000000008152601c810191909152603c0190565b6001600160a01b0391909116815260200190565b6001600160a01b039384168152919092166020820152604081019190915260600190565b6001600160a01b03929092168252602082015260400190565b90815260200190565b600086825285602083015284604083015260a060608301526119f460a083018561188a565b8281036080840152611a0681856118cd565b98975050505050505050565b60006101608d83528c6020840152806040840152611a328184018d6118cd565b90508281036060840152611a46818c61188a565b90508281036080840152611a5a818b6118cd565b905082810360a0840152611a6e818a61188a565b6001600160a01b03891660c085015283810360e08501529050611a9181886118fc565b61010084019690965250506101208101929092526101409091015298975050505050505050565b60006101008a835260208a81850152816040850152611ad98285018b6118cd565b91508382036060850152818951611af081856119c6565b9150828b019350845b81811015611b1a57611b0c83865161186c565b948401949250600101611af9565b50508481036080860152611b2e818a6118cd565b93505050508460a0830152611b4660c083018561187d565b8260e08301529998505050505050505050565b600085825284602083015260806040830152611b7860808301856118fc565b905082606083015295945050505050565b93845260ff9290921660208401526040830152606082015260800190565b600060208252610d1b60208301846118fc565b6020808252601690820152754d616c666f726d6564206c697374206f66206665657360501b604082015260600190565b6020808252603f908201527f537570706c6965642063757272656e742076616c696461746f727320616e642060408201527f706f7765727320646f206e6f74206d6174636820636865636b706f696e742e00606082015260800190565b60208082526036908201527f4e6577206261746368206e6f6e6365206d7573742062652067726561746572206040820152757468616e207468652063757272656e74206e6f6e636560501b606082015260800190565b6020808252601f908201527f4d616c666f726d6564206261746368206f66207472616e73616374696f6e7300604082015260600190565b6020808252603d908201527f4e657720696e76616c69646174696f6e206e6f6e6365206d757374206265206760408201527f726561746572207468616e207468652063757272656e74206e6f6e6365000000606082015260800190565b6020808252601b908201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604082015260600190565b602080825260099082015268151a5b5959081bdd5d60ba1b604082015260600190565b60208082526023908201527f56616c696461746f72207369676e617475726520646f6573206e6f74206d617460408201526231b41760e91b606082015260800190565b6020808252603b908201527f42617463682074696d656f7574206d757374206265206772656174657220746860408201527f616e207468652063757272656e7420626c6f636b206865696768740000000000606082015260800190565b60208082526021908201527f4d616c666f726d6564206c697374206f6620746f6b656e207472616e736665726040820152607360f81b606082015260800190565b6020808252601b908201527f4d616c666f726d6564206e65772076616c696461746f72207365740000000000604082015260600190565b6020808252603c908201527f5375626d69747465642076616c696461746f7220736574207369676e6174757260408201527f657320646f206e6f74206861766520656e6f75676820706f7765722e00000000606082015260800190565b60208082526037908201527f4e65772076616c736574206e6f6e6365206d757374206265206772656174657260408201527f207468616e207468652063757272656e74206e6f6e6365000000000000000000606082015260800190565b6020808252601d908201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e7472616374000000604082015260600190565b6020808252602a908201527f5361666545524332303a204552433230206f7065726174696f6e20646964206e6040820152691bdd081cdd58d8d9595960b21b606082015260800190565b6020808252601f908201527f4d616c666f726d65642063757272656e742076616c696461746f722073657400604082015260600190565b6020808252601f908201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604082015260600190565b600084825260606020830152612065606083018561188a565b828103604084015261207781856118cd565b9695505050505050565b918252602082015260400190565b6040518181016001600160401b03811182821017156120ad57600080fd5b604052919050565b60006001600160401b038211156120ca578081fd5b5060209081020190565b60005b838110156120ef5781810151838201526020016120d7565b83811115610d6a5750506000910152565b6001600160a01b038116811461211557600080fd5b5056fea26469706673582212209c6ac43b22126d1243d25dad52a05a7c4a8c4b17602ce5f1bef1af3a16f9f6af64736f6c634300060c0033",
}

// Hub2ABI is the input ABI used to generate the binding from.
// Deprecated: Use Hub2MetaData.ABI instead.
var Hub2ABI = Hub2MetaData.ABI

// Hub2Bin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use Hub2MetaData.Bin instead.
var Hub2Bin = Hub2MetaData.Bin

// DeployHub2 deploys a new Ethereum contract, binding an instance of Hub2 to it.
func DeployHub2(auth *bind.TransactOpts, backend bind.ContractBackend, _gravityId [32]byte, _powerThreshold *big.Int, _validators []common.Address, _powers []*big.Int) (common.Address, *types.Transaction, *Hub2, error) {
	parsed, err := Hub2MetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(Hub2Bin), backend, _gravityId, _powerThreshold, _validators, _powers)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Hub2{Hub2Caller: Hub2Caller{contract: contract}, Hub2Transactor: Hub2Transactor{contract: contract}, Hub2Filterer: Hub2Filterer{contract: contract}}, nil
}

// Hub2 is an auto generated Go binding around an Ethereum contract.
type Hub2 struct {
	Hub2Caller     // Read-only binding to the contract
	Hub2Transactor // Write-only binding to the contract
	Hub2Filterer   // Log filterer for contract events
}

// Hub2Caller is an auto generated read-only Go binding around an Ethereum contract.
type Hub2Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Hub2Transactor is an auto generated write-only Go binding around an Ethereum contract.
type Hub2Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Hub2Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Hub2Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Hub2Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Hub2Session struct {
	Contract     *Hub2             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Hub2CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Hub2CallerSession struct {
	Contract *Hub2Caller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// Hub2TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Hub2TransactorSession struct {
	Contract     *Hub2Transactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Hub2Raw is an auto generated low-level Go binding around an Ethereum contract.
type Hub2Raw struct {
	Contract *Hub2 // Generic contract binding to access the raw methods on
}

// Hub2CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Hub2CallerRaw struct {
	Contract *Hub2Caller // Generic read-only contract binding to access the raw methods on
}

// Hub2TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Hub2TransactorRaw struct {
	Contract *Hub2Transactor // Generic write-only contract binding to access the raw methods on
}

// NewHub2 creates a new instance of Hub2, bound to a specific deployed contract.
func NewHub2(address common.Address, backend bind.ContractBackend) (*Hub2, error) {
	contract, err := bindHub2(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Hub2{Hub2Caller: Hub2Caller{contract: contract}, Hub2Transactor: Hub2Transactor{contract: contract}, Hub2Filterer: Hub2Filterer{contract: contract}}, nil
}

// NewHub2Caller creates a new read-only instance of Hub2, bound to a specific deployed contract.
func NewHub2Caller(address common.Address, caller bind.ContractCaller) (*Hub2Caller, error) {
	contract, err := bindHub2(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Hub2Caller{contract: contract}, nil
}

// NewHub2Transactor creates a new write-only instance of Hub2, bound to a specific deployed contract.
func NewHub2Transactor(address common.Address, transactor bind.ContractTransactor) (*Hub2Transactor, error) {
	contract, err := bindHub2(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Hub2Transactor{contract: contract}, nil
}

// NewHub2Filterer creates a new log filterer instance of Hub2, bound to a specific deployed contract.
func NewHub2Filterer(address common.Address, filterer bind.ContractFilterer) (*Hub2Filterer, error) {
	contract, err := bindHub2(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Hub2Filterer{contract: contract}, nil
}

// bindHub2 binds a generic wrapper to an already deployed contract.
func bindHub2(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(Hub2ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hub2 *Hub2Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hub2.Contract.Hub2Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hub2 *Hub2Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hub2.Contract.Hub2Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hub2 *Hub2Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hub2.Contract.Hub2Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hub2 *Hub2CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hub2.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hub2 *Hub2TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hub2.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hub2 *Hub2TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hub2.Contract.contract.Transact(opts, method, params...)
}

// LastBatchNonce is a free data retrieval call binding the contract method 0x011b2174.
//
// Solidity: function lastBatchNonce(address _erc20Address) view returns(uint256)
func (_Hub2 *Hub2Caller) LastBatchNonce(opts *bind.CallOpts, _erc20Address common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "lastBatchNonce", _erc20Address)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastBatchNonce is a free data retrieval call binding the contract method 0x011b2174.
//
// Solidity: function lastBatchNonce(address _erc20Address) view returns(uint256)
func (_Hub2 *Hub2Session) LastBatchNonce(_erc20Address common.Address) (*big.Int, error) {
	return _Hub2.Contract.LastBatchNonce(&_Hub2.CallOpts, _erc20Address)
}

// LastBatchNonce is a free data retrieval call binding the contract method 0x011b2174.
//
// Solidity: function lastBatchNonce(address _erc20Address) view returns(uint256)
func (_Hub2 *Hub2CallerSession) LastBatchNonce(_erc20Address common.Address) (*big.Int, error) {
	return _Hub2.Contract.LastBatchNonce(&_Hub2.CallOpts, _erc20Address)
}

// LastLogicCallNonce is a free data retrieval call binding the contract method 0xc9d194d5.
//
// Solidity: function lastLogicCallNonce(bytes32 _invalidation_id) view returns(uint256)
func (_Hub2 *Hub2Caller) LastLogicCallNonce(opts *bind.CallOpts, _invalidation_id [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "lastLogicCallNonce", _invalidation_id)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastLogicCallNonce is a free data retrieval call binding the contract method 0xc9d194d5.
//
// Solidity: function lastLogicCallNonce(bytes32 _invalidation_id) view returns(uint256)
func (_Hub2 *Hub2Session) LastLogicCallNonce(_invalidation_id [32]byte) (*big.Int, error) {
	return _Hub2.Contract.LastLogicCallNonce(&_Hub2.CallOpts, _invalidation_id)
}

// LastLogicCallNonce is a free data retrieval call binding the contract method 0xc9d194d5.
//
// Solidity: function lastLogicCallNonce(bytes32 _invalidation_id) view returns(uint256)
func (_Hub2 *Hub2CallerSession) LastLogicCallNonce(_invalidation_id [32]byte) (*big.Int, error) {
	return _Hub2.Contract.LastLogicCallNonce(&_Hub2.CallOpts, _invalidation_id)
}

// StateGravityId is a free data retrieval call binding the contract method 0xbdda81d4.
//
// Solidity: function state_gravityId() view returns(bytes32)
func (_Hub2 *Hub2Caller) StateGravityId(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "state_gravityId")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// StateGravityId is a free data retrieval call binding the contract method 0xbdda81d4.
//
// Solidity: function state_gravityId() view returns(bytes32)
func (_Hub2 *Hub2Session) StateGravityId() ([32]byte, error) {
	return _Hub2.Contract.StateGravityId(&_Hub2.CallOpts)
}

// StateGravityId is a free data retrieval call binding the contract method 0xbdda81d4.
//
// Solidity: function state_gravityId() view returns(bytes32)
func (_Hub2 *Hub2CallerSession) StateGravityId() ([32]byte, error) {
	return _Hub2.Contract.StateGravityId(&_Hub2.CallOpts)
}

// StateInvalidationMapping is a free data retrieval call binding the contract method 0x7dfb6f86.
//
// Solidity: function state_invalidationMapping(bytes32 ) view returns(uint256)
func (_Hub2 *Hub2Caller) StateInvalidationMapping(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "state_invalidationMapping", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StateInvalidationMapping is a free data retrieval call binding the contract method 0x7dfb6f86.
//
// Solidity: function state_invalidationMapping(bytes32 ) view returns(uint256)
func (_Hub2 *Hub2Session) StateInvalidationMapping(arg0 [32]byte) (*big.Int, error) {
	return _Hub2.Contract.StateInvalidationMapping(&_Hub2.CallOpts, arg0)
}

// StateInvalidationMapping is a free data retrieval call binding the contract method 0x7dfb6f86.
//
// Solidity: function state_invalidationMapping(bytes32 ) view returns(uint256)
func (_Hub2 *Hub2CallerSession) StateInvalidationMapping(arg0 [32]byte) (*big.Int, error) {
	return _Hub2.Contract.StateInvalidationMapping(&_Hub2.CallOpts, arg0)
}

// StateLastBatchNonces is a free data retrieval call binding the contract method 0xdf97174b.
//
// Solidity: function state_lastBatchNonces(address ) view returns(uint256)
func (_Hub2 *Hub2Caller) StateLastBatchNonces(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "state_lastBatchNonces", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StateLastBatchNonces is a free data retrieval call binding the contract method 0xdf97174b.
//
// Solidity: function state_lastBatchNonces(address ) view returns(uint256)
func (_Hub2 *Hub2Session) StateLastBatchNonces(arg0 common.Address) (*big.Int, error) {
	return _Hub2.Contract.StateLastBatchNonces(&_Hub2.CallOpts, arg0)
}

// StateLastBatchNonces is a free data retrieval call binding the contract method 0xdf97174b.
//
// Solidity: function state_lastBatchNonces(address ) view returns(uint256)
func (_Hub2 *Hub2CallerSession) StateLastBatchNonces(arg0 common.Address) (*big.Int, error) {
	return _Hub2.Contract.StateLastBatchNonces(&_Hub2.CallOpts, arg0)
}

// StateLastEventNonce is a free data retrieval call binding the contract method 0x73b20547.
//
// Solidity: function state_lastEventNonce() view returns(uint256)
func (_Hub2 *Hub2Caller) StateLastEventNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "state_lastEventNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StateLastEventNonce is a free data retrieval call binding the contract method 0x73b20547.
//
// Solidity: function state_lastEventNonce() view returns(uint256)
func (_Hub2 *Hub2Session) StateLastEventNonce() (*big.Int, error) {
	return _Hub2.Contract.StateLastEventNonce(&_Hub2.CallOpts)
}

// StateLastEventNonce is a free data retrieval call binding the contract method 0x73b20547.
//
// Solidity: function state_lastEventNonce() view returns(uint256)
func (_Hub2 *Hub2CallerSession) StateLastEventNonce() (*big.Int, error) {
	return _Hub2.Contract.StateLastEventNonce(&_Hub2.CallOpts)
}

// StateLastValsetCheckpoint is a free data retrieval call binding the contract method 0xf2b53307.
//
// Solidity: function state_lastValsetCheckpoint() view returns(bytes32)
func (_Hub2 *Hub2Caller) StateLastValsetCheckpoint(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "state_lastValsetCheckpoint")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// StateLastValsetCheckpoint is a free data retrieval call binding the contract method 0xf2b53307.
//
// Solidity: function state_lastValsetCheckpoint() view returns(bytes32)
func (_Hub2 *Hub2Session) StateLastValsetCheckpoint() ([32]byte, error) {
	return _Hub2.Contract.StateLastValsetCheckpoint(&_Hub2.CallOpts)
}

// StateLastValsetCheckpoint is a free data retrieval call binding the contract method 0xf2b53307.
//
// Solidity: function state_lastValsetCheckpoint() view returns(bytes32)
func (_Hub2 *Hub2CallerSession) StateLastValsetCheckpoint() ([32]byte, error) {
	return _Hub2.Contract.StateLastValsetCheckpoint(&_Hub2.CallOpts)
}

// StateLastValsetNonce is a free data retrieval call binding the contract method 0xb56561fe.
//
// Solidity: function state_lastValsetNonce() view returns(uint256)
func (_Hub2 *Hub2Caller) StateLastValsetNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "state_lastValsetNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StateLastValsetNonce is a free data retrieval call binding the contract method 0xb56561fe.
//
// Solidity: function state_lastValsetNonce() view returns(uint256)
func (_Hub2 *Hub2Session) StateLastValsetNonce() (*big.Int, error) {
	return _Hub2.Contract.StateLastValsetNonce(&_Hub2.CallOpts)
}

// StateLastValsetNonce is a free data retrieval call binding the contract method 0xb56561fe.
//
// Solidity: function state_lastValsetNonce() view returns(uint256)
func (_Hub2 *Hub2CallerSession) StateLastValsetNonce() (*big.Int, error) {
	return _Hub2.Contract.StateLastValsetNonce(&_Hub2.CallOpts)
}

// StatePowerThreshold is a free data retrieval call binding the contract method 0xe5a2b5d2.
//
// Solidity: function state_powerThreshold() view returns(uint256)
func (_Hub2 *Hub2Caller) StatePowerThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "state_powerThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StatePowerThreshold is a free data retrieval call binding the contract method 0xe5a2b5d2.
//
// Solidity: function state_powerThreshold() view returns(uint256)
func (_Hub2 *Hub2Session) StatePowerThreshold() (*big.Int, error) {
	return _Hub2.Contract.StatePowerThreshold(&_Hub2.CallOpts)
}

// StatePowerThreshold is a free data retrieval call binding the contract method 0xe5a2b5d2.
//
// Solidity: function state_powerThreshold() view returns(uint256)
func (_Hub2 *Hub2CallerSession) StatePowerThreshold() (*big.Int, error) {
	return _Hub2.Contract.StatePowerThreshold(&_Hub2.CallOpts)
}

// TestCheckValidatorSignatures is a free data retrieval call binding the contract method 0xdb7c4e57.
//
// Solidity: function testCheckValidatorSignatures(address[] _currentValidators, uint256[] _currentPowers, uint8[] _v, bytes32[] _r, bytes32[] _s, bytes32 _theHash, uint256 _powerThreshold) pure returns()
func (_Hub2 *Hub2Caller) TestCheckValidatorSignatures(opts *bind.CallOpts, _currentValidators []common.Address, _currentPowers []*big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _theHash [32]byte, _powerThreshold *big.Int) error {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "testCheckValidatorSignatures", _currentValidators, _currentPowers, _v, _r, _s, _theHash, _powerThreshold)

	if err != nil {
		return err
	}

	return err

}

// TestCheckValidatorSignatures is a free data retrieval call binding the contract method 0xdb7c4e57.
//
// Solidity: function testCheckValidatorSignatures(address[] _currentValidators, uint256[] _currentPowers, uint8[] _v, bytes32[] _r, bytes32[] _s, bytes32 _theHash, uint256 _powerThreshold) pure returns()
func (_Hub2 *Hub2Session) TestCheckValidatorSignatures(_currentValidators []common.Address, _currentPowers []*big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _theHash [32]byte, _powerThreshold *big.Int) error {
	return _Hub2.Contract.TestCheckValidatorSignatures(&_Hub2.CallOpts, _currentValidators, _currentPowers, _v, _r, _s, _theHash, _powerThreshold)
}

// TestCheckValidatorSignatures is a free data retrieval call binding the contract method 0xdb7c4e57.
//
// Solidity: function testCheckValidatorSignatures(address[] _currentValidators, uint256[] _currentPowers, uint8[] _v, bytes32[] _r, bytes32[] _s, bytes32 _theHash, uint256 _powerThreshold) pure returns()
func (_Hub2 *Hub2CallerSession) TestCheckValidatorSignatures(_currentValidators []common.Address, _currentPowers []*big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _theHash [32]byte, _powerThreshold *big.Int) error {
	return _Hub2.Contract.TestCheckValidatorSignatures(&_Hub2.CallOpts, _currentValidators, _currentPowers, _v, _r, _s, _theHash, _powerThreshold)
}

// TestMakeCheckpoint is a free data retrieval call binding the contract method 0xc227c30b.
//
// Solidity: function testMakeCheckpoint(address[] _validators, uint256[] _powers, uint256 _valsetNonce, bytes32 _gravityId) pure returns()
func (_Hub2 *Hub2Caller) TestMakeCheckpoint(opts *bind.CallOpts, _validators []common.Address, _powers []*big.Int, _valsetNonce *big.Int, _gravityId [32]byte) error {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "testMakeCheckpoint", _validators, _powers, _valsetNonce, _gravityId)

	if err != nil {
		return err
	}

	return err

}

// TestMakeCheckpoint is a free data retrieval call binding the contract method 0xc227c30b.
//
// Solidity: function testMakeCheckpoint(address[] _validators, uint256[] _powers, uint256 _valsetNonce, bytes32 _gravityId) pure returns()
func (_Hub2 *Hub2Session) TestMakeCheckpoint(_validators []common.Address, _powers []*big.Int, _valsetNonce *big.Int, _gravityId [32]byte) error {
	return _Hub2.Contract.TestMakeCheckpoint(&_Hub2.CallOpts, _validators, _powers, _valsetNonce, _gravityId)
}

// TestMakeCheckpoint is a free data retrieval call binding the contract method 0xc227c30b.
//
// Solidity: function testMakeCheckpoint(address[] _validators, uint256[] _powers, uint256 _valsetNonce, bytes32 _gravityId) pure returns()
func (_Hub2 *Hub2CallerSession) TestMakeCheckpoint(_validators []common.Address, _powers []*big.Int, _valsetNonce *big.Int, _gravityId [32]byte) error {
	return _Hub2.Contract.TestMakeCheckpoint(&_Hub2.CallOpts, _validators, _powers, _valsetNonce, _gravityId)
}

// WethAddress is a free data retrieval call binding the contract method 0x4f0e0ef3.
//
// Solidity: function wethAddress() view returns(address)
func (_Hub2 *Hub2Caller) WethAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hub2.contract.Call(opts, &out, "wethAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WethAddress is a free data retrieval call binding the contract method 0x4f0e0ef3.
//
// Solidity: function wethAddress() view returns(address)
func (_Hub2 *Hub2Session) WethAddress() (common.Address, error) {
	return _Hub2.Contract.WethAddress(&_Hub2.CallOpts)
}

// WethAddress is a free data retrieval call binding the contract method 0x4f0e0ef3.
//
// Solidity: function wethAddress() view returns(address)
func (_Hub2 *Hub2CallerSession) WethAddress() (common.Address, error) {
	return _Hub2.Contract.WethAddress(&_Hub2.CallOpts)
}

// SendToHub is a paid mutator transaction binding the contract method 0xe08bf6ea.
//
// Solidity: function sendToHub(address _tokenContract, bytes32 _destination, uint256 _amount) returns()
func (_Hub2 *Hub2Transactor) SendToHub(opts *bind.TransactOpts, _tokenContract common.Address, _destination [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Hub2.contract.Transact(opts, "sendToHub", _tokenContract, _destination, _amount)
}

// SendToHub is a paid mutator transaction binding the contract method 0xe08bf6ea.
//
// Solidity: function sendToHub(address _tokenContract, bytes32 _destination, uint256 _amount) returns()
func (_Hub2 *Hub2Session) SendToHub(_tokenContract common.Address, _destination [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Hub2.Contract.SendToHub(&_Hub2.TransactOpts, _tokenContract, _destination, _amount)
}

// SendToHub is a paid mutator transaction binding the contract method 0xe08bf6ea.
//
// Solidity: function sendToHub(address _tokenContract, bytes32 _destination, uint256 _amount) returns()
func (_Hub2 *Hub2TransactorSession) SendToHub(_tokenContract common.Address, _destination [32]byte, _amount *big.Int) (*types.Transaction, error) {
	return _Hub2.Contract.SendToHub(&_Hub2.TransactOpts, _tokenContract, _destination, _amount)
}

// SubmitBatch is a paid mutator transaction binding the contract method 0x83b435db.
//
// Solidity: function submitBatch(address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s, uint256[] _amounts, address[] _destinations, uint256[] _fees, uint256 _batchNonce, address _tokenContract, uint256 _batchTimeout) returns()
func (_Hub2 *Hub2Transactor) SubmitBatch(opts *bind.TransactOpts, _currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _amounts []*big.Int, _destinations []common.Address, _fees []*big.Int, _batchNonce *big.Int, _tokenContract common.Address, _batchTimeout *big.Int) (*types.Transaction, error) {
	return _Hub2.contract.Transact(opts, "submitBatch", _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s, _amounts, _destinations, _fees, _batchNonce, _tokenContract, _batchTimeout)
}

// SubmitBatch is a paid mutator transaction binding the contract method 0x83b435db.
//
// Solidity: function submitBatch(address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s, uint256[] _amounts, address[] _destinations, uint256[] _fees, uint256 _batchNonce, address _tokenContract, uint256 _batchTimeout) returns()
func (_Hub2 *Hub2Session) SubmitBatch(_currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _amounts []*big.Int, _destinations []common.Address, _fees []*big.Int, _batchNonce *big.Int, _tokenContract common.Address, _batchTimeout *big.Int) (*types.Transaction, error) {
	return _Hub2.Contract.SubmitBatch(&_Hub2.TransactOpts, _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s, _amounts, _destinations, _fees, _batchNonce, _tokenContract, _batchTimeout)
}

// SubmitBatch is a paid mutator transaction binding the contract method 0x83b435db.
//
// Solidity: function submitBatch(address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s, uint256[] _amounts, address[] _destinations, uint256[] _fees, uint256 _batchNonce, address _tokenContract, uint256 _batchTimeout) returns()
func (_Hub2 *Hub2TransactorSession) SubmitBatch(_currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _amounts []*big.Int, _destinations []common.Address, _fees []*big.Int, _batchNonce *big.Int, _tokenContract common.Address, _batchTimeout *big.Int) (*types.Transaction, error) {
	return _Hub2.Contract.SubmitBatch(&_Hub2.TransactOpts, _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s, _amounts, _destinations, _fees, _batchNonce, _tokenContract, _batchTimeout)
}

// SubmitLogicCall is a paid mutator transaction binding the contract method 0x0c246c82.
//
// Solidity: function submitLogicCall(address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s, (uint256[],address[],uint256[],address[],address,bytes,uint256,bytes32,uint256) _args) returns()
func (_Hub2 *Hub2Transactor) SubmitLogicCall(opts *bind.TransactOpts, _currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _args LogicCallArgs) (*types.Transaction, error) {
	return _Hub2.contract.Transact(opts, "submitLogicCall", _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s, _args)
}

// SubmitLogicCall is a paid mutator transaction binding the contract method 0x0c246c82.
//
// Solidity: function submitLogicCall(address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s, (uint256[],address[],uint256[],address[],address,bytes,uint256,bytes32,uint256) _args) returns()
func (_Hub2 *Hub2Session) SubmitLogicCall(_currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _args LogicCallArgs) (*types.Transaction, error) {
	return _Hub2.Contract.SubmitLogicCall(&_Hub2.TransactOpts, _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s, _args)
}

// SubmitLogicCall is a paid mutator transaction binding the contract method 0x0c246c82.
//
// Solidity: function submitLogicCall(address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s, (uint256[],address[],uint256[],address[],address,bytes,uint256,bytes32,uint256) _args) returns()
func (_Hub2 *Hub2TransactorSession) SubmitLogicCall(_currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte, _args LogicCallArgs) (*types.Transaction, error) {
	return _Hub2.Contract.SubmitLogicCall(&_Hub2.TransactOpts, _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s, _args)
}

// UpdateValset is a paid mutator transaction binding the contract method 0xe3cb9f62.
//
// Solidity: function updateValset(address[] _newValidators, uint256[] _newPowers, uint256 _newValsetNonce, address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
func (_Hub2 *Hub2Transactor) UpdateValset(opts *bind.TransactOpts, _newValidators []common.Address, _newPowers []*big.Int, _newValsetNonce *big.Int, _currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte) (*types.Transaction, error) {
	return _Hub2.contract.Transact(opts, "updateValset", _newValidators, _newPowers, _newValsetNonce, _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s)
}

// UpdateValset is a paid mutator transaction binding the contract method 0xe3cb9f62.
//
// Solidity: function updateValset(address[] _newValidators, uint256[] _newPowers, uint256 _newValsetNonce, address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
func (_Hub2 *Hub2Session) UpdateValset(_newValidators []common.Address, _newPowers []*big.Int, _newValsetNonce *big.Int, _currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte) (*types.Transaction, error) {
	return _Hub2.Contract.UpdateValset(&_Hub2.TransactOpts, _newValidators, _newPowers, _newValsetNonce, _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s)
}

// UpdateValset is a paid mutator transaction binding the contract method 0xe3cb9f62.
//
// Solidity: function updateValset(address[] _newValidators, uint256[] _newPowers, uint256 _newValsetNonce, address[] _currentValidators, uint256[] _currentPowers, uint256 _currentValsetNonce, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
func (_Hub2 *Hub2TransactorSession) UpdateValset(_newValidators []common.Address, _newPowers []*big.Int, _newValsetNonce *big.Int, _currentValidators []common.Address, _currentPowers []*big.Int, _currentValsetNonce *big.Int, _v []uint8, _r [][32]byte, _s [][32]byte) (*types.Transaction, error) {
	return _Hub2.Contract.UpdateValset(&_Hub2.TransactOpts, _newValidators, _newPowers, _newValsetNonce, _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s)
}

// Hub2LogicCallEventIterator is returned from FilterLogicCallEvent and is used to iterate over the raw logs and unpacked data for LogicCallEvent events raised by the Hub2 contract.
type Hub2LogicCallEventIterator struct {
	Event *Hub2LogicCallEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hub2LogicCallEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hub2LogicCallEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hub2LogicCallEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hub2LogicCallEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hub2LogicCallEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hub2LogicCallEvent represents a LogicCallEvent event raised by the Hub2 contract.
type Hub2LogicCallEvent struct {
	InvalidationId    [32]byte
	InvalidationNonce *big.Int
	ReturnData        []byte
	EventNonce        *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterLogicCallEvent is a free log retrieval operation binding the contract event 0x7c2bb24f8e1b3725cb613d7f11ef97d9745cc97a0e40f730621c052d684077a1.
//
// Solidity: event LogicCallEvent(bytes32 _invalidationId, uint256 _invalidationNonce, bytes _returnData, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) FilterLogicCallEvent(opts *bind.FilterOpts) (*Hub2LogicCallEventIterator, error) {

	logs, sub, err := _Hub2.contract.FilterLogs(opts, "LogicCallEvent")
	if err != nil {
		return nil, err
	}
	return &Hub2LogicCallEventIterator{contract: _Hub2.contract, event: "LogicCallEvent", logs: logs, sub: sub}, nil
}

// WatchLogicCallEvent is a free log subscription operation binding the contract event 0x7c2bb24f8e1b3725cb613d7f11ef97d9745cc97a0e40f730621c052d684077a1.
//
// Solidity: event LogicCallEvent(bytes32 _invalidationId, uint256 _invalidationNonce, bytes _returnData, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) WatchLogicCallEvent(opts *bind.WatchOpts, sink chan<- *Hub2LogicCallEvent) (event.Subscription, error) {

	logs, sub, err := _Hub2.contract.WatchLogs(opts, "LogicCallEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hub2LogicCallEvent)
				if err := _Hub2.contract.UnpackLog(event, "LogicCallEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLogicCallEvent is a log parse operation binding the contract event 0x7c2bb24f8e1b3725cb613d7f11ef97d9745cc97a0e40f730621c052d684077a1.
//
// Solidity: event LogicCallEvent(bytes32 _invalidationId, uint256 _invalidationNonce, bytes _returnData, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) ParseLogicCallEvent(log types.Log) (*Hub2LogicCallEvent, error) {
	event := new(Hub2LogicCallEvent)
	if err := _Hub2.contract.UnpackLog(event, "LogicCallEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hub2SendToHubEventIterator is returned from FilterSendToHubEvent and is used to iterate over the raw logs and unpacked data for SendToHubEvent events raised by the Hub2 contract.
type Hub2SendToHubEventIterator struct {
	Event *Hub2SendToHubEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hub2SendToHubEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hub2SendToHubEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hub2SendToHubEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hub2SendToHubEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hub2SendToHubEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hub2SendToHubEvent represents a SendToHubEvent event raised by the Hub2 contract.
type Hub2SendToHubEvent struct {
	TokenContract common.Address
	Sender        common.Address
	Destination   [32]byte
	Amount        *big.Int
	EventNonce    *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSendToHubEvent is a free log retrieval operation binding the contract event 0x8a90c92ae4d3ad99d0e0af0871264ba019979b4532daba15508810d224e8bbf6.
//
// Solidity: event SendToHubEvent(address indexed _tokenContract, address indexed _sender, bytes32 indexed _destination, uint256 _amount, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) FilterSendToHubEvent(opts *bind.FilterOpts, _tokenContract []common.Address, _sender []common.Address, _destination [][32]byte) (*Hub2SendToHubEventIterator, error) {

	var _tokenContractRule []interface{}
	for _, _tokenContractItem := range _tokenContract {
		_tokenContractRule = append(_tokenContractRule, _tokenContractItem)
	}
	var _senderRule []interface{}
	for _, _senderItem := range _sender {
		_senderRule = append(_senderRule, _senderItem)
	}
	var _destinationRule []interface{}
	for _, _destinationItem := range _destination {
		_destinationRule = append(_destinationRule, _destinationItem)
	}

	logs, sub, err := _Hub2.contract.FilterLogs(opts, "SendToHubEvent", _tokenContractRule, _senderRule, _destinationRule)
	if err != nil {
		return nil, err
	}
	return &Hub2SendToHubEventIterator{contract: _Hub2.contract, event: "SendToHubEvent", logs: logs, sub: sub}, nil
}

// WatchSendToHubEvent is a free log subscription operation binding the contract event 0x8a90c92ae4d3ad99d0e0af0871264ba019979b4532daba15508810d224e8bbf6.
//
// Solidity: event SendToHubEvent(address indexed _tokenContract, address indexed _sender, bytes32 indexed _destination, uint256 _amount, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) WatchSendToHubEvent(opts *bind.WatchOpts, sink chan<- *Hub2SendToHubEvent, _tokenContract []common.Address, _sender []common.Address, _destination [][32]byte) (event.Subscription, error) {

	var _tokenContractRule []interface{}
	for _, _tokenContractItem := range _tokenContract {
		_tokenContractRule = append(_tokenContractRule, _tokenContractItem)
	}
	var _senderRule []interface{}
	for _, _senderItem := range _sender {
		_senderRule = append(_senderRule, _senderItem)
	}
	var _destinationRule []interface{}
	for _, _destinationItem := range _destination {
		_destinationRule = append(_destinationRule, _destinationItem)
	}

	logs, sub, err := _Hub2.contract.WatchLogs(opts, "SendToHubEvent", _tokenContractRule, _senderRule, _destinationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hub2SendToHubEvent)
				if err := _Hub2.contract.UnpackLog(event, "SendToHubEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSendToHubEvent is a log parse operation binding the contract event 0x8a90c92ae4d3ad99d0e0af0871264ba019979b4532daba15508810d224e8bbf6.
//
// Solidity: event SendToHubEvent(address indexed _tokenContract, address indexed _sender, bytes32 indexed _destination, uint256 _amount, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) ParseSendToHubEvent(log types.Log) (*Hub2SendToHubEvent, error) {
	event := new(Hub2SendToHubEvent)
	if err := _Hub2.contract.UnpackLog(event, "SendToHubEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hub2TransactionBatchExecutedEventIterator is returned from FilterTransactionBatchExecutedEvent and is used to iterate over the raw logs and unpacked data for TransactionBatchExecutedEvent events raised by the Hub2 contract.
type Hub2TransactionBatchExecutedEventIterator struct {
	Event *Hub2TransactionBatchExecutedEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hub2TransactionBatchExecutedEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hub2TransactionBatchExecutedEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hub2TransactionBatchExecutedEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hub2TransactionBatchExecutedEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hub2TransactionBatchExecutedEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hub2TransactionBatchExecutedEvent represents a TransactionBatchExecutedEvent event raised by the Hub2 contract.
type Hub2TransactionBatchExecutedEvent struct {
	BatchNonce *big.Int
	Token      common.Address
	EventNonce *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterTransactionBatchExecutedEvent is a free log retrieval operation binding the contract event 0x02c7e81975f8edb86e2a0c038b7b86a49c744236abf0f6177ff5afc6986ab708.
//
// Solidity: event TransactionBatchExecutedEvent(uint256 indexed _batchNonce, address indexed _token, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) FilterTransactionBatchExecutedEvent(opts *bind.FilterOpts, _batchNonce []*big.Int, _token []common.Address) (*Hub2TransactionBatchExecutedEventIterator, error) {

	var _batchNonceRule []interface{}
	for _, _batchNonceItem := range _batchNonce {
		_batchNonceRule = append(_batchNonceRule, _batchNonceItem)
	}
	var _tokenRule []interface{}
	for _, _tokenItem := range _token {
		_tokenRule = append(_tokenRule, _tokenItem)
	}

	logs, sub, err := _Hub2.contract.FilterLogs(opts, "TransactionBatchExecutedEvent", _batchNonceRule, _tokenRule)
	if err != nil {
		return nil, err
	}
	return &Hub2TransactionBatchExecutedEventIterator{contract: _Hub2.contract, event: "TransactionBatchExecutedEvent", logs: logs, sub: sub}, nil
}

// WatchTransactionBatchExecutedEvent is a free log subscription operation binding the contract event 0x02c7e81975f8edb86e2a0c038b7b86a49c744236abf0f6177ff5afc6986ab708.
//
// Solidity: event TransactionBatchExecutedEvent(uint256 indexed _batchNonce, address indexed _token, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) WatchTransactionBatchExecutedEvent(opts *bind.WatchOpts, sink chan<- *Hub2TransactionBatchExecutedEvent, _batchNonce []*big.Int, _token []common.Address) (event.Subscription, error) {

	var _batchNonceRule []interface{}
	for _, _batchNonceItem := range _batchNonce {
		_batchNonceRule = append(_batchNonceRule, _batchNonceItem)
	}
	var _tokenRule []interface{}
	for _, _tokenItem := range _token {
		_tokenRule = append(_tokenRule, _tokenItem)
	}

	logs, sub, err := _Hub2.contract.WatchLogs(opts, "TransactionBatchExecutedEvent", _batchNonceRule, _tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hub2TransactionBatchExecutedEvent)
				if err := _Hub2.contract.UnpackLog(event, "TransactionBatchExecutedEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransactionBatchExecutedEvent is a log parse operation binding the contract event 0x02c7e81975f8edb86e2a0c038b7b86a49c744236abf0f6177ff5afc6986ab708.
//
// Solidity: event TransactionBatchExecutedEvent(uint256 indexed _batchNonce, address indexed _token, uint256 _eventNonce)
func (_Hub2 *Hub2Filterer) ParseTransactionBatchExecutedEvent(log types.Log) (*Hub2TransactionBatchExecutedEvent, error) {
	event := new(Hub2TransactionBatchExecutedEvent)
	if err := _Hub2.contract.UnpackLog(event, "TransactionBatchExecutedEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Hub2ValsetUpdatedEventIterator is returned from FilterValsetUpdatedEvent and is used to iterate over the raw logs and unpacked data for ValsetUpdatedEvent events raised by the Hub2 contract.
type Hub2ValsetUpdatedEventIterator struct {
	Event *Hub2ValsetUpdatedEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *Hub2ValsetUpdatedEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Hub2ValsetUpdatedEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(Hub2ValsetUpdatedEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *Hub2ValsetUpdatedEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Hub2ValsetUpdatedEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Hub2ValsetUpdatedEvent represents a ValsetUpdatedEvent event raised by the Hub2 contract.
type Hub2ValsetUpdatedEvent struct {
	NewValsetNonce *big.Int
	EventNonce     *big.Int
	Validators     []common.Address
	Powers         []*big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterValsetUpdatedEvent is a free log retrieval operation binding the contract event 0xb119f1f36224601586b5037da909ecf37e83864dddea5d32ad4e32ac1d97e62b.
//
// Solidity: event ValsetUpdatedEvent(uint256 indexed _newValsetNonce, uint256 _eventNonce, address[] _validators, uint256[] _powers)
func (_Hub2 *Hub2Filterer) FilterValsetUpdatedEvent(opts *bind.FilterOpts, _newValsetNonce []*big.Int) (*Hub2ValsetUpdatedEventIterator, error) {

	var _newValsetNonceRule []interface{}
	for _, _newValsetNonceItem := range _newValsetNonce {
		_newValsetNonceRule = append(_newValsetNonceRule, _newValsetNonceItem)
	}

	logs, sub, err := _Hub2.contract.FilterLogs(opts, "ValsetUpdatedEvent", _newValsetNonceRule)
	if err != nil {
		return nil, err
	}
	return &Hub2ValsetUpdatedEventIterator{contract: _Hub2.contract, event: "ValsetUpdatedEvent", logs: logs, sub: sub}, nil
}

// WatchValsetUpdatedEvent is a free log subscription operation binding the contract event 0xb119f1f36224601586b5037da909ecf37e83864dddea5d32ad4e32ac1d97e62b.
//
// Solidity: event ValsetUpdatedEvent(uint256 indexed _newValsetNonce, uint256 _eventNonce, address[] _validators, uint256[] _powers)
func (_Hub2 *Hub2Filterer) WatchValsetUpdatedEvent(opts *bind.WatchOpts, sink chan<- *Hub2ValsetUpdatedEvent, _newValsetNonce []*big.Int) (event.Subscription, error) {

	var _newValsetNonceRule []interface{}
	for _, _newValsetNonceItem := range _newValsetNonce {
		_newValsetNonceRule = append(_newValsetNonceRule, _newValsetNonceItem)
	}

	logs, sub, err := _Hub2.contract.WatchLogs(opts, "ValsetUpdatedEvent", _newValsetNonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Hub2ValsetUpdatedEvent)
				if err := _Hub2.contract.UnpackLog(event, "ValsetUpdatedEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseValsetUpdatedEvent is a log parse operation binding the contract event 0xb119f1f36224601586b5037da909ecf37e83864dddea5d32ad4e32ac1d97e62b.
//
// Solidity: event ValsetUpdatedEvent(uint256 indexed _newValsetNonce, uint256 _eventNonce, address[] _validators, uint256[] _powers)
func (_Hub2 *Hub2Filterer) ParseValsetUpdatedEvent(log types.Log) (*Hub2ValsetUpdatedEvent, error) {
	event := new(Hub2ValsetUpdatedEvent)
	if err := _Hub2.contract.UnpackLog(event, "ValsetUpdatedEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
