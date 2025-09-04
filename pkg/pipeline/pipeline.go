package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sdcio/sdc-lite/pkg/utils"
)

type Pipeline struct {
	filename string
}

func NewPipeline(filename string) *Pipeline {
	result := &Pipeline{
		filename: filename,
	}

	return result
}

func (p *Pipeline) Run(ctx context.Context) error {
	// Check if file exists
	fw := utils.NewFileWrapper(p.filename)

	readCloser, err := fw.ReadCloser()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	cmdReg := params.GetCommandRegistry()

	jd := json.NewDecoder(readCloser)

	cdc, err := config.NewConfig()
	if err != nil {
		return err
	}
	cd, err := configdiff.NewConfigDiff(ctx, cdc)
	if err != nil {
		return err
	}

	step := 1
	for {
		envelope := &params.JsonRpcMessageRaw{}
		if err := jd.Decode(envelope); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		factory, err := cmdReg.GetCommandFactory(envelope.Method)
		if err != nil {
			return err
		}

		params := factory()

		// Unmarshal params only
		if err := json.Unmarshal(envelope.Params, params); err != nil {
			return err
		}

		// Execute command
		cmd, err := params.UnRaw()
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "executing step: %d - %s\n", step, cmd.String())
		output, err := cmd.Run(ctx, cd)
		if err != nil {
			return err
		}
		if output != nil {
			fmt.Println(output.ToStringDetails())
		}
		step++
	}

	return nil
}

func (p *Pipeline) AppendStep(s PipelineStep) error {
	f, err := os.OpenFile(p.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// wrap in the JsonRPCHeader
	jrpcr := params.NewJsonRpcMessage(s.GetMethod(), rand.Int(), s)

	// json encode and write to file
	enc := json.NewEncoder(f)
	if err := enc.Encode(jrpcr); err != nil {
		return err
	}
	return nil
}

type PipelineStep interface {
	UnRaw() (params.RunCommand, error)
	GetMethod() types.CommandType
}
