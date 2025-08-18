package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sdcio/config-diff/pkg/utils"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/encoding/protojson"
)

type Target struct {
	config  TargetConfig
	schema  *sdcpb.Schema
	intents Intents
}

func NewTarget(w TargetConfig) *Target {
	wi := &Target{
		config: w,
	}
	err := wi.loadSchemaInfo()
	if err != nil {
		return wi
	}
	err = wi.loadIntent()
	if err != nil {
		return wi
	}

	return wi
}

func (wi *Target) Create() error {
	err := utils.CreateFolder(wi.config.TargetPath())
	if err != nil {
		return err
	}
	return nil
}

func (wi *Target) AddIntent(i *Intent) error {
	wi.intents.AddIntent(i)
	return nil
}

func (wi *Target) DeleteIntent(name string) error {
	_, exists := wi.intents[name]
	if !exists {
		return fmt.Errorf("error deleting intent %s - not found", name)
	}
	delete(wi.intents, name)
	err := os.Remove(wi.config.ConfigFileName(name))
	if err != nil {
		return err
	}
	return nil
}

func (wi *Target) GetIntents() Intents {
	return wi.intents
}

func (wi *Target) schemaPersist() error {
	schemaByte, err := protojson.Marshal(wi.schema)
	if err != nil {
		return err
	}

	err = os.WriteFile(wi.config.SchemaDefinitionFilePath(), schemaByte, 0644)
	return err
}

func (wi *Target) intentsPersist() error {
	for name, content := range wi.intents {
		data, err := json.Marshal(content)
		if err != nil {
			return err
		}
		err = os.WriteFile(wi.config.ConfigFileName(name), data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (wi *Target) SetSchema(s *sdcpb.Schema) {
	wi.schema = s
}

func (wi *Target) Persist() error {
	err := wi.schemaPersist()
	if err != nil {
		return err
	}
	err = wi.intentsPersist()
	if err != nil {
		return err
	}
	return nil
}

func (wi *Target) loadIntent() error {
	matches, err := filepath.Glob(wi.config.ConfigFileGlob())
	if err != nil {
		return err
	}
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		ii := &Intent{}
		err = json.Unmarshal(data, ii)
		if err != nil {
			return err
		}
		wi.intents.AddIntent(ii)
	}
	return nil
}

func (wi *Target) loadSchemaInfo() (err error) {
	wi.schema, err = utils.SchemaLoadSdcpbSchemaFile(wi.config.SchemaDefinitionFilePath())
	if err != nil {
		return err
	}
	return nil
}

func (wi *Target) GetSchema() *sdcpb.Schema {
	return wi.schema
}

func (wi *Target) String() string {
	schemaDetail := "unknown"
	if wi.schema != nil {
		schemaDetail = fmt.Sprintf("%s %s", wi.schema.Vendor, wi.schema.Version)
	}
	return fmt.Sprintf("%s [ %s ]\n", wi.config.TargetName(), schemaDetail)
}

func (wi *Target) StringDetail() string {
	sb := strings.Builder{}
	indentTarget := ""
	indentTargetInfos := fmt.Sprintf("%s  ", indentTarget)
	indentTargetIntent := fmt.Sprintf("%s  ", indentTargetInfos)
	indentTargetIntentInfos := fmt.Sprintf("%s  ", indentTargetIntent)
	sb.WriteString(fmt.Sprintf("%sTarget: %s (%s)\n", indentTarget, wi.config.TargetName(), wi.config.TargetPath()))
	if wi.schema != nil {
		sb.WriteString(fmt.Sprintf("%sSchema:\n", indentTargetIntent))
		sb.WriteString(fmt.Sprintf("%sName: %s\n", indentTargetIntentInfos, wi.schema.GetVendor()))
		sb.WriteString(fmt.Sprintf("%sVersion: %s\n", indentTargetIntentInfos, wi.schema.GetVersion()))
	}
	for _, i := range wi.intents {
		sb.WriteString(fmt.Sprintf("%sIntent: %s\n", indentTargetIntent, i.GetName()))
		sb.WriteString(fmt.Sprintf("%sPrio: %d\n", indentTargetIntentInfos, i.GetPrio()))
		sb.WriteString(fmt.Sprintf("%sFlag: %s\n", indentTargetIntentInfos, i.GetFlag()))
		sb.WriteString(fmt.Sprintf("%sFormat: %s\n", indentTargetIntentInfos, i.GetFormat()))
	}

	return sb.String()
}

type Targets []*Target

func (t *Targets) Add(w *Target) {
	*t = append(*t, w)
}

func (t Targets) String() string {
	sb := &strings.Builder{}
	for _, w := range t {
		sb.WriteString(w.String())
	}
	return sb.String()
}

func (t Targets) StringDetail() string {
	sb := &strings.Builder{}
	for _, w := range t {
		sb.WriteString(w.StringDetail())
	}
	return sb.String()
}

type TargetConfig interface {
	TargetName() string
	ConfigFileName(intentName string) string
	ConfigFileGlob() string
	SchemaDefinitionFilePath() string
	TargetPath() string
}
