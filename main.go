package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/henderiw/config-diff/schemaclient"
	"github.com/henderiw/config-diff/schemaloader"
	"github.com/henderiw/logger/log"
	"github.com/sdcio/data-server/pkg/tree"
	treejson "github.com/sdcio/data-server/pkg/tree/importer/json"
	"github.com/sdcio/schema-server/pkg/store/memstore"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"k8s.io/component-base/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	l := log.NewLogger(&log.HandlerOptions{Name: "config-server-logger", AddSource: false})
	slog.SetDefault(l)
	ctx := log.IntoContext(context.Background(), l)
	//log := log.FromContext(ctx)

	//opts := zap.Options{
	//	TimeEncoder: zapcore.RFC3339NanoTimeEncoder,
	//}
	args := os.Args
	if len(args) < 1 {
		panic("cannot execute need config and base dir")
	}

	schemastore := memstore.New()
	schemaLoader, err := schemaloader.New(schemastore)
	if err != nil {
		panic(err)
	}
	rsp, err := schemaLoader.LoadSchema(ctx, args[1])
	if err != nil {
		fmt.Println("rsp", rsp)
		panic(err)
	}
	scb := schemaclient.NewMemSchemaClientBound(schemastore, &sdcpb.Schema{
		Vendor:  rsp.Schema.Vendor,
		Version: rsp.Schema.Version,
	})
	fmt.Println(scb)

	tc := tree.NewTreeContext(tree.NewTreeSchemaCacheClient("dev1", nil, scb), "test")
	root, err := tree.NewTreeRoot(ctx, tc)
	if err != nil {
		panic(err)
	}
	fmt.Println(root.String())

	jsonBytes, err := os.ReadFile("data/config/running/running.json")
	if err != nil {
		panic(err)
	}

	var j any
	err = json.Unmarshal(jsonBytes, &j)
	if err != nil {
		panic(err)
	}
	jti := treejson.NewJsonTreeImporter(j)
	err = root.ImportConfig(ctx, jti, tree.RunningIntentName, tree.RunningValuesPrio)
	if err != nil {
		panic(err)
	}

	root.FinishInsertionPhase()
	fmt.Println(root.String())

}
