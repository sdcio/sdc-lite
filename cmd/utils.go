package cmd

import (
	"fmt"
	"os"

	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/pipeline"
)

func WriteOutput(o interfaces.Output) (err error) {
	if o == nil {
		return nil
	}
	switch {
	case jsonOutput:
		err = o.WriteToJson(os.Stdout)
	case detailed:
		_, err = fmt.Print(o.ToStringDetails())
	default:
		_, err = fmt.Print(o.ToString())
	}
	return err
}

func RunFromRaw(ctx context.Context, opts config.ConfigOpts, optsP config.ConfigPersistentOpts, persist bool, rpcParams ...params.RpcRawParams) (interfaces.Output, error) {

	c, err := config.NewConfigPersistent(opts, optsP)
	if err != nil {
		return nil, err
	}

	cd, err := configdiff.NewConfigDiffPersistence(ctx, c)
	if err != nil {
		return nil, err
	}
	err = cd.InitTargetFolder(ctx)
	if err != nil {
		return nil, err
	}

	outs := []interfaces.Output{}

	for _, rpcParam := range rpcParams {
		rpc, err := rpcParam.UnRaw()
		if err != nil {
			return nil, err
		}

		out, err := rpc.Run(ctx, cd)
		if err != nil {
			return nil, err
		}
		outs = append(outs, out)
	}
	if persist {
		err = cd.Persist(ctx)
		if err != nil {
			return nil, err
		}
	}
	// TODO
	return outs[0], nil
}

func AppendToPipelineFile(pipelineFile string, ps pipeline.PipelineStep) error {
	pipel := pipeline.NewPipeline(pipelineFile)
	return pipel.AppendStep(ps)
}
