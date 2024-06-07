package amr

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.Set("ACCEPTED_USER_AGREEMENT", "Y")
	AmrCmd.PersistentFlags().Bool("verbose", false, "")
	code := m.Run()
	os.Exit(code)
}
