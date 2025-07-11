package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	schemaFileName = "schema.json"
	intentFileName = "intent-%s.json"
	intentFileGlob = "intent-*.json"
)

type WorkspaceDisk struct {
	name          string
	workspacePath string
	schema        *sdcpb.Schema
	intents       IntentInfos
	hooks         WorkspaceHooks
}

func NewWorkspaceDisk() WorkspaceInit {
	return &WorkspaceDisk{
		intents: IntentInfos{},
	}
}

func (w *WorkspaceDisk) Init(p *WorkspaceInitParams) (Workspace, error) {
	w.name = p.Name
	w.workspacePath = p.WorkspaceBasePath
	return w, nil
}

func (w *WorkspaceDisk) SchemaStore(ctx context.Context, schema *sdcpb.Schema) error {
	w.schema = schema
	w.hooks.HookPostSchemaSet(ctx)
	return nil
}

func (w *WorkspaceDisk) HooksEndpointSet(ctx context.Context, wh WorkspaceHooks) {
	w.hooks = wh
	_ = w.schemaLoad(ctx)
}

func (w *WorkspaceDisk) SchemaGet(ctx context.Context) (*sdcpb.Schema, error) {
	if w.schema != nil {
		return w.schema, nil
	}
	err := w.schemaLoad(ctx)
	if err != nil {
		return nil, err
	}

	return w.schema, nil
}

func (w *WorkspaceDisk) schemaPersist() error {
	schemaByte, err := protojson.Marshal(w.schema)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(w.workspacePath, schemaFileName), schemaByte, 0644)
	return err
}

func (w *WorkspaceDisk) schemaLoad(ctx context.Context) error {
	schemaByte, err := os.ReadFile(path.Join(w.workspacePath, schemaFileName))
	if err != nil {
		return err
	}

	schema := &sdcpb.Schema{}
	err = protojson.Unmarshal(schemaByte, schema)
	if err != nil {
		return err
	}
	w.schema = schema
	w.hooks.HookPostSchemaSet(ctx)
	return nil
}

func (w *WorkspaceDisk) IntentStore(ii *IntentInfo) error {
	w.intents[ii.Name] = ii
	return nil
}

func (w *WorkspaceDisk) IntentDelete(intentName string) error {
	_, exists := w.intents[intentName]
	if !exists {
		return fmt.Errorf("delete Intent - Intent %s not found", intentName)
	}
	delete(w.intents, intentName)

	filename := path.Join(w.workspacePath, fmt.Sprintf(intentFileName, intentName))
	// stat the file to check existence
	_, err := os.Stat(filename)
	// if no error, then file exist, go ahaead remove it
	if err == nil {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WorkspaceDisk) IntentsGet() (IntentInfos, error) {
	if len(w.intents) > 0 {
		return w.intents, nil
	}

	matches, err := filepath.Glob(path.Join(w.workspacePath, intentFileGlob))
	if err != nil {
		return nil, err
	}
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		ii := &IntentInfo{}
		err = json.Unmarshal(data, ii)
		if err != nil {
			return nil, err
		}
		err = w.IntentStore(ii)
		if err != nil {
			return nil, err
		}
	}

	return w.intents, nil
}

func (w *WorkspaceDisk) intentPersist() error {
	for name, content := range w.intents {
		data, err := json.Marshal(content)
		if err != nil {
			return err
		}
		err = os.WriteFile(path.Join(w.workspacePath, fmt.Sprintf(intentFileName, name)), data, 0644)
		if err != nil {
			return err
		}
	}
	// reset intents, to load them from disk on next access
	w.intents = nil
	return nil
}

// func (w *WorkspaceDisk) TreeGet(ctx context.Context) (*tree.RootEntry, error) {
// 	entries, err := os.ReadDir(w.workspacePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	pattern := regexp.MustCompile(treeFilePattern)
// 	for _, entry := range entries {
// 		if !entry.IsDir() && pattern.MatchString(entry.Name()) {

// 			fileContent, err := os.ReadFile(entry.Name())
// 			if err != nil {
// 				return nil, err
// 			}

//			}
//		}
//		return w.tree, nil
//	}
// func (w *WorkspaceDisk) treePersist() error {
// 	ownerPrioMap := map[string]int32{}
// 	w.tree.GetOwnerPriorityMap(ownerPrioMap)

// 	for owner, prio := range ownerPrioMap {
// 		persitTree, err := w.tree.TreeExport(owner, prio)
// 		if err != nil {
// 			return err
// 		}

// 		treeByte, err := protojson.Marshal(persitTree)
// 		if err != nil {
// 			return err
// 		}

//			err = os.WriteFile(path.Join(w.workspacePath, fmt.Sprintf(treeFileName, owner)), treeByte, 0x644)
//			if err != nil {
//				return err
//			}
//		}
//		return nil
//	}
func (w *WorkspaceDisk) GetName() string {
	return w.name
}

func (w *WorkspaceDisk) Persist() error {
	err := w.schemaPersist()
	if err != nil {
		return err
	}
	err = w.intentPersist()
	if err != nil {
		return err
	}
	return nil
}
