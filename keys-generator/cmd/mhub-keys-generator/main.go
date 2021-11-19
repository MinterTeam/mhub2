package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/MinterTeam/mhub2/module/app"
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/status-im/keycard-go/hexutils"
	"os"
)

const ethSignaturePrefix = "\x19Ethereum Signed Message:\n32"

func main() {
	if len(os.Args) == 3 && os.Args[0] == "make_delegate_sign" {
		ethPrivateKey := os.Args[1]
		cosmosAddr := os.Args[2]

		valAddress, err := sdk.AccAddressFromBech32(cosmosAddr)
		if err != nil {
			panic(err)
		}
		signMsgBz := app.MakeEncodingConfig().Marshaler.MustMarshal(&types.DelegateKeysSignMsg{
			ValidatorAddress: sdk.ValAddress(valAddress).String(),
			Nonce:            0,
		})
		hash := append([]uint8(ethSignaturePrefix), crypto.Keccak256Hash(signMsgBz).Bytes()...)
		sig, err := secp256k1.Sign(crypto.Keccak256Hash(hash).Bytes(), hexutils.HexToBytes(ethPrivateKey))
		if err != nil {
			panic(err)
		}

		fmt.Printf("%x", sig)

		return
	}

	privateKey, _ := crypto.GenerateKey()
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Printf("Private key: %s\n", hexutil.Encode(privateKeyBytes)[2:])
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Printf("Ethereum address: %s\n", address)
	fmt.Printf("Minter address: Mx%s\n", address[:2])
}
