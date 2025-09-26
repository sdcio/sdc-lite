package cmd

import (
	"os"

	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rawparams"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	schemaDefinitionFile string
	schemaPathCleanup    bool
)

// SchemaLoadCmd represents the list command
var SchemaLoadCmd = &cobra.Command{
	Use:          "load",
	Short:        "load a schema",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		ctx := cmd.Context()

		rawParam := rawparams.NewSchemaLoadConfigRaw()
		rawParam.SetFile(schemaDefinitionFile)

		// if pipelineFile is set, then we need to generate just the pieline instruction equivalent of the actual command and exist
		if rpcOutput {
			return pipeline.PipelineAppendStep(os.Stdout, rawParam)
		}

		opts := config.ConfigOpts{
			// config.WithSchemaDefinition(schemaDefinitionFile),
			config.WithSchemaPathCleanup(schemaPathCleanup),
		}
		optsP = append(optsP, config.WithTargetName(targetName))

		out, err := RunFromRaw(ctx, opts, optsP, true, rawParam)
		if err != nil {
			return err
		}
		err = WriteOutput(out)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	schemaCmd.AddCommand(SchemaLoadCmd)
	SchemaLoadCmd.PersistentFlags().StringVarP(&schemaDefinitionFile, "schema-def", "f", "", "The krm that defined the schema")
	SchemaLoadCmd.PersistentFlags().BoolVarP(&schemaPathCleanup, "cleanup", "c", true, "Cleanup the Schemas directory after loading the schema")
	AddRpcOutputFlag(SchemaLoadCmd)
	EnableFlagAndDisableFileCompletion(SchemaLoadCmd)
	err := SchemaLoadCmd.MarkPersistentFlagRequired("schema-def")
	if err != nil {
		log.Error(err)
	}
	err = SchemaLoadCmd.MarkPersistentFlagFilename("schema-def")
	if err != nil {
		log.Error(err)
	}
	err = AddTargetPersistentFlag(SchemaLoadCmd)
	if err != nil {
		log.Error(err)
	}
}
