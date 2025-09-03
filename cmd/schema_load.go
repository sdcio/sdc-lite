package cmd

import (
	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/sdcio/sdc-lite/pkg/types"
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

		slcRaw := params.NewSchemaLoadConfigRaw()
		slcRaw.SetFile(schemaDefinitionFile)

		// if pipelineFile is set, then we need to generate just the pieline instruction equivalent of the actual command and exist
		if pipelineFile != "" {
			pipel := pipeline.NewPipeline(pipelineFile)
			pipel.AppendStep(slcRaw)
			return nil
		}

		log.Infof("Schema - Loading Start")
		slc, err := slcRaw.ToSchemaLoadConfig()
		if err != nil {
			return err
		}

		opts := config.ConfigOpts{
			// config.WithSchemaDefinition(schemaDefinitionFile),
			config.WithSchemaPathCleanup(schemaPathCleanup),
		}
		optsP = append(optsP, config.WithTargetName(targetName))
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cd, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}
		err = cd.InitTargetFolder(ctx)
		if err != nil {
			return err
		}

		// download the given schema
		_, err = cd.SchemaDownload(ctx, slc)
		if err != nil {
			return err
		}

		err = cd.Persist(ctx)
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
	AddPipelineCommandOutputFlags(SchemaLoadCmd)
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

	params.GetCommandRegistry().Register(types.CommandTypeSchemaLoad, func() params.RpcRawParams { return params.NewSchemaLoadConfigRaw() })
}
