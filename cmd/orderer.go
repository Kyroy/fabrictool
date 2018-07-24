package cmd

import (
	"github.com/spf13/cobra"
)

var (
	ordererDir string
)

func init() {
	rootCmd.AddCommand(ordererCommand)
	ordererCommand.PersistentFlags().StringVar(&ordererDir, "ordererDir", "", "Path to the orderer directory.")
	if err := ordererCommand.MarkPersistentFlagRequired("ordererDir"); err != nil {
		panic(err)
	}
}

var ordererCommand = &cobra.Command{
	Use:   "orderer",
	Short: "This is a orderer command",
	Long:  "A super long version or a random description",
}
