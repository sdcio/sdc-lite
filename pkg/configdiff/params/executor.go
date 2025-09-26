package params

import (
	"context"

	"github.com/beevik/etree"
	"github.com/sdcio/data-server/pkg/tree"
	"github.com/sdcio/data-server/pkg/tree/types"

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
	Walk(ctx context.Context, v tree.EntryVisitor) error
	// ToJson returns the Tree contained structure as JSON
	// use e.g. json.MarshalIndent() on the returned struct
	ToJson(onlyNewOrUpdated bool) (any, error)
	// ToJsonIETF returns the Tree contained structure as JSON_IETF
	// use e.g. json.MarshalIndent() on the returned struct
	ToJsonIETF(onlyNewOrUpdated bool) (any, error)
	ToXML(onlyNewOrUpdated bool, honorNamespace bool, operationWithNamespace bool, useOperationRemove bool) (*etree.Document, error)
	// ToProto basically
	GetHighestPrecedence(result tree.LeafVariantSlice, onlyNewOrUpdated bool, includeDefaults bool, includeExplicitDelete bool) tree.LeafVariantSlice
	// deletes for proto and json(_ietf)
	GetDeletes(deletes []types.DeleteEntry, aggregatePaths bool) ([]types.DeleteEntry, error)
}
