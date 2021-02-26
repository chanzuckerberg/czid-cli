package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// printSecretCmd represents the printSecret command
var printSecretCmd = &cobra.Command{
	Use:   "print-secret",
	Short: "Print authentication secret",
	Long: `You can set the authentication secret via the
IDSEQ_CLI_SECRET environment variable or by adding it
manually to your config file. You must login with 'idseq login'
to obtain a secret. You can then use idseq print-secret to
access this secret for use in automated systems where you
can't log in manually.

WARNING: this is a long lived access token, be extremely
careful while handling it.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("printSecret called")
	},
}

func init() {
	RootCmd.AddCommand(printSecretCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// printSecretCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// printSecretCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
