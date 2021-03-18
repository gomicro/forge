package config

import (
	"github.com/spf13/cobra"

	"github.com/gomicro/forge/confile"
	"github.com/gomicro/forge/fmt"
)

func init() {
	ConfigCmd.AddCommand(fmtCmd)
}

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format the forge file",
	Long:  `Format and adjust the forge file for consistency.`,
	Run:   fmtFunc,
}

func fmtFunc(cmd *cobra.Command, args []string) {
	conf, err := confile.ParseFromFile()
	if err != nil {
		fmt.Printf("Failed: %v", err.Error())
	}

	err = conf.Fmt()
	if err != nil {
		fmt.Printf("Failed: %v", err.Error())
	}
}
