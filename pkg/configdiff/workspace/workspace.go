package workspace

import (
	"context"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type Workspace interface {
	GetName() string
	// schemaInfo based fucntions
	SchemaStore(context.Context, *sdcpb.Schema) error
	SchemaGet(ctx context.Context) (*sdcpb.Schema, error)
	// Hooks
	HooksEndpointSet(context.Context, WorkspaceHooks)
	// intent based functions
	IntentStore(*IntentInfo) error
	IntentDelete(intenName string) error
	IntentsGet() (IntentInfos, error)
	// Persist General persist
	Persist() error
}

type WorkspaceInitParams struct {
	// The workspace name
	Name string
	// Persistence base folder in case persistence is needed
	WorkspaceBasePath string
}

type WorkspaceInit interface {
	Init(p *WorkspaceInitParams) (Workspace, error)
	Persist() error
}

type WorkspaceHooks interface {
	HookPostSchemaSet(ctx context.Context) error
}
