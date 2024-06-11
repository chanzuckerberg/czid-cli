package amr

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectName string
var stringMetadata map[string]string
var metadataCSVPath string
var disableBuffer bool


// AmrCmd represents the Amr command
var AmrCmd = &cobra.Command{
	Use:   "amr",
	Short: "Commands related to the amr pipeline",
	Long:  "Commands related to the amr pipeline",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if strings.ToLower(viper.GetString("accepted_user_agreement")) != "y" {
			fmt.Println("Cannot upload samples until the user agreement is accepted, run czid accept-user-agreement or set CZID_CLI_ACCEPTED_USER_AGREEMENT=Y")
			os.Exit(2)
		}
	},
}

func loadSharedFlags(c *cobra.Command) {
	c.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website (required)")
	c.Flags().StringToStringVarP(&stringMetadata, "metadatum", "m", map[string]string{}, "Metadatum name and value for your sample, ex. 'host=Human'")
	c.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
	c.Flags().BoolVar(&disableBuffer, "disable-buffer", false, "Disable shared buffer pool (useful if running out of memory)")
}

func validateCommonArgs() error {
	if projectName == "" {
		return errors.New("missing required argument: project")
	}

	return nil
}
