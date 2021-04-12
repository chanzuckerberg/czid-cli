package consensusGenome

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
var technology string
var wetlabProtocol string

var technologies = map[string]string{
	"Illumina": "Illumina",
	"Nanopore": "ONT",
}
var technologyOptions []string
var technologyOptionsString = ""

var wetlabProtocols = map[string]string{
	"ARTIC v3 - Short Amplicons (275 bp)": "artic_short_amplicons",
	"ARTIC v3":                            "artic",
	"AmpliSeq":                            "ampliseq",
	"Combined MSSPE & ARTIC v3":           "combined_msspe_artic",
	"MSSPE":                               "msspe",
	"SNAP":                                "snap",
}

var wetlabProtocolOptions []string
var wetlabProtocolOptionsString = ""

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
	i := 0
	technologyOptions = make([]string, len(technologies))
	for key := range technologies {
		technologyOptions[i] = key
		i++
	}
	technologyOptionsString = fmt.Sprintf("\"%s\"", strings.Join(technologyOptions, "\", \""))

	i = 0
	wetlabProtocolOptions = make([]string, len(wetlabProtocols))
	for key := range wetlabProtocols {
		wetlabProtocolOptions[i] = key
		i++
	}
	wetlabProtocolOptionsString = fmt.Sprintf("\"%s\"", strings.Join(wetlabProtocolOptions, "\", \""))

	c.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website")
	c.Flags().StringToStringVarP(&stringMetadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	c.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
	c.Flags().StringVar(&technology, "sequencing-platform", "", fmt.Sprintf("Sequencing platform used to sequence the sample, options: %s", technologyOptionsString))
	c.Flags().StringVar(&wetlabProtocol, "wetlab-protocol", "", fmt.Sprintf("Wetlab protocol followed (only supported/required for illumina), options: %s", wetlabProtocolOptionsString))
}

func validateCommonArgs() error {
	if projectName == "" {
		return errors.New("missing required argument: project")
	}
	if technology == "" {
		return errors.New("missing required argument: sequencing-platform")
	}
	if _, has := technologies[technology]; !has {
		return fmt.Errorf("sequencing platform \"%s\" not supported, please choose one of: %s", technology, technologyOptionsString)
	}
	if technology == "Illumina" && wetlabProtocol == "" {
		return errors.New("missing required argument: wetlab-protocol")
	}
	if _, has := wetlabProtocols[wetlabProtocol]; wetlabProtocol != "" && !has {
		return fmt.Errorf("wetlab protocol \"%s\" not supported, please choose one of: %s", wetlabProtocol, wetlabProtocolOptionsString)
	}
	if technology == "Nanopore" && wetlabProtocol != "" {
		return errors.New("wetlab-protocol not supported for Nanopore")
	}
	return nil
}
