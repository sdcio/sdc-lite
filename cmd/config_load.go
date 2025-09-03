package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sdcio/config-server/apis/config/v1alpha1"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sdcio/sdc-lite/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var configurationFileFormatStr string
var configurationFile string
var intentName string
var priority int32

// configLoadCmd represents the list command
var configLoadCmd = &cobra.Command{
	Use:          "load",
	Short:        "load config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		fmt.Fprintf(os.Stderr, "Target: %s\n", targetName)

		ctx := cmd.Context()

		opts := config.ConfigOpts{}
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cdp, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}

		intentRaw := params.NewConfigLoadRaw()
		intentRaw.SetFormat(configurationFileFormatStr).SetFile(configurationFile).SetFlags(&params.UpdateInsertFlagsRaw{})

		configFormat, err := types.ParseConfigFormat(configurationFileFormatStr)
		if err != nil {
			return err
		}

		fw := utils.NewFileWrapper(configurationFile)
		configByte, err := fw.Bytes()
		if err != nil {
			return err
		}

		switch configFormat {
		case types.ConfigFormatJson, types.ConfigFormatJsonIetf, types.ConfigFormatXml:
			// TODO:
			// lc.SetFlags()
			intentRaw.SetName(intentName).SetPrio(priority)
			// if data is comming from stdin, store it in data
			if strings.TrimSpace(configurationFile) == "-" {
				intentRaw.SetData(configByte)
			}
		case types.ConfigFormatSdc:

			sdcC, err := LoadSDCConfigCR(configByte)
			if err != nil {
				return err
			}
			intentRaw, err = ConvertSDCConfigToInternalIntent(ctx, cdp.ConfigDiff, sdcC)
			if err != nil {
				return err
			}
		}

		// if pipelineFile is set, then we need to generate just the pieline instruction equivalent of the actual command and exist
		if pipelineFile != "" {
			pipel := pipeline.NewPipeline(pipelineFile)
			pipel.AppendStep(intentRaw)
			return nil
		}

		err = cdp.InitTargetFolder(ctx)
		if err != nil {
			return err
		}

		configLoad, err := intentRaw.UnRaw()
		if err != nil {
			return err
		}

		err = cdp.TreeLoadData(ctx, configLoad)
		if err != nil {
			return err
		}

		err = cdp.Persist(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("File: %s - %s - successfully loaded\n", configurationFile, configLoad)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configLoadCmd)
	configLoadCmd.Flags().StringVar(&configurationFile, "file", "", "The configuration file to load")
	configLoadCmd.Flags().StringVar(&configurationFileFormatStr, "file-format", "", fmt.Sprintf("The format of the config to be loaded [ %s ]", strings.Join(types.ConfigFormatsList.StringSlice(), ", ")))
	configLoadCmd.Flags().Int32Var(&priority, "priority", 500, "The defined priority of the configuration")
	configLoadCmd.Flags().StringVar(&intentName, "intent-name", "", "The name of the configuration intent")
	AddPipelineCommandOutputFlags(configLoadCmd)
	EnableFlagAndDisableFileCompletion(configLoadCmd)

	params.GetCommandRegistry().Register(types.CommandTypeConfigLoad, func() params.RpcRawParams { return params.NewConfigLoadRaw() })
}

func LoadSDCConfigCR(configByte []byte) (*v1alpha1.Config, error) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	decoder := k8sJson.NewYAMLSerializer(k8sJson.DefaultMetaFactory, scheme, scheme)

	sdcC := &v1alpha1.Config{}
	_, _, err := decoder.Decode(configByte, nil, sdcC)
	if err != nil {
		return nil, err
	}
	return sdcC, nil
}

func ConvertSDCConfigToInternalIntent(ctx context.Context, cdp *configdiff.ConfigDiff, sdcConfig *v1alpha1.Config) (*params.ConfigLoadRaw, error) {
	// create a new config diff instance that we can use to aggregate the multiple path / values from the cr spec
	cd, err := cdp.CopyEmptyConfigDiff(ctx)
	if err != nil {
		return nil, err
	}

	// create the intent
	intent := types.NewIntent(sdcConfig.Name, int32(sdcConfig.Spec.Priority), treetypes.NewUpdateInsertFlags())
	for _, c := range sdcConfig.Spec.Config {
		jsonConfigVal := c.Value.Raw
		if err != nil {
			return nil, err
		}
		intent.SetData(types.ConfigFormatJson, jsonConfigVal).SetBasePath(c.Path)

		err = cd.TreeLoadData(ctx, params.NewConfigLoad(intent))
		if err != nil {
			return nil, err
		}
	}

	jsonConf, err := cd.GetJson(false)
	if err != nil {
		return nil, err
	}

	jsonConfByte, err := json.Marshal(jsonConf)
	if err != nil {
		return nil, err
	}

	intentRaw := params.NewConfigLoadRaw()
	intentRaw.SetName(sdcConfig.Name).SetPrio(int32(sdcConfig.Spec.Priority)).SetFlags(&params.UpdateInsertFlagsRaw{})
	intentRaw.SetFormat(types.ConfigFormatJson.String()).SetData(jsonConfByte)

	return intentRaw, nil
}
