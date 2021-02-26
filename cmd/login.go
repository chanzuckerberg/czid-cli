package cmd

import (
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/auth0"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with IDSeq",
	Long: `Log into IDSeq so you can upload samples.
This will either open a web page or provide you with
a link to a web page if you use the --headless
option. Once you log in on that web page on any
device (not necessarily the one you ran the command on)
you will be authorized to upload samples to your
IDSeq account.

By default you will remain authenticated for a short
time. If you would like to obtain a secret that
allows you to stay persistently authenticated use the
--persistent option. If you do this a long lived
secret will be added to your configuration file
so please exercise caution when handling this
file. If you suspect your secret has been
comprimised, please reach out to IDSeq support
at https://chanzuckerberg.zendesk.com/hc/en-us/requests/new.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		headless, err := cmd.Flags().GetBool("headless")
		if err != nil {
			return err
		}
		persistent, err := cmd.Flags().GetBool("persistent")
		if err != nil {
			return err
		}
		return auth0.Login(headless, persistent)
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().Bool("headless", false, "don't open the login form in a browser")
	loginCmd.PersistentFlags().Bool("persistent", false, "remain logged in on this device (see description)")
}
