package cmd

import (
	"errors"
	"log"
	"strings"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/idseq"
	"github.com/spf13/cobra"
)

var listMetadataForHostGenome = &cobra.Command{
	Use:   "list-metadata-for-host-organism [host-organism-name]",
	Short: "List metadata options for a host organism",
	Long: `In IDSeq some metadata fields are only supported for
certain host organisms. This command lists the available fields
for a given host organism.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("missing required positional argument: host-organism-name")
		}
		client := idseq.DefaultClient
		fields, err := client.GetMetadataForHostGenome(args[0])
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range fields {
			if f.Name != "" {
				cmd.Println(f.Name)
			}

			if f.Description != "" {
				cmd.Printf("  %s\n", f.Description)
			}

			if len(f.Example.All) != 0 {
				cmd.Printf("  Options: %s\n", strings.Join(f.Example.All, ", "))
			} else if len(f.Example.One) != 0 {
				cmd.Printf("  Examples: %s\n", strings.Join(f.Example.One, ", "))
			}
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(listMetadataForHostGenome)
}
