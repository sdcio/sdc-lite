package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/command_registry"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
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

func (p *Pipeline) Run(ctx context.Context, outputChan chan<- *PipelineResult) {
	defer close(outputChan)
	// Check if file exists
	fw := utils.NewFileWrapper(p.filename)

	cmdReg := command_registry.GetCommandRegistry()

	fileRC, err := fw.ReadCloser(ctx)
	if err != nil {
		outputChan <- NewPipelineResultError(0, err)
		return
	}

	cdc, err := config.NewConfig()
	if err != nil {
		outputChan <- NewPipelineResultError(0, err)
		return
	}
	cd, err := configdiff.NewConfigDiff(ctx, cdc)
	if err != nil {
		outputChan <- NewPipelineResultError(0, err)
		return
	}

	jd := json.NewDecoder(fileRC)

	isFIFO := false
	if strings.TrimSpace(p.filename) != "-" && strings.TrimSpace(p.filename) != "" {
		fi, err := os.Stat(p.filename)
		if err != nil {
			outputChan <- NewPipelineResultError(0, err)
			return
		}

		isFIFO = (fi.Mode() & os.ModeNamedPipe) != 0
	}

	step := 1
	for {

		// context cancellation check
		select {
		case <-ctx.Done():
			outputChan <- NewPipelineResultError(0, ctx.Err())
			return
		default:
		}

		envelope := &rpc.JsonRpcMessageRaw{}
		if err := jd.Decode(envelope); err != nil {
			if errors.Is(err, io.EOF) {
				if isFIFO {
					// Reopen FIFO and wait for new writer
					fileRC.Close()
					fileRC, err = fw.ReadCloser(ctx)
					if err != nil {
						outputChan <- NewPipelineResultError(0, err)
						return
					}
					// re-create the decoder
					jd = json.NewDecoder(fileRC)
					// start all over waiting for the next incomming rpc
					continue
				} else {
					err = fileRC.Close()
					if err != nil {
						outputChan <- NewPipelineResultError(0, err)
						return
					}
					break
				}
			}
			outputChan <- NewPipelineResultError(0, err)
			return
		}

		// stop command stops the pipeline
		if envelope.Method == types.CommandTypePipelineStop {
			fmt.Fprintln(os.Stderr, "stop command received, closing!")
			break
		}

		factory, err := cmdReg.GetCommandFactory(envelope.Method)
		if err != nil {
			outputChan <- NewPipelineResultError(envelope.Id, err)
			return
		}

		params := factory()

		// Unmarshal params only
		if err := json.Unmarshal(envelope.Params, params); err != nil {
			outputChan <- NewPipelineResultError(envelope.Id, err)
			return
		}

		// Execute command
		cmd, err := params.UnRaw()
		if err != nil {
			outputChan <- NewPipelineResultError(envelope.Id, err)
			return
		}

		fmt.Fprintf(os.Stderr, "executing step: %d - %s\n", step, cmd.String())
		output, err := cmd.Run(ctx, cd)
		if err != nil {
			outputChan <- NewPipelineResultError(envelope.Id, err)
			return
		}

		outputChan <- NewPipelineResultOutput(envelope.Id, output)

		step++
	}
}

func PipelineAppendStep(file *os.File, s PipelineStep) error {

	// wrap in the JsonRPCHeader
	jrpcr := rpc.NewJsonRpcMessage(s.GetMethod(), rand.Int(), s)

	// json encode and write to file
	enc := json.NewEncoder(file)
	if err := enc.Encode(jrpcr); err != nil {
		return err
	}
	return nil
}

type PipelineStep interface {
	UnRaw() (executor.RunCommand, error)
	GetMethod() types.CommandType
}
