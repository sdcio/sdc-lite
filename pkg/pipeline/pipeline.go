package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
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

// Pipeline executes JSON-RPC 2.0 Steps from a file, stdin, or FIFO against a Workspace.
// Batch runs (file or stdin) abort on the first Step error; FIFO runs report the error
// and continue serving subsequent Steps until pipeline-stop.
type Pipeline struct {
	filename string
	// persistentOpts, when set, selects the persisted Target the Workspace starts from.
	persistentOpts config.ConfigPersistentOpts
	// persist writes the resulting Workspace state back to the Target after a clean exit.
	persist bool
}

// PipelineOption configures a Pipeline.
type PipelineOption func(*Pipeline)

// WithTarget starts the pipeline from the persisted Target described by optsP.
// The Target must already exist with a Schema.
func WithTarget(optsP config.ConfigPersistentOpts) PipelineOption {
	return func(p *Pipeline) { p.persistentOpts = optsP }
}

// WithPersist writes the Workspace back to disk when the pipeline ends cleanly.
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

	if p.persist && p.persistentOpts == nil {
		outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, fmt.Errorf("persisting requires a target"))
		return
	}

	isFIFO := false
	if strings.TrimSpace(p.filename) != "-" && strings.TrimSpace(p.filename) != "" {
		fi, err := os.Stat(p.filename)
		if err != nil {
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
			return
		}
		isFIFO = (fi.Mode() & os.ModeNamedPipe) != 0
	}

	fw := utils.NewFileWrapper(p.filename)
	fileRC, err := fw.ReadCloser(ctx)
	if err != nil {
		outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
		return
	}

	cmdReg := commandregistry.GetCommandRegistry()

	var cd params.Executor
	var cdp *configdiff.ConfigDiffPersistence
	if p.persistentOpts != nil {
		cpc, err := config.NewConfigPersistent(config.ConfigOpts{}, p.persistentOpts)
		if err != nil {
			_ = fileRC.Close()
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
			return
		}
		cdp, err = configdiff.NewConfigDiffPersistence(ctx, cpc)
		if err != nil {
			_ = fileRC.Close()
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
			return
		}
		defer func() {
			if st := cdp.GetSchemaStore(); st != nil {
				_ = st.Close()
			}
		}()
		if err = cdp.InitTargetFolder(ctx); err != nil {
			_ = fileRC.Close()
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
			return
		}
		cd = cdp
	} else {
		cdc, err := config.NewConfig()
		if err != nil {
			_ = fileRC.Close()
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
			return
		}
		cd, err = configdiff.NewConfigDiff(ctx, cdc)
		if err != nil {
			_ = fileRC.Close()
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
			return
		}
	}

	reopenFIFO := func() bool {
		_ = fileRC.Close()
		var openErr error
		fileRC, openErr = fw.ReadCloser(ctx)
		if openErr != nil {
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, openErr)
			return false
		}
		return true
	}

	jd := json.NewDecoder(fileRC)
	step := 1
	for {
		select {
		case <-ctx.Done():
			_ = fileRC.Close()
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, ctx.Err())
			return
		default:
		}

		var rawStep json.RawMessage
		if err := jd.Decode(&rawStep); err != nil {
			if errors.Is(err, io.EOF) {
				if isFIFO {
					if !reopenFIFO() {
						return
					}
					jd = json.NewDecoder(fileRC)
					continue
				}
				if closeErr := fileRC.Close(); closeErr != nil {
					outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, closeErr)
					return
				}
				break
			}
			fmt.Fprintf(os.Stderr, "step %d parse error: %v\n", step, err)
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, !isFIFO, "", rpc.ErrCodeParseError, err)
			if isFIFO {
				if !reopenFIFO() {
					return
				}
				jd = json.NewDecoder(fileRC)
				continue
			}
			_ = fileRC.Close()
			return
		}

		trimmed := bytes.TrimSpace(rawStep)
		if len(trimmed) == 0 || trimmed[0] != '{' {
			err := fmt.Errorf("invalid JSON-RPC 2.0 request: expected JSON object")
			fmt.Fprintf(os.Stderr, "step %d invalid request: %v\n", step, err)
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, !isFIFO, "", rpc.ErrCodeInvalidRequest, err)
			if isFIFO {
				continue
			}
			_ = fileRC.Close()
			return
		}

		var envelope rpc.JsonRpcMessageRaw
		if err := json.Unmarshal(trimmed, &envelope); err != nil {
			fmt.Fprintf(os.Stderr, "step %d invalid request: %v\n", step, err)
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, !isFIFO, "", rpc.ErrCodeInvalidRequest, err)
			if isFIFO {
				continue
			}
			_ = fileRC.Close()
			return
		}

		if envelope.JsonRpc != "2.0" || envelope.Method == "" {
			respId := rpc.NullRpcId()
			if envelope.Id.IsSet() {
				respId = envelope.Id
			}
			err := fmt.Errorf("invalid JSON-RPC 2.0 request: jsonrpc must be \"2.0\" and method must be non-empty")
			fmt.Fprintf(os.Stderr, "step %d invalid request: %v\n", step, err)
			outputChan <- NewPipelineResultError(respId, true, !isFIFO, envelope.Method, rpc.ErrCodeInvalidRequest, err)
			if isFIFO {
				continue
			}
			_ = fileRC.Close()
			return
		}

		hasResp := envelope.Id.IsSet()

		paramsBytes := bytes.TrimSpace(envelope.Params)
		if len(paramsBytes) == 0 || bytes.Equal(paramsBytes, []byte("null")) {
			paramsBytes = []byte("{}")
		} else if paramsBytes[0] != '{' {
			err := fmt.Errorf("invalid params: expected JSON object")
			fmt.Fprintf(os.Stderr, "step %d (%s) invalid params: %v\n", step, envelope.Method, err)
			outputChan <- NewPipelineResultError(envelope.Id, hasResp, !isFIFO, envelope.Method, rpc.ErrCodeInvalidParams, err)
			if isFIFO {
				step++
				continue
			}
			_ = fileRC.Close()
			return
		}

		if envelope.Method == types.CommandTypePipelineStop {
			fmt.Fprintln(os.Stderr, "stop command received, closing!")
			if hasResp {
				outputChan <- NewPipelineResultOutput(envelope.Id, true, types.CommandTypePipelineStop, "", nil)
			}
			_ = fileRC.Close()
			break
		}

		factory, err := cmdReg.GetCommandFactory(envelope.Method)
		if err != nil {
			fmt.Fprintf(os.Stderr, "step %d (%s) error: %v\n", step, envelope.Method, err)
			outputChan <- NewPipelineResultError(envelope.Id, hasResp, !isFIFO, envelope.Method, rpc.ErrCodeMethodNotFound, err)
			if isFIFO {
				step++
				continue
			}
			_ = fileRC.Close()
			return
		}

		rawParams := factory()
		if err := json.Unmarshal(paramsBytes, rawParams); err != nil {
			fmt.Fprintf(os.Stderr, "step %d (%s) invalid params: %v\n", step, envelope.Method, err)
			outputChan <- NewPipelineResultError(envelope.Id, hasResp, !isFIFO, envelope.Method, rpc.ErrCodeInvalidParams, err)
			if isFIFO {
				step++
				continue
			}
			_ = fileRC.Close()
			return
		}

		cmd, err := rawParams.UnRaw()
		if err != nil {
			fmt.Fprintf(os.Stderr, "step %d (%s) invalid params: %v\n", step, envelope.Method, err)
			outputChan <- NewPipelineResultError(envelope.Id, hasResp, !isFIFO, envelope.Method, rpc.ErrCodeInvalidParams, err)
			if isFIFO {
				step++
				continue
			}
			_ = fileRC.Close()
			return
		}

		fmt.Fprintf(os.Stderr, "executing step: %d - %s\n", step, cmd.String())
		output, err := cmd.Run(ctx, cd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "step %d (%s) error: %v\n", step, envelope.Method, err)
			outputChan <- NewPipelineResultError(envelope.Id, hasResp, !isFIFO, envelope.Method, rpc.ErrCodeServerError, err)
			if isFIFO {
				step++
				continue
			}
			_ = fileRC.Close()
			return
		}

		var format types.ConfigFormat
		if fc, ok := cmd.(executor.FormattedCommand); ok {
			format = fc.ResultFormat()
		}

		if hasResp {
			outputChan <- NewPipelineResultOutput(envelope.Id, true, envelope.Method, format, output)
		}

		step++
	}

	if p.persist {
		if err := cdp.Persist(ctx); err != nil {
			outputChan <- NewPipelineResultError(rpc.NullRpcId(), true, true, "", rpc.ErrCodeServerError, err)
			return
		}
		fmt.Fprintln(os.Stderr, "target persisted")
	}
}

// PipelineAppendStep writes s as a single JSON-RPC 2.0 Request line with a 32-bit positive id
// (safe across IEEE-754 float64 consumers such as JavaScript and jq).
func PipelineAppendStep(file *os.File, s PipelineStep) error {
	id := rpc.NewRpcIdInt32(rand.Int31n(math.MaxInt32) + 1)
	jrpcr := rpc.NewJsonRpcMessage(s.GetMethod(), id, s)

	enc := json.NewEncoder(file)
	return enc.Encode(jrpcr)
}

type PipelineStep interface {
	UnRaw() (executor.RunCommand, error)
	GetMethod() types.CommandType
}
