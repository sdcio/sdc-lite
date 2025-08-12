package cmd

import (
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	outFormatStr string
	outFormat    types.ConfigFormat
	targetName   string
)

// datastoreCmd represents the datastore command
var configCmd = &cobra.Command{
	Use:          "config",
	Short:        "config based actions",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		optsP = append(optsP, config.WithSuccessfullSchemaLoad(), config.WithTargetName(targetName))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	EnableFlagAndDisableFileCompletion(configCmd)
	err := AddTargetPersistentFlag(configCmd)
	if err != nil {
		log.Error(err)
	}
}

func parseConfigFormat() (types.ConfigFormat, error) {
	var err error
	outFormat, err = types.ParseConfigFormat(outFormatStr)
	if err != nil {
		return types.ConfigFormatUnknown, err
	}
	return outFormat, nil
}
