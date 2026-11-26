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
	"github.com/sdcio/sdc-lite/pkg/configdiff/commandregistry"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sdcio/sdc-lite/pkg/utils"
)

// Pipeline executes JSON-RPC steps from a file, stdin or FIFO against a ConfigDiff.
// By default it starts empty and in-memory; with a persistent target it starts from the
// target's stored schema and intents.
type Pipeline struct {
	filename string
	// persistentOpts, when set, selects the persisted target the pipeline starts from.
	persistentOpts config.ConfigPersistentOpts
	// persist writes the resulting state back to the target after a successful run.
	persist bool
}

// PipelineOption configures a Pipeline.
type PipelineOption func(*Pipeline)

// WithTarget starts the pipeline from the persisted target described by optsP.
// The target must already exist with a schema.
func WithTarget(optsP config.ConfigPersistentOpts) PipelineOption {
	return func(p *Pipeline) { p.persistentOpts = optsP }
}

// WithPersist writes the target back to disk when the pipeline ends without a step error.
// It requires WithTarget.
func WithPersist() PipelineOption {
	return func(p *Pipeline) { p.persist = true }
}

func NewPipeline(filename string, opts ...PipelineOption) *Pipeline {
	result := &Pipeline{
		filename: filename,
	}
	for _, o := range opts {
		o(result)
	}

	return result
}

func (p *Pipeline) Run(ctx context.Context, outputChan chan<- *PipelineResult) {
	defer close(outputChan)
	// Check if file exists
	fw := utils.NewFileWrapper(p.filename)

	cmdReg := commandregistry.GetCommandRegistry()

	fileRC, err := fw.ReadCloser(ctx)
	if err != nil {
		outputChan <- NewPipelineResultError(0, err)
		return
	}

	if p.persist && p.persistentOpts == nil {
		outputChan <- NewPipelineResultError(0, fmt.Errorf("persisting requires a target"))
		return
	}

	var cd params.Executor
	var cdp *configdiff.ConfigDiffPersistence
	if p.persistentOpts != nil {
		cpc, err := config.NewConfigPersistent(config.ConfigOpts{}, p.persistentOpts)
		if err != nil {
			outputChan <- NewPipelineResultError(0, err)
			return
		}
		cdp, err = configdiff.NewConfigDiffPersistence(ctx, cpc)
		if err != nil {
			outputChan <- NewPipelineResultError(0, err)
			return
		}
		defer func() {
			if st := cdp.GetSchemaStore(); st != nil {
				_ = st.Close()
			}
		}()
		if err = cdp.InitTargetFolder(ctx); err != nil {
			outputChan <- NewPipelineResultError(0, err)
			return
		}
		cd = cdp
	} else {
		cdc, err := config.NewConfig()
		if err != nil {
			outputChan <- NewPipelineResultError(0, err)
			return
		}
		cd, err = configdiff.NewConfigDiff(ctx, cdc)
		if err != nil {
			outputChan <- NewPipelineResultError(0, err)
			return
		}
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

	if p.persist {
		if err := cdp.Persist(ctx); err != nil {
			outputChan <- NewPipelineResultError(0, err)
			return
		}
		fmt.Fprintln(os.Stderr, "target persisted")
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
