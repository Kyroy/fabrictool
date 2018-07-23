package cmd

import (
	"fmt"
	"github.com/kyroy/fabrictool/cmd/orderer"
	"github.com/spf13/cobra"
	"os"
)

var Verbose bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(orderer.Cmd)
}

var rootCmd = &cobra.Command{
	Use:     "fabrictool",
	Short:   "fabrictool is an example how to build a command in Go.",
	Long:    `An example project to build a command-line tool.`,
	Version: fmt.Sprintf("%s, build %s", version, gitCommit),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
