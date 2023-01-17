package metagenomics

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chanzuckerberg/czid-cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var projectName string
var stringMetadata map[string]string
var metadataCSVPath string
var disableBuffer bool
var technology string
var guppyBasecallerSetting string
var workflow string = "short-read-mngs"

var Technologies = map[string]string{
	"Illumina": "Illumina",
	"Nanopore": "ONT",
}
var technologyOptionsString string

var GuppyBasecallerSettings = map[string]string{
	"fast":  "fast",
	"hac":   "hac",
	"super": "super",
}
var guppBasecallerSettingOptionsString string

// MetagenomicsCmd represents the MetagenomicsCmd command
var MetagenomicsCmd = &cobra.Command{
	Use:   "metagenomics",
	Short: "Commands related to metagenomics pipelines",
	Long:  "Commands related to metagenomics pipelines",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if strings.ToLower(viper.GetString("accepted_user_agreement")) != "y" {
			fmt.Println("Cannot upload samples until the user agreement is accepted, run czid accept-user-agreement or set CZID_CLI_ACCEPTED_USER_AGREEMENT=Y")
			os.Exit(2)
		}
	},
}

func loadSharedFlags(c *cobra.Command) {
	technologyOptionsString = fmt.Sprintf(
		"\"%s\"",
		strings.Join(util.StringMapKeys(Technologies), "\", \""),
	)

	guppBasecallerSettingOptionsString = fmt.Sprintf(
		"\"%s\"",
		strings.Join(util.StringMapKeys(GuppyBasecallerSettings), "\", \""),
	)

	c.Flags().StringVarP(&projectName, "project", "p", "", "Project name. Make sure the project is created on the website")
	c.Flags().StringToStringVarP(&stringMetadata, "metadatum", "m", map[string]string{}, "metadatum name and value for your sample, ex. 'host=Human'")
	c.Flags().StringVar(&metadataCSVPath, "metadata-csv", "", "Metadata local file path.")
	c.Flags().StringVar(&technology, "sequencing-platform", "", fmt.Sprintf("Sequencing platform used to sequence the sample, options: %s", technologyOptionsString))
	c.Flags().StringVar(
		&guppyBasecallerSetting,
		"guppy-basecaller-setting",
		"",
		fmt.Sprintf("Specifies which basecalling model of 'Guppy' was used to generate the data. Required for Nanopore, not supported for Illumina. options: %s",
			guppBasecallerSettingOptionsString),
	)
	c.Flags().BoolVar(&disableBuffer, "disable-buffer", false, "Disable shared buffer pool (useful if running out of memory)")
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

	if technology == "Nanopore" && guppyBasecallerSetting == "" {
		return errors.New("missing required argument for sequencing-platform 'Nanopore': guppy-basecaller-setting")
	}

	if technology == "Illumina" && guppyBasecallerSetting != "" {
		return errors.New("guppy-basecaller-setting is not supported for sequencing-platform 'Illumina'")
	}

	if _, has := GuppyBasecallerSettings[guppyBasecallerSetting]; !has && guppyBasecallerSetting != "" {
		return fmt.Errorf(
			"guppy-basecaller-setting \"%s\" not supported, please choose one of: %s",
			guppyBasecallerSetting,
			guppBasecallerSettingOptionsString,
		)
	}

	if technology == "Nanopore" {
		workflow = "long-read-mngs"
	}
	return nil
}
