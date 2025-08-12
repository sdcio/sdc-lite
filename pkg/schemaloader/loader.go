package schemaloader

import (
	"context"
	"os"
	"path/filepath"

	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	loader "github.com/sdcio/config-server/pkg/schema"
	"github.com/sdcio/schema-server/pkg/store"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
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
	b, err := os.ReadFile(schemaConfigPath)
	if err != nil {
		panic(err)
	}

	schema := &invv1alpha1.Schema{}
	if err := yaml.Unmarshal(b, schema); err != nil {
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

	schemaLoader.AddRef(ctx, schema)
	_, dirExists, err := schemaLoader.GetRef(ctx, schema.Spec.GetKey())
	if err != nil {
		return nil, err
	}
	if !dirExists {
		log.Info("loading...")
		if _, err := schemaLoader.Load(ctx, schema.Spec.GetKey()); err != nil {
			return nil, err
		}
	}

	return r.schemastore.CreateSchema(ctx, &sdcpb.CreateSchemaRequest{
		Schema: &sdcpb.Schema{
			Name:    "",
			Vendor:  schema.Spec.Provider,
			Version: schema.Spec.Version,
		},
		File:      schema.Spec.GetNewSchemaBase(schemasPath).Models,
		Directory: schema.Spec.GetNewSchemaBase(schemasPath).Includes,
		Exclude:   schema.Spec.GetNewSchemaBase(schemasPath).Excludes,
	})
}
