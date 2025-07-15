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

type Workspace struct {
	config  WorkspaceConfig
	schema  *sdcpb.Schema
	intents Intents
}

func NewWorkspace(w WorkspaceConfig) *Workspace {
	wi := &Workspace{
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

func (wi *Workspace) Create() error {
	err := utils.CreateFolder(wi.config.WorkspacePath())
	if err != nil {
		return err
	}
	return nil
}

func (wi *Workspace) AddIntent(i *Intent) error {
	wi.intents.AddIntent(i)
	return nil
}

func (wi *Workspace) DeleteIntent(name string) error {
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

func (wi *Workspace) GetIntents() Intents {
	return wi.intents
}

func (wi *Workspace) schemaPersist() error {
	schemaByte, err := protojson.Marshal(wi.schema)
	if err != nil {
		return err
	}

	err = os.WriteFile(wi.config.SchemaDefinitionFilePath(), schemaByte, 0644)
	return err
}

func (wi *Workspace) intentsPersist() error {
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

func (wi *Workspace) SetSchema(s *sdcpb.Schema) {
	wi.schema = s
}

func (wi *Workspace) Persist() error {
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

func (wi *Workspace) loadIntent() error {
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

func (wi *Workspace) loadSchemaInfo() (err error) {
	wi.schema, err = utils.SchemaLoadSdcpbSchemaFile(wi.config.SchemaDefinitionFilePath())
	if err != nil {
		return err
	}
	return nil
}

func (wi *Workspace) GetSchema() *sdcpb.Schema {
	return wi.schema
}

func (wi *Workspace) String() string {
	sb := strings.Builder{}
	indentWorkspace := ""
	indentWorkspaceInfos := fmt.Sprintf("%s  ", indentWorkspace)
	indentWorkspaceIntent := fmt.Sprintf("%s  ", indentWorkspaceInfos)
	indentWorkspaceIntentInfos := fmt.Sprintf("%s  ", indentWorkspaceIntent)
	sb.WriteString(fmt.Sprintf("%sWorkspace: %s (%s)\n", indentWorkspace, wi.config.WorkspaceName(), wi.config.WorkspacePath()))
	if wi.schema != nil {
		sb.WriteString(fmt.Sprintf("%sSchema:\n", indentWorkspaceIntent))
		sb.WriteString(fmt.Sprintf("%sName: %s\n", indentWorkspaceIntentInfos, wi.schema.GetVendor()))
		sb.WriteString(fmt.Sprintf("%sVersion: %s\n", indentWorkspaceIntentInfos, wi.schema.GetVersion()))
	}
	for _, i := range wi.intents {
		sb.WriteString(fmt.Sprintf("%sIntent: %s\n", indentWorkspaceIntent, i.GetName()))
		sb.WriteString(fmt.Sprintf("%sPrio: %d\n", indentWorkspaceIntentInfos, i.GetPrio()))
		sb.WriteString(fmt.Sprintf("%sFlag: %s\n", indentWorkspaceIntentInfos, i.GetFlag()))
		sb.WriteString(fmt.Sprintf("%sFormat: %s\n", indentWorkspaceIntentInfos, i.GetFormat()))
	}

	return sb.String()
}

type Workspaces []*Workspace

func (wis *Workspaces) Add(w *Workspace) {
	*wis = append(*wis, w)
}

func (wis *Workspaces) String() string {
	sb := &strings.Builder{}
	for _, w := range *wis {
		sb.WriteString(w.String())
	}
	return sb.String()
}

type WorkspaceConfig interface {
	WorkspaceName() string
	ConfigFileName(intentName string) string
	ConfigFileGlob() string
	SchemaDefinitionFilePath() string
	WorkspacePath() string
}
