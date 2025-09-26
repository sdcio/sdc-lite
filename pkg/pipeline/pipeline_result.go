package pipeline

import "github.com/sdcio/sdc-lite/cmd/interfaces"

type PipelineResult struct {
	err    error
	output interfaces.Output
	id     int
}

func NewPipelineResultError(id int, err error) *PipelineResult {
	return &PipelineResult{
		err: err,
		id:  id,
	}
}

func NewPipelineResultOutput(id int, output interfaces.Output) *PipelineResult {
	return &PipelineResult{
		output: output,
		id:     id,
	}
}

func (p *PipelineResult) IsError() bool {
	return p.err != nil
}

func (p *PipelineResult) GetError() error {
	return p.err
}

func (p *PipelineResult) GetOutput() interfaces.Output {
	return p.output
}

func (p *PipelineResult) GetId() int {
	return p.id
}
