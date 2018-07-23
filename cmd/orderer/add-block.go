package orderer

import (
	"github.com/kyroy/fabrictool/pkg/orderer"
	"github.com/spf13/cobra"
	"strconv"
)

var ledger string

func init() {
	ledgerCmd.AddCommand(addBlockCmd)
	addBlockCmd.PersistentFlags().StringVar(&ledger, "ledger", "", "Ledger name.")
	if err := addBlockCmd.MarkPersistentFlagRequired("ledger"); err != nil {
		panic(err)
	}
}

var addBlockCmd = &cobra.Command{
	Use:   "add-block",
	Short: "This is a add-block command",
	Long:  "A super long version or a random description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lastOffsetPersisted, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		provider := orderer.BlockStoreProdiver(ordererDir)
		block, err := orderer.LastBlock(provider, ledger)
		if err != nil {
			return err
		}
		kafkaMetadata, err := orderer.LedgerKafkaMetadata(block, ledger)
		if err != nil {
			return err
		}
		kafkaMetadata.LastOffsetPersisted = int64(lastOffsetPersisted)
		block, err = orderer.CreateNoOpBlock(block, kafkaMetadata)
		if err != nil {
			return err
		}
		return orderer.AddBlock(provider, ledger, block)
	},
}
