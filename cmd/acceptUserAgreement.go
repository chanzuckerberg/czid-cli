package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chanzuckerberg/czid-cli/pkg/czid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var autoAccept bool

// acceptUserAgreementCmd represents the acceptUserAgreement command
var acceptUserAgreementCmd = &cobra.Command{
	Use:   "accept-user-agreement",
	Short: "Accept the CZID User Agreement",
	Long:  "Accept the CZID User Agreement",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(`I agree that the data I am uploading to Chan Zuckerberg ID
has been lawfully collected and that I have all the necessary consents,
permissions, and authorizations needed to collect, share, and export data to
Chan Zuckerberg ID as outlined in the Terms (https://czid.org/terms) and Data
Privacy Notice (https://czid.org/privacy).

Accept (y/n)? y/Y for yes (any other input to cancel): `)
		accepted := false
		if !autoAccept {
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			response := input.Text()
			accepted = strings.ToLower(response) == "y"
		}
		if accepted || autoAccept {
			viper.Set("accepted_user_agreement", "Y")

			// If the user's profile_form_version is 0, then they have not yet
			// filled out their profile form, so direct them to the web app to
			// do so.
			// TODO: after accepting user agreement, should stop user from all
			// CLI actions until profile form is completed
			client := czid.DefaultClient
			// TODO: how to get user id?
			fields, err := client.GetUserInfo(1)
			if err != nil {
				cmd.Println("fields:", fields)
				cmd.Println("err:", err) // currently returning Get "/users/1/edit": unsupported protocol scheme ""
			}
			// if fields.ProfileFormVersion == 0 {
			// 	cmd.Println("Please fill out your profile form at https://czid.org/")
			// }

			return viper.WriteConfig()
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(acceptUserAgreementCmd)
	acceptUserAgreementCmd.Flags().BoolP("yes", "y", false, "Accept without prompting")
}
