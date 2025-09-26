// Copyright 2024 Nokia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package schemaclient

import (
	"context"
	"sync"

	schemaClient "github.com/sdcio/data-server/pkg/datastore/clients/schema"
	"github.com/sdcio/schema-server/pkg/store"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type SchemaClientBoundImpl struct {
	schemastore store.Store
	schemaRef   *sdcpb.Schema

	index      sync.Map // string -> schemaIndexEntry
	indexMutex sync.RWMutex
}

func NewMemSchemaClientBound(schemastore store.Store, schemaRef *sdcpb.Schema) schemaClient.SchemaClientBound {
	return &SchemaClientBoundImpl{
		schemastore: schemastore,
		schemaRef:   schemaRef,
		index:       sync.Map{},
	}
}

// GetSchemaSdcpbPath retrieves the schema for the given path
func (r *SchemaClientBoundImpl) GetSchemaSdcpbPath(ctx context.Context, path *sdcpb.Path) (*sdcpb.GetSchemaResponse, error) {

	// convert the path into a keyless path, for schema index lookups.
	keylessPath := path.ToXPath(true)

	entryAny, loaded := r.index.LoadOrStore(keylessPath, schemaClient.NewSchemaIndexEntry(nil, nil))
	entry := entryAny.(*schemaClient.SchemaIndexEntry)

	// Lock the entry to prevent race conditions
	entry.Lock()
	defer entry.Unlock()

	// if it existed, some other goroutine is already fetching the schema
	if loaded && entry.GetReady() {
		return entry.GetSchemaResponse(), entry.GetError()
	}

	// retrieve Schema via schemaclient
	schema, err := r.schemastore.GetSchema(ctx, &sdcpb.GetSchemaRequest{
		Schema:          r.schemaRef,
		Path:            path,
		WithDescription: false,
	})
	entry.SetSchemaResponseAndError(schema, err)

	return entry.Get()

}

func (r *SchemaClientBoundImpl) GetSchemaElements(ctx context.Context, p *sdcpb.Path, done chan struct{}) (chan *sdcpb.GetSchemaResponse, error) {
	gsr := &sdcpb.GetSchemaRequest{
		Path:   p,
		Schema: r.schemaRef,
	}
	och, err := r.schemastore.GetSchemaElements(ctx, gsr)
	if err != nil {
		return nil, err
	}
	ch := make(chan *sdcpb.GetSchemaResponse)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case se, ok := <-och:
				if !ok {
					return
				}
				ch <- &sdcpb.GetSchemaResponse{
					Schema: se,
				}
			}
		}
	}()
	return ch, nil
}
