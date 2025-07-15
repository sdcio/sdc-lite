package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sdcio/config-diff/pkg/configdiff/config"
	"github.com/sdcio/config-diff/pkg/utils"
)

type WorkspaceInfo struct {
	name          string
	path          string
	schemaName    string
	schemaVersion string
	intents       IntentInfos
}

func NewWorkspaceInfo(name string, path string) *WorkspaceInfo {
	wi := &WorkspaceInfo{
		name: name,
		path: path,
	}
	err := wi.loadSchemaInfo()
	if err != nil {
		return wi
	}
	err = wi.loadIntentInfos()
	if err != nil {
		return wi
	}

	return wi
}

func (wi *WorkspaceInfo) Create() error {
	err := utils.CreateFolder(wi.path)
	if err != nil {
		return err
	}
	return nil
}

func (wi *WorkspaceInfo) loadIntentInfos() error {
	matches, err := filepath.Glob(path.Join(wi.path, config.IntentFileGlob))
	if err != nil {
		return err
	}
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		ii := &IntentInfo{}
		err = json.Unmarshal(data, ii)
		if err != nil {
			return err
		}
		wi.intents.AddIntentInfo(ii)
	}
	return nil
}

func (wi *WorkspaceInfo) loadSchemaInfo() (err error) {
	schema, err := utils.SchemaLoadSdcpbSchemaFile(path.Join(wi.path, config.SchemaFileName))
	if err != nil {
		return err
	}

	wi.schemaName = schema.GetVendor()
	wi.schemaVersion = schema.GetVersion()
	return nil
}

func (wi *WorkspaceInfo) String() string {
	sb := strings.Builder{}
	indentWorkspace := ""
	indentWorkspaceInfos := fmt.Sprintf("%s  ", indentWorkspace)
	indentWorkspaceIntent := fmt.Sprintf("%s  ", indentWorkspaceInfos)
	indentWorkspaceIntentInfos := fmt.Sprintf("%s  ", indentWorkspaceIntent)
	sb.WriteString(fmt.Sprintf("%sWorkspace: %s\n", indentWorkspace, wi.name))
	if wi.schemaName != "" {
		sb.WriteString(fmt.Sprintf("%sSchema:\n", indentWorkspaceIntent))
		sb.WriteString(fmt.Sprintf("%sName: %s\n", indentWorkspaceIntentInfos, wi.schemaName))
		sb.WriteString(fmt.Sprintf("%sVersion: %s\n", indentWorkspaceIntentInfos, wi.schemaVersion))
	}
	sep := ""
	idx := 0
	for _, i := range wi.intents {
		if idx != 0 {
			sep = "-------\n"
		}
		idx++
		sb.WriteString(fmt.Sprintf("%sIntent: %s\n", indentWorkspaceIntent, i.GetName()))
		sb.WriteString(fmt.Sprintf("%sPrio: %d\n", indentWorkspaceIntentInfos, i.GetPrio()))
		sb.WriteString(fmt.Sprintf("%sFlag: %s\n", indentWorkspaceIntentInfos, i.GetFlag()))
		sb.WriteString(fmt.Sprintf("%sFormat: %s\n", indentWorkspaceIntentInfos, i.GetFormat()))
		sb.WriteString(sep)
	}

	return sb.String()
}

type WorkspaceInfos []*WorkspaceInfo

func (wis *WorkspaceInfos) Add(w *WorkspaceInfo) {
	*wis = append(*wis, w)
}

func (wis *WorkspaceInfos) String() string {
	sb := &strings.Builder{}
	for _, w := range *wis {
		sb.WriteString(w.String())
	}
	return sb.String()
}
