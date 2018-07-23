package orderer

import (
	"github.com/spf13/cobra"
)

var (
	ordererDir string
)

func init() {
	Cmd.PersistentFlags().StringVar(&ordererDir, "ordererDir", "", "Path to the orderer directory.")
	if err := Cmd.MarkPersistentFlagRequired("ordererDir"); err != nil {
		panic(err)
	}
}

var Cmd = &cobra.Command{
	Use:   "orderer",
	Short: "This is a orderer command",
	Long:  "A super long version or a random description",
	//Args: cobra.MinimumNArgs(1),
	//Run: func(cmd *cobra.Command, args []string) {
	//	orderer.OrdererInfo(ordererDir)
	//},
}
