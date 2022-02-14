package cmd

import (
	"fmt"

	"github.com/chanzuckerberg/czid-cli/pkg/auth0"
	"github.com/spf13/cobra"
)

// printSecretCmd represents the printSecret command
var printSecretCmd = &cobra.Command{
	Use:   "print-secret",
	Short: "Print authentication secret",
	Long: `You can set the authentication secret via the
CZID_CLI_SECRET environment variable or by adding it
manually to your config file. You must login with 'czid login'
to obtain a secret. You can then use czid print-secret to
access this secret for use in automated systems where you
can't log in manually.

WARNING: this is a long lived access token, be extremely
careful while handling it.`,
	Run: func(cmd *cobra.Command, args []string) {
		secret, hasSecret := auth0.DefaultClient.Secret()
		if !hasSecret {
			fmt.Println("no secret defined, try running 'czid login' to generate one")
		} else {
			fmt.Println(secret)
		}
	},
}

func init() {
	RootCmd.AddCommand(printSecretCmd)
}
