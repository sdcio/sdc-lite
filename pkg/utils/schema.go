package utils

import (
	"fmt"
	"os"

	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	"github.com/sdcio/config-server/pkg/utils"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func GetSchemaConfig(schemaConfigPath string) (*invv1alpha1.Schema, error) {
	b, err := os.ReadFile(schemaConfigPath)
	if err != nil {
		return nil, err
	}

	schema := &invv1alpha1.Schema{}
	if err := yaml.Unmarshal(b, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func SchemaLoadSdcpbSchemaFile(path string) (*sdcpb.Schema, error) {
	if !utils.FileExists(path) {
		errTxt := "schema defintion file does not exist. Need to load schema first"
		log.Info(errTxt)
		return nil, fmt.Errorf("%s", errTxt)
	}
	schemaByte, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	schema := &sdcpb.Schema{}
	err = protojson.Unmarshal(schemaByte, schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
