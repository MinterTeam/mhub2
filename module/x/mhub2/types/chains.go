package types

type ChainID string

func (c ChainID) Bytes() []byte {
	return []byte(c)
}

func (c ChainID) String() string {
	return string(c)
}

var chains = []ChainID{
	"ethereum",
	"minter",
	"bsc",
}

func GetChains() []ChainID {
	return chains
}
