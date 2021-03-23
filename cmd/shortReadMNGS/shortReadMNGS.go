package shortReadMNGS

import (
	"github.com/spf13/cobra"
)

var projectName string
var stringMetadata map[string]string
var metadataCSVPath string

// shortReadMNGSCmd represents the shortReadMNGSCmd command
var ShortReadMNGSCmd = &cobra.Command{
	Use:   "short-read-mngs",
	Short: "Commands related to the short-read-mngs pipeline",
	Long:  "Commands related to the short-read-mngs pipeline",
}

func loadSharedFlags(c *cobra.Command) {
	c.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website")
	c.Flags().StringToStringVarP(&stringMetadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	c.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
}
