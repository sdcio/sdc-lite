package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sdcio/config-diff/pkg/configdiff"
	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/types"
	"github.com/sdcio/config-server/apis/config/v1alpha1"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"
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

		ctx := context.Background()

		opts := config.ConfigOpts{}
		c, err := config.NewConfigPersistent(opts, optsP)
		if err != nil {
			return err
		}

		cdp, err := configdiff.NewConfigDiffPersistence(ctx, c)
		if err != nil {
			return err
		}
		err = cdp.InitTargetFolder(ctx)
		if err != nil {
			return err
		}
		configFormat, err := types.ParseConfigFormat(configurationFileFormatStr)
		if err != nil {
			return err
		}

		var configByte []byte
		// process the input, read from file or stdin
		if configurationFile == "-" {
			// read from stdin
			configByte, err = io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		} else {
			// read config from file
			configByte, err = os.ReadFile(configurationFile)
			if err != nil {
				return err
			}
		}

		var intent *types.Intent

		switch configFormat {
		case types.ConfigFormatJson, types.ConfigFormatJsonIetf, types.ConfigFormatXml:
			intent = types.NewIntent(intentName, priority, treetypes.NewUpdateInsertFlags())
			intent.SetData(configFormat, configByte)
		case types.ConfigFormatSdc:

			sdcC, err := LoadSDCConfigCR(configByte)
			if err != nil {
				return err
			}
			intent, err = ConvertSDCConfigToInternalIntent(ctx, cdp.ConfigDiff, sdcC)
			if err != nil {
				return err
			}
		}
		err = cdp.TreeLoadData(ctx, intent)
		if err != nil {
			return err
		}

		err = cdp.Persist(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("File: %s - %s - successfully loaded\n", configurationFile, intent)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configLoadCmd)
	configLoadCmd.Flags().StringVar(&configurationFile, "file", "", "The configuration file to load")
	configLoadCmd.Flags().StringVar(&configurationFileFormatStr, "file-format", "", fmt.Sprintf("The format of the config to be loaded [ %s ]", strings.Join(types.ConfigFormatsList.StringSlice(), ", ")))
	configLoadCmd.Flags().Int32Var(&priority, "priority", 500, "The defined priority of the configuration")
	configLoadCmd.Flags().StringVar(&intentName, "intent-name", "", "The name of the configuration intent")
	EnableFlagAndDisableFileCompletion(configLoadCmd)
}

func LoadSDCConfigCR(configByte []byte) (*v1alpha1.Config, error) {
	// // Set up scheme
	// scheme := runtime.NewScheme()
	// _ = v1alpha1.AddToScheme(scheme)

	// // Setup decoder
	// dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(configByte), 4096)
	// serializer := k8sJson.NewYAMLSerializer(k8sJson.DefaultMetaFactory, scheme, scheme)

	// for {
	// 	// Dynamically decode into raw runtime.Object
	// 	var rawObj runtime.RawExtension
	// 	if err := dec.Decode(&rawObj); err != nil {
	// 		if err.Error() == "EOF" {
	// 			break
	// 		}
	// 		fmt.Fprintf(os.Stderr, "decode error: %v\n", err)
	// 		break
	// 	}

	// 	obj, gvk, err := serializer.Decode(rawObj.Raw, nil, nil)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "serializer decode error: %v\n", err)
	// 		continue
	// 	}

	// 	fmt.Printf("Loaded object of kind: %s\n", gvk.Kind)

	// 	// Optionally: cast to your known types
	// 	switch o := obj.(type) {
	// 	case *v1alpha1.Config:
	// 		fmt.Printf("Parsed Config: %+v\n", o.Spec)
	// 	case *v1alpha1.ConfigSet:
	// 		fmt.Printf()
	// 	default:
	// 		fmt.Printf("skipping unknown type: %s\n", obj.GetObjectKind().GroupVersionKind().String())
	// 	}
	// }
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

func ConvertSDCConfigToInternalIntent(ctx context.Context, cdp *configdiff.ConfigDiff, sdcConfig *v1alpha1.Config) (*types.Intent, error) {
	// create a new config diff instance that we can use to aggregate the multiple path / values from the cr spec
	cd, err := cdp.CopyEmptyConfigDiff(ctx)
	if err != nil {
		return nil, err
	}

	// create the
	intent := types.NewIntent(sdcConfig.Name, int32(sdcConfig.Spec.Priority), treetypes.NewUpdateInsertFlags())
	for _, c := range sdcConfig.Spec.Config {
		jsonConfigVal := c.Value.Raw
		if err != nil {
			return nil, err
		}
		intent.SetData(types.ConfigFormatJson, jsonConfigVal).SetBasePath(c.Path)

		err = cd.TreeLoadData(ctx, intent)
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

	intent = types.NewIntent(sdcConfig.Name, int32(sdcConfig.Spec.Priority), treetypes.NewUpdateInsertFlags())
	intent.SetData(types.ConfigFormatJson, jsonConfByte)

	return intent, nil
}
