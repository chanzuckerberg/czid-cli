package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chanzuckerberg/idseq-cli-v2/cmd/consensusGenome"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/auth0"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/util"
)

func getInput(cmd *cobra.Command, reader *bufio.Reader, message string) string {
	io.WriteString(cmd.OutOrStdout(), message+"\n")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSuffix(input, "\n")
}

func optionsSelect(cmd *cobra.Command, reader *bufio.Reader, message string, options []string) string {
	optionsString := strings.Join(options, ", ")
	io.WriteString(cmd.OutOrStdout(), message+"\n")
	opt := fmt.Sprintf("Enter one of (%s):", optionsString)
	input := getInput(cmd, reader, opt)

	for _, o := range options {
		if o == input {
			return input
		}
	}

	log.Fatalf("%s is not one of (%s)", input, options)
	return ""
}

// versionCmd represents the guided-upload command
var guidedUploadCmd = &cobra.Command{
	Use:   "guided-upload",
	Short: "guides you through an upload to IDSeq",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := auth0.DefaultClient.IDToken()
		if err != nil {
			io.WriteString(cmd.OutOrStdout(), "you are not currently logged in, please log in\n")
			err = auth0.DefaultClient.Login(false, false)
			if err != nil {
				log.Fatal(err)
			}
		}

		if strings.ToLower(viper.GetString("accepted_user_agreement")) != "y" {
			RootCmd.SetArgs([]string{"accept-user-agreement"})
			err = RootCmd.Execute()
			if err != nil {
				log.Fatal(err)
			}
		}

		reader := bufio.NewReader(cmd.InOrStdin())

		workflow := optionsSelect(
			cmd,
			reader,
			"What pipeline would you like to run on your sample?",
			[]string{"short-read-mngs", "consensus-genome"},
		)

		uploadArgs := []string{workflow}
		quantity := optionsSelect(
			cmd,
			reader,
			"Would you like to upload a single sample or many samples?",
			[]string{"single", "many"},
		)

		if quantity == "single" {
			fOne := getInput(cmd, reader, "Enter R1 fielpath for a paired-end sample or filepath to single file for a single-end sample:")
			uploadArgs = append(uploadArgs, "upload-sample", fOne)

			fTwo := getInput(cmd, reader, "Enter R2 filepath for a paired-end sample or press enter for a single-end sample:")
			if fTwo != "" {
				uploadArgs = append(uploadArgs, fTwo)
			}

			sampleName := getInput(
				cmd,
				reader,
				"Enter a name for your sample:",
			)
			uploadArgs = append(uploadArgs, "--sample-name", sampleName)
		} else {
			dirname := getInput(cmd, reader, "Enter path to directory containing samples:")
			uploadArgs = append(uploadArgs, "upload-samples", dirname)
		}

		projectName := getInput(
			cmd,
			reader,
			"Enter the name of the project to upload to:",
		)
		uploadArgs = append(uploadArgs, "--project", projectName)

		if workflow == "consensus-genome" {
			technology := optionsSelect(
				cmd,
				reader,
				"What sequencing platform did you use?",
				util.StringMapKeys(consensusGenome.Technologies),
			)
			uploadArgs = append(uploadArgs, "--sequencing-platform", technology)

			if technology == "Illumina" {
				wetlabProtocol := optionsSelect(
					cmd,
					reader,
					"What wetlab protocol did you follow?",
					util.StringMapKeys(consensusGenome.WetlabProtocols),
				)
				uploadArgs = append(uploadArgs, "--wetlab-protocol", wetlabProtocol)
			}
		}

		metadataFile := getInput(
			cmd,
			reader,
			"Enter metadata file path:",
		)

		uploadArgs = append(uploadArgs, "--metadata-csv", metadataFile)
		RootCmd.SetArgs(uploadArgs)
		err = RootCmd.Execute()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(guidedUploadCmd)
}
