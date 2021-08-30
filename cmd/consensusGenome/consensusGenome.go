package consensusGenome

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/util"
)

var projectName string
var stringMetadata map[string]string
var metadataCSVPath string
var technology string
var wetlabProtocol string
var medakaModel string
var clearLabs bool

var Technologies = map[string]string{
	"Illumina": "Illumina",
	"Nanopore": "ONT",
}
var technologyOptionsString string

var WetlabProtocols = map[string]string{
	"ARTIC v4":                            "artic_v4",
	"ARTIC v3 - Short Amplicons (275 bp)": "artic_short_amplicons",
	"ARTIC v3":                            "artic",
	"AmpliSeq":                            "ampliseq",
	"Combined MSSPE & ARTIC v3":           "combined_msspe_artic",
	"MSSPE":                               "msspe",
	"SNAP":                                "snap",
	"COVIDseq":                            "covidseq",
	"Midnight":                            "midnight",
}
var wetlabProtocolOptionsString string

var nanoporeWetLabProtocols = map[string]string{
	"ARTIC v4": "artic_v4",
	"Midnight": "midnight",
	"ARTIC v3": "artic",
}
var nanoporeWetlabProtocolOptionsString string
var nanoporeDefaultWetlabProtocol = "ARTIC v3"

var MedakaModels = map[string]string{
	"r941_min_fast_g303":     "r941_min_fast_g303",
	"r941_min_high_g303":     "r941_min_high_g303",
	"r941_min_high_g330":     "r941_min_high_g330",
	"r941_min_high_g340_rle": "r941_min_high_g340_rle",
	"r941_min_high_g344":     "r941_min_high_g344",
	"r941_min_high_g351":     "r941_min_high_g351",
	// split
	"r103_prom_high_g360":     "r103_prom_high_g360",
	"r103_prom_snp_g3210":     "r103_prom_snp_g3210",
	"r103_prom_variant_g3210": "r103_prom_variant_g3210",
	"r941_prom_fast_g303":     "r941_prom_fast_g303",
	"r941_prom_high_g303":     "r941_prom_high_g303",
	"r941_prom_high_g330":     "r941_prom_high_g330",
	"r941_prom_high_g344":     "r941_prom_high_g344",
	"r941_prom_high_g360":     "r941_prom_high_g360",
	"r941_prom_high_g4011":    "r941_prom_high_g4011",
	"r941_prom_snp_g303":      "r941_prom_snp_g303",
	"r941_prom_snp_g322":      "r941_prom_snp_g322",
	"r941_prom_snp_g360":      "r941_prom_snp_g360",
	"r941_prom_variant_g303":  "r941_prom_variant_g303",
	"r941_prom_variant_g322":  "r941_prom_variant_g322",
	"r941_prom_variant_g360":  "r941_prom_variant_g360",
}
var medakaModelsString string
var defaultMedakaModel = "r941_min_high_g360"

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
	technologyOptionsString = fmt.Sprintf(
		"\"%s\"",
		strings.Join(util.StringMapKeys(Technologies), "\", \""),
	)

	wetlabProtocolOptionsString = fmt.Sprintf(
		"\"%s\"",
		strings.Join(util.StringMapKeys(WetlabProtocols), "\", \""),
	)

	medakaModelsString = fmt.Sprintf(
		"\"%s\"",
		strings.Join(util.StringMapKeys(MedakaModels), "\", \""),
	)

	nanoporeWetlabProtocolOptionsString = fmt.Sprintf(
		"\"%s\"",
		strings.Join(util.StringMapKeys(nanoporeWetLabProtocols), "\", \""),
	)

	c.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website (required)")
	c.Flags().StringToStringVarP(&stringMetadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	c.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
	c.Flags().StringVar(&technology, "sequencing-platform", "", fmt.Sprintf("Sequencing platform used to sequence the sample, options: %s", technologyOptionsString))
	c.Flags().StringVar(&wetlabProtocol, "wetlab-protocol", "", fmt.Sprintf(
		"Wetlab protocol followed.\n  Options for Nanopore (optional, default: \"%s\"): %s\n  Options for Illumina (reqired): %s",
		nanoporeDefaultWetlabProtocol,
		nanoporeWetlabProtocolOptionsString,
		wetlabProtocolOptionsString,
	))
	c.Flags().StringVar(&medakaModel, "medaka-model", "", fmt.Sprintf(
		"Medaka model (only supported for Nanopore, optional default: %s)\n  Medaka is a tool to create consensus sequences and variant calls from Nanopore sequencing data.\n  Options: %s",
		defaultMedakaModel,
		medakaModelsString),
	)
	c.Flags().BoolVar(&clearLabs, "clearlabs", false, fmt.Sprintf(
		"Pipeline will be adjusted to accomodate Clear Lab fastq files which have undergone the length filtering and trimming steps.\n  Only supported for Nanopore. Requires default wetlab-protocol (\"%s\") and default medaka-model (\"%s\")",
		nanoporeDefaultWetlabProtocol,
		defaultMedakaModel,
	))
}

func validateCommonArgs() error {
	if projectName == "" {
		return errors.New("missing required argument: project")
	}
	if technology == "" {
		return errors.New("missing required argument: sequencing-platform")
	}
	if _, has := Technologies[technology]; !has {
		return fmt.Errorf("sequencing platform \"%s\" not supported, please choose one of: %s", technology, technologyOptionsString)
	}
	if technology == "Illumina" && wetlabProtocol == "" {
		return errors.New("missing required argument: wetlab-protocol")
	}
	if _, has := WetlabProtocols[wetlabProtocol]; wetlabProtocol != "" && !has {
		return fmt.Errorf("wetlab protocol \"%s\" not supported, please choose one of: %s", wetlabProtocol, wetlabProtocolOptionsString)
	}
	if technology == "Nanopore" && wetlabProtocol == "" {
		wetlabProtocol = "ARTIC v3"
	}
	if _, has := nanoporeWetLabProtocols[wetlabProtocol]; wetlabProtocol != "" && technology == "Nanopore" && !has {
		return fmt.Errorf("wetlab protocol \"%s\" not supported, please choose one of: %s", wetlabProtocol, nanoporeWetlabProtocolOptionsString)
	}
	if technology == "Nanopore" && medakaModel == "" {
		medakaModel = defaultMedakaModel
	}
	if clearLabs && technology == "Illumina" {
		return fmt.Errorf("clearlabs is only supported for Nanore")
	}
	if clearLabs && technology == "Nanopore" {
		if wetlabProtocol != nanoporeDefaultWetlabProtocol {
			return fmt.Errorf("wetlab-protocol %s is required with clearlabs", nanoporeDefaultWetlabProtocol)
		}
		if medakaModel != defaultMedakaModel {
			return fmt.Errorf("medaka-model %s is required with clearlabs", defaultMedakaModel)
		}
	}
	return nil
}
