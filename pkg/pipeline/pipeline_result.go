package pipeline

import (
	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
)

// PipelineResult carries the outcome of one Pipeline Step (or a pipeline-level failure).
// HasResponse is false for JSON-RPC 2.0 Notifications (omitted "id"), and IsFatal marks
// failures that terminate a batch run and cause a non-zero CLI exit.
type PipelineResult struct {
	id          rpc.RpcId
	hasResponse bool
	fatal       bool
	method      types.CommandType
	format      types.ConfigFormat
	errCode     int
	err         error
	output      interfaces.Output
}

func NewPipelineResultError(id rpc.RpcId, hasResponse bool, fatal bool, method types.CommandType, errCode int, err error) *PipelineResult {
	return &PipelineResult{
		id:          id,
		hasResponse: hasResponse,
		fatal:       fatal,
		method:      method,
		errCode:     errCode,
		err:         err,
	}
}

func NewPipelineResultOutput(id rpc.RpcId, hasResponse bool, method types.CommandType, format types.ConfigFormat, output interfaces.Output) *PipelineResult {
	return &PipelineResult{
		id:          id,
		hasResponse: hasResponse,
		method:      method,
		format:      format,
		output:      output,
	}
}

func (p *PipelineResult) IsError() bool {
	return p.err != nil
}

func (p *PipelineResult) IsFatal() bool {
	return p.fatal
}

func (p *PipelineResult) HasResponse() bool {
	return p.hasResponse
}

func (p *PipelineResult) GetError() error {
	return p.err
}

func (p *PipelineResult) GetErrorCode() int {
	return p.errCode
}

func (p *PipelineResult) GetOutput() interfaces.Output {
	return p.output
}

func (p *PipelineResult) GetId() rpc.RpcId {
	return p.id
}

func (p *PipelineResult) GetMethod() types.CommandType {
	return p.method
}

func (p *PipelineResult) GetFormat() types.ConfigFormat {
	return p.format
}
