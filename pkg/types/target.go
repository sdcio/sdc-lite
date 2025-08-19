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

func (t *Target) AddIntent(i *Intent) error {
	t.intents.AddIntent(i)
	return nil
}

func (t *Target) DeleteIntent(name string) error {
	_, exists := t.intents[name]
	if !exists {
		return fmt.Errorf("error deleting intent %s - not found", name)
	}
	delete(t.intents, name)
	err := os.Remove(t.config.ConfigFileName(name))
	if err != nil {
		return err
	}
	return nil
}

func (t *Target) GetIntents() Intents {
	return t.intents
}

func (t *Target) schemaPersist() error {
	schemaByte, err := protojson.Marshal(t.schema)
	if err != nil {
		return err
	}

	err = os.WriteFile(t.config.SchemaDefinitionFilePath(), schemaByte, 0644)
	return err
}

func (t *Target) intentsPersist() error {
	for name, content := range t.intents {
		data, err := json.Marshal(content)
		if err != nil {
			return err
		}
		err = os.WriteFile(t.config.ConfigFileName(name), data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Target) SetSchema(s *sdcpb.Schema) {
	t.schema = s
}

func (t *Target) Persist() error {
	err := utils.CreateFolder(t.config.TargetPath())
	if err != nil {
		return err
	}
	err = t.schemaPersist()
	if err != nil {
		return err
	}
	err = t.intentsPersist()
	if err != nil {
		return err
	}
	return nil
}

func (t *Target) loadIntent() error {
	matches, err := filepath.Glob(t.config.ConfigFileGlob())
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
		t.intents.AddIntent(ii)
	}
	return nil
}

func (t *Target) loadSchemaInfo() (err error) {
	t.schema, err = utils.SchemaLoadSdcpbSchemaFile(t.config.SchemaDefinitionFilePath())
	if err != nil {
		return err
	}
	return nil
}

func (t *Target) GetSchema() *sdcpb.Schema {
	return t.schema
}

func (t *Target) Export() *TargetExport {
	result := &TargetExport{
		TargetName: t.config.TargetName(),
		Intents:    t.intents.Export(),
		TargetPath: t.config.TargetPath(),
	}
	if t.schema != nil {
		result.Schema = &SchemaExport{
			Vendor:  t.schema.Vendor,
			Version: t.schema.Version,
		}
	}

	return result
}

func (t *Target) String() string {
	schemaDetail := "unknown"
	if t.schema != nil {
		schemaDetail = fmt.Sprintf("%s %s", t.schema.Vendor, t.schema.Version)
	}
	return fmt.Sprintf("%s [ %s ]\n", t.config.TargetName(), schemaDetail)
}

func (t *Target) StringDetail() string {
	sb := strings.Builder{}
	indentTarget := ""
	indentTargetInfos := fmt.Sprintf("%s  ", indentTarget)
	indentTargetIntent := fmt.Sprintf("%s  ", indentTargetInfos)
	indentTargetIntentInfos := fmt.Sprintf("%s  ", indentTargetIntent)
	sb.WriteString(fmt.Sprintf("%sTarget: %s (%s)\n", indentTarget, t.config.TargetName(), t.config.TargetPath()))
	if t.schema != nil {
		sb.WriteString(fmt.Sprintf("%sSchema:\n", indentTargetIntent))
		sb.WriteString(fmt.Sprintf("%sName: %s\n", indentTargetIntentInfos, t.schema.GetVendor()))
		sb.WriteString(fmt.Sprintf("%sVersion: %s\n", indentTargetIntentInfos, t.schema.GetVersion()))
	}
	for _, i := range t.intents {
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

func (t Targets) Export() []*TargetExport {
	result := make([]*TargetExport, 0, len(t))
	for _, target := range t {
		result = append(result, target.Export())
	}
	return result
}

type TargetConfig interface {
	TargetName() string
	ConfigFileName(intentName string) string
	ConfigFileGlob() string
	SchemaDefinitionFilePath() string
	TargetPath() string
}

type TargetExport struct {
	TargetName string          `json:"target"`
	TargetPath string          `json:"target-path"`
	Schema     *SchemaExport   `json:"schema"`
	Intents    []*IntentExport `json:"intents"`
}
