package cmd

import (
	"github.com/kyroy/fabrictool/pkg/orderer"
	"github.com/spf13/cobra"
)

func init() {
	ordererCommand.AddCommand(ledgerCmd)
}

var ledgerCmd = &cobra.Command{
	Use:   "ledger",
	Short: "This is a ledger command",
	Long:  "A super long version or a random description",
	Run: func(cmd *cobra.Command, args []string) {
		provider := orderer.BlockStoreProdiver(ordererDir, "")
		defer provider.Close()
		orderer.ListLedgers(provider)
	},
}
