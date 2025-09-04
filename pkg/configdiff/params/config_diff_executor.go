package params

import (
	"context"

	"github.com/sdcio/data-server/pkg/tree/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type Executor interface {
	GetTreeJson(ctx context.Context, path *sdcpb.Path) (any, error)
	GetDiff(ctx context.Context, dc *DiffConfig) (string, error)
	SchemaDownload(ctx context.Context, schemaDefinition *SchemaLoadConfig) (*sdcpb.Schema, error)
	TreeLoadData(ctx context.Context, cl *ConfigLoad) error
	TreeGetString(ctx context.Context, config *ConfigShowConfig) (string, error)
	TreeValidate(ctx context.Context) (types.ValidationResults, *types.ValidationStatOverall, error)
	TreeBlame(ctx context.Context, cb *ConfigBlameParams) (*sdcpb.BlameTreeElement, error)
}
