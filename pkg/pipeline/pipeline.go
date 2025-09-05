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

	cmdReg := params.GetCommandRegistry()

	fileRC, err := fw.ReadCloser()
	if err != nil {
		return err
	}

	cdc, err := config.NewConfig()
	if err != nil {
		return err
	}
	cd, err := configdiff.NewConfigDiff(ctx, cdc)
	if err != nil {
		return err
	}

	jd := json.NewDecoder(fileRC)

	fi, err := os.Stat(p.filename)
	if err != nil {
		return err
	}

	isFIFO := (fi.Mode() & os.ModeNamedPipe) != 0

	step := 1
	for {
		envelope := &params.JsonRpcMessageRaw{}
		if err := jd.Decode(envelope); err != nil {
			if errors.Is(err, io.EOF) {
				if isFIFO {
					// Reopen FIFO and wait for new writer
					fileRC.Close()
					fileRC, err = fw.ReadCloser()
					if err != nil {
						return err
					}
					jd = json.NewDecoder(fileRC)
					continue
				} else {
					err = fileRC.Close()
					if err != nil {
						return err
					}
					break
				}
			}
			return err
		}

		// stop command stops the pipeline
		if envelope.Method == types.CommandTypePipelineStop {
			fmt.Fprintln(os.Stderr, "stop command received, closing!")
			break
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

func PipelineAppendStep(file *os.File, s PipelineStep) error {

	// wrap in the JsonRPCHeader
	jrpcr := params.NewJsonRpcMessage(s.GetMethod(), rand.Int(), s)

	// json encode and write to file
	enc := json.NewEncoder(file)
	if err := enc.Encode(jrpcr); err != nil {
		return err
	}
	return nil
}

type PipelineStep interface {
	UnRaw() (params.RunCommand, error)
	GetMethod() types.CommandType
}
