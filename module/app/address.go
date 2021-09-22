package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetAddressConfig sets the mhub2 app's address configuration.
func SetAddressConfig() {
	config := sdk.GetConfig()

	config.SetAddressVerifier(VerifyAddressFormat)
	config.SetBech32PrefixForAccount("hub", "hubpub")
	config.Seal()
}
