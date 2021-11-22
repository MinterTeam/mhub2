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
	"strconv"
)

const ethSignaturePrefix = "\x19Ethereum Signed Message:\n32"

func main() {
	app.SetAddressConfig()

	if len(os.Args) == 5 && os.Args[1] == "make_delegate_sign" {
		ethPrivateKey := os.Args[2]
		cosmosAddr := os.Args[3]
		nonce, err := strconv.Atoi(os.Args[4])

		valAddress, err := sdk.AccAddressFromBech32(cosmosAddr)
		if err != nil {
			panic(err)
		}
		signMsgBz := app.MakeEncodingConfig().Marshaler.MustMarshal(&types.DelegateKeysSignMsg{
			ValidatorAddress: sdk.ValAddress(valAddress).String(),
			Nonce:            uint64(nonce),
		})
		hash := append([]uint8(ethSignaturePrefix), crypto.Keccak256Hash(signMsgBz).Bytes()...)
		sig, err := secp256k1.Sign(crypto.Keccak256Hash(hash).Bytes(), hexutils.HexToBytes(ethPrivateKey))
		if err != nil {
			panic(err)
		}

		fmt.Printf("0x%x\n", sig)

		return
	}

	privateKey, _ := crypto.GenerateKey()
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Printf("Private key: %s\n", hexutil.Encode(privateKeyBytes)[2:])
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Printf("Ethereum address: %s\n", address)
	fmt.Printf("Minter address: Mx%s\n", address[2:])
}
