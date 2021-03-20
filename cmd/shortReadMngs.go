package cmd

import (
	"github.com/spf13/cobra"
)

// shortReadMNGSCmd represents the shortReadMNGSCmd command
var shortReadMNGSCmd = &cobra.Command{
	Use:   "short-read-mngs",
	Short: "Commands related to the short-read-mngs pipeline",
	Long:  "Commands related to the short-read-mngs pipeline",
}

func init() {
	RootCmd.AddCommand(shortReadMNGSCmd)
}
