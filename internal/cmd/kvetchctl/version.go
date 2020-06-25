package kvetchctl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	goversion "go.hein.dev/go-version"
)

var (
	version    = "dev"
	commit     = "none"
	date       = "unknown"
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  ``,
		PreRunE: func(command *cobra.Command, args []string) error {
			err := viper.BindPFlags(command.Flags())
			if err != nil {
				return err
			}
			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			shortened := viper.GetBool("short")
			var response string
			versionOutput := goversion.New(version, commit, date)

			if shortened {
				response = versionOutput.ToShortened()
			} else {
				response = versionOutput.ToJSON()
			}
			fmt.Printf("%+v", response)
			return
		},
	}
)

func init() {
	RootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("short", "s", false, "Use shortened output for version information.")
}
