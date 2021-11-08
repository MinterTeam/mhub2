package context

import (
	"encoding/json"
	"github.com/MinterTeam/mhub2/minter-connector/config"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client"
	"github.com/MinterTeam/minter-go-sdk/v2/wallet"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"os"
)

type statusData struct {
	LastCheckedMinterBlock uint64 `json:"last_checked_minter_block"`
	LastEventNonce         uint64 `json:"last_event_nonce"`
	LastBatchNonce         uint64 `json:"last_batch_nonce"`
	LastValsetNonce        uint64 `json:"last_valset_nonce"`
}

type Context struct {
	status         statusData
	statusFilePath string

	MinterMultisigAddr string

	CosmosConn   *grpc.ClientConn
	MinterClient *http_client.Client

	OrcAddress   sdk.AccAddress
	OrcPriv      *secp256k1.PrivKey
	MinterWallet *wallet.Wallet
	Logger       log.Logger
}

func (c *Context) LoadStatus(file string, defaultStatus config.MinterConfig) {
	c.statusFilePath = file
	data, err := os.ReadFile(c.statusFilePath)
	if err != nil {
		c.status = statusData{
			LastCheckedMinterBlock: defaultStatus.StartBlock,
			LastEventNonce:         defaultStatus.StartEventNonce,
			LastBatchNonce:         defaultStatus.StartBatchNonce,
			LastValsetNonce:        defaultStatus.StartValsetNonce,
		}
		return
	}

	status := statusData{}
	if err := json.Unmarshal(data, &status); err != nil {
		c.status = statusData{
			LastCheckedMinterBlock: defaultStatus.StartBlock,
			LastEventNonce:         defaultStatus.StartEventNonce,
			LastBatchNonce:         defaultStatus.StartBatchNonce,
			LastValsetNonce:        defaultStatus.StartValsetNonce,
		}
	}

	c.status = status
}

func (c *Context) LastCheckedMinterBlock() uint64 {
	return c.status.LastCheckedMinterBlock
}

func (c *Context) LastEventNonce() uint64 {
	return c.status.LastEventNonce
}

func (c *Context) LastBatchNonce() uint64 {
	return c.status.LastBatchNonce
}

func (c *Context) LastValsetNonce() uint64 {
	return c.status.LastValsetNonce
}

func (c *Context) SetLastCheckedMinterBlock(lastCheckedMinterBlock uint64) {
	c.status.LastCheckedMinterBlock = lastCheckedMinterBlock
	c.Commit()
}

func (c *Context) SetLastEventNonce(lastEventNonce uint64) {
	c.status.LastEventNonce = lastEventNonce
	c.Commit()
}

func (c *Context) SetLastBatchNonce(lastBatchNonce uint64) {
	c.status.LastBatchNonce = lastBatchNonce
	c.Commit()
}

func (c *Context) SetLastValsetNonce(lastValsetNonce uint64) {
	c.status.LastValsetNonce = lastValsetNonce
	c.Commit()
}

func (c *Context) Commit() {
	data, _ := json.Marshal(c.status)
	err := os.WriteFile(c.statusFilePath, data, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
