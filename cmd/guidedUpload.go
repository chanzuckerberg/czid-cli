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
	_, err := io.WriteString(cmd.OutOrStdout(), message+"\n")
	if err != nil {
		log.Fatal(err)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.WriteString(cmd.OutOrStdout(), "\n")
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSuffix(input, "\n")
}

func optionsSelect(cmd *cobra.Command, reader *bufio.Reader, message string, options []string) string {
	optionsString := strings.Join(options, ", ")
	_, err := io.WriteString(cmd.OutOrStdout(), message+"\n")
	if err != nil {
		log.Fatal(err)
	}

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

func czidExec(cmd *cobra.Command, reason string, args ...string) {
	prettyArgs := make([]string, len(args))
	for i, s := range args {
		if strings.ContainsRune(s, ' ') {
			prettyArgs[i] = fmt.Sprintf("\"%s\"", s)
		} else {
			prettyArgs[i] = s
		}
	}
	msg := fmt.Sprintf("%s, running `czid %s`\n", reason, strings.Join(prettyArgs, " "))
	_, err := io.WriteString(cmd.OutOrStdout(), msg)
	if err != nil {
		log.Fatal(err)
	}

	RootCmd.SetArgs(args)
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// versionCmd represents the guided-upload command
var guidedUploadCmd = &cobra.Command{
	Use:   "guided-upload",
	Short: "guides you through an upload to CZID",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := auth0.DefaultClient.IDToken()
		if err != nil {
			czidExec(cmd, "You are not currently logged in", "login")
		}

		if strings.ToLower(viper.GetString("accepted_user_agreement")) != "y" {
			czidExec(cmd, "You have not yet accepted the user agreement", "accept-user-agreement")
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

		var sampleName string
		var dirname string
		if quantity == "single" {
			fOne := getInput(cmd, reader, "Enter R1 fielpath for a paired-end sample or filepath to single file for a single-end sample:")
			uploadArgs = append(uploadArgs, "upload-sample", fOne)

			fTwo := getInput(cmd, reader, "Enter R2 filepath for a paired-end sample or press enter for a single-end sample:")
			if fTwo != "" {
				uploadArgs = append(uploadArgs, fTwo)
			}

			sampleName = getInput(
				cmd,
				reader,
				"Enter a name for your sample:",
			)
			uploadArgs = append(uploadArgs, "--sample-name", sampleName)
		} else {
			dirname = getInput(cmd, reader, "Enter path to directory containing samples:")
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

		metadataMsg := `To continue you will need a metadata csv file
Here are instructions on how to make one: https://czid.org/metadata/instructions
Here is a dictionary of our supported metadata: https://czid.org/metadata/dictionary
Would you like to create one yourself or generate a template?`
		generate := optionsSelect(
			cmd,
			reader,
			metadataMsg,
			[]string{"self-create", "generate"},
		)

		var metadataFile string
		if generate == "self-create" {
			metadataFile = getInput(
				cmd,
				reader,
				"Enter path to your metadata csv:",
			)
		} else {
			metadataFile = getInput(
				cmd,
				reader,
				"Enter path you would like for your newly generated metadata csv:",
			)
			hostGenome := getInput(
				cmd,
				reader,
				"If your sample or samples are all from the same Host Organism please enter it below, otherwise press enter:",
			)
			templateArgs := []string{"-o", metadataFile}
			if hostGenome != "" {
				templateArgs = append(templateArgs, "-m", fmt.Sprintf("Host Organism=%s", hostGenome))
			}
			if quantity == "single" {
				templateArgs = append([]string{"generate-metadata-template", "for-sample-name", sampleName}, templateArgs...)
			} else {
				templateArgs = append([]string{"generate-metadata-template", "for-sample-directory", dirname}, templateArgs...)
			}
			czidExec(cmd, "Generating metadata template", templateArgs...)

			getInput(
				cmd,
				reader,
				fmt.Sprintf("Please fill in your template file at '%s', save, and press enter", metadataFile),
			)
		}

		uploadArgs = append(uploadArgs, "--metadata-csv", metadataFile)
		czidExec(cmd, "Performing your upload", uploadArgs...)
	},
}

func init() {
	RootCmd.AddCommand(guidedUploadCmd)
}
