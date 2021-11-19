package app

import (
	mhub2params "github.com/MinterTeam/mhub2/module/app/params"
	"github.com/cosmos/cosmos-sdk/std"
)

// MakeEncodingConfig creates an EncodingConfig for mhub2.
func MakeEncodingConfig() mhub2params.EncodingConfig {
	encodingConfig := mhub2params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
