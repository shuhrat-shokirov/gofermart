package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var greetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Send a greeting",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from GopherMart!")
	},
}

func init() {
	rootCmd.AddCommand(greetCmd)
}
