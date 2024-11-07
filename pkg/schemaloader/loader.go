package schemaloader

import (
	"context"
	"os"
	"path/filepath"

	"github.com/henderiw/config-diff/pkg/utils"
	loader "github.com/sdcio/config-server/pkg/schema"
	"github.com/sdcio/schema-server/pkg/store"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
)

const (
	tmpPath     = "tmp/tmp"
	schemasPath = "tmp/schemas"
)

func New(schemastore store.Store) (*SchemaLoader, error) {
	if err := os.MkdirAll(tmpPath, 0755|os.ModeDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(schemasPath, 0755|os.ModeDir); err != nil {
		return nil, err
	}

	return &SchemaLoader{
		schemastore: schemastore,
	}, nil
}

type SchemaLoader struct {
	schemastore store.Store
}

func (r *SchemaLoader) LoadSchema(ctx context.Context, schemaConfigPath string) (*sdcpb.CreateSchemaResponse, error) {
	schemacr, err := utils.GetConfig(schemaConfigPath)
	if err != nil {
		return nil, err
	}

	schemaLoader, err := loader.NewLoader(
		filepath.Join(tmpPath),
		filepath.Join(schemasPath),
		NewNopResolver(),
	)
	if err != nil {
		return nil, err
	}

	schemaLoader.AddRef(ctx, schemacr)
	_, dirExists, err := schemaLoader.GetRef(ctx, schemacr.Spec.GetKey())
	if err != nil {
		return nil, err
	}
	if !dirExists {
		log.Info("loading...")
		if err := schemaLoader.Load(ctx, schemacr.Spec.GetKey()); err != nil {
			return nil, err
		}
	}

	return r.schemastore.CreateSchema(ctx, &sdcpb.CreateSchemaRequest{
		Schema: &sdcpb.Schema{
			Name:    "",
			Vendor:  schemacr.Spec.Provider,
			Version: schemacr.Spec.Version,
		},
		File:      schemacr.Spec.GetNewSchemaBase(schemasPath).Models,
		Directory: schemacr.Spec.GetNewSchemaBase(schemasPath).Includes,
		Exclude:   schemacr.Spec.GetNewSchemaBase(schemasPath).Excludes,
	})
}
