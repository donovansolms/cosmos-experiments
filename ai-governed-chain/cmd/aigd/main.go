package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	aiapp "github.com/donovansolms/cosmos-experiments/ai-governed-chain/app"
)

func main() {
	setAddressPrefixes(aiapp.AccountAddressPrefix)
	rootCmd := NewRootCmd(aiapp.MakeEncodingConfig())
	if err := svrcmd.Execute(rootCmd, "AIG", aiapp.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
