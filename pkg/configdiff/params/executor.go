package params

import (
	"context"

	"github.com/beevik/etree"
	"github.com/sdcio/data-server/pkg/tree/api"
	"github.com/sdcio/data-server/pkg/tree/types"
	treetypes "github.com/sdcio/data-server/pkg/tree/types"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type Executor interface {
	GetTreeJson(ctx context.Context, path *sdcpb.Path) (any, error)
	GetDiff(ctx context.Context, dc *DiffConfig) (string, error)
	SchemaDownload(ctx context.Context, schemaDefinition *SchemaLoadConfig) (*sdcpb.Schema, error)
	TreeLoadData(ctx context.Context, cl *ConfigLoad) error
	TreeShow(ctx context.Context, config *ConfigShowConfig) (ConfigShowInterface, error)
	TreeValidate(ctx context.Context) (types.ValidationResults, *types.ValidationStats, error)
	TreeBlame(ctx context.Context, cb *ConfigBlameParams) (*sdcpb.BlameTreeElement, error)
}

type ConfigShowInterface interface {
	// ToJson returns the Tree contained structure as JSON
	// use e.g. json.MarshalIndent() on the returned struct
	ToJson(ctx context.Context, onlyNewOrUpdated bool) (any, error)
	// ToJsonIETF returns the Tree contained structure as JSON_IETF
	// use e.g. json.MarshalIndent() on the returned struct
	ToJsonIETF(ctx context.Context, onlyNewOrUpdated bool) (any, error)
	ToXML(ctx context.Context, onlyNewOrUpdated bool, honorNamespace bool, operationWithNamespace bool, useOperationRemove bool) (*etree.Document, error)
	// ToProto basically
	GetHighestPrecedence(ctx context.Context, onlyNewOrUpdated bool, includeDefaults bool, includeExplicitDelete bool) api.LeafVariantSlice
	// deletes for proto and json(_ietf)
	GetDeletes(ctx context.Context, aggregatePaths bool) (treetypes.DeleteEntriesList, error)
}
