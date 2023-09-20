package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/notional-labs/centauri/v5/app"
	cmd "github.com/notional-labs/centauri/v5/cmd/centaurid/cmd"
	cmdcfg "github.com/notional-labs/centauri/v5/cmd/centaurid/config"
)

func main() {
	cmdcfg.SetupConfig()
	cmdcfg.RegisterDenoms()

	rootCmd, _ := cmd.NewRootCmd()

	rootCmd.AddCommand(AddConsumerSectionCmd(app.DefaultNodeHome))

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
