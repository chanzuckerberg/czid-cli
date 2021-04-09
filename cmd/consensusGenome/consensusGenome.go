package consensusGenome

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectName string
var stringMetadata map[string]string
var metadataCSVPath string
var technology string
var wetlabProtocol string

var technologies = map[string]string{
	"Illumina": "Illumina",
	"Nanopore": "ONT",
}

var wetlabProtocols = map[string]string{
	"ARTIC v3 - Short Amplicons (275 bp)": "artic_short_amplicons",
	"ARTIC v3":                            "artic",
	"AmpliSeq":                            "ampliseq",
	"Combined MSSPE & ARTIC v3":           "combined_msspe_artic",
	"MSSPE":                               "msspe",
	"SNAP":                                "snap",
}

// ConsensusGenomeCmd represents the ConsensusGenome command
var ConsensusGenomeCmd = &cobra.Command{
	Use:   "consensus-genome",
	Short: "Commands related to the consensus-genome pipeline",
	Long:  "Commands related to the consensus-genome pipeline",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if strings.ToLower(viper.GetString("accepted_user_agreement")) != "y" {
			fmt.Println("Cannot upload samples until the user agreement is accepted, run idseq accept-user-agreement or set IDSEQ_CLI_ACCEPTED_USER_AGREEMENT=Y")
			os.Exit(2)
		}
	},
}

func loadSharedFlags(c *cobra.Command) {
	technologyOptions := make([]string, 0, len(technologies))
	for key := range technologies {
		technologyOptions = append(technologyOptions, key)
	}

	wetlabProtocolOptions := make([]string, 0, len(wetlabProtocol))
	for key := range wetlabProtocols {
		wetlabProtocolOptions = append(wetlabProtocolOptions, key)
	}

	c.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website")
	c.Flags().StringToStringVarP(&stringMetadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	c.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
	c.Flags().StringVar(&technology, "sequencing-platform", "", fmt.Sprintf("Sequencing platform used to sequence the sample, options: '%s'", strings.Join(technologyOptions, "', '")))
	c.Flags().StringVar(&wetlabProtocol, "wetlab-protocol", "", fmt.Sprintf("Wetlab protocol followed (only supported/required for illumina), options: '%s'", strings.Join(wetlabProtocolOptions, "', '")))
}
