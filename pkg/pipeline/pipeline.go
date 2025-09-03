package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sdcio/sdc-lite/pkg/configdiff"
	"github.com/sdcio/sdc-lite/pkg/configdiff/config"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
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

		var envelope params.JsonRpcMessageRaw
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

		output, err := cmd.Run(ctx, cd)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "executing step: %d - %s\n", step, cmd.String())
		if output != nil {
			fmt.Println(output.ToStringDetails())
		}
		step++
	}

	return nil
}

func (p *Pipeline) AppendStep(s PipelineStep) error {

	var rawEntries []json.RawMessage

	// Read file
	data, err := os.ReadFile(p.filename)
	if err == nil && len(data) > 0 {
		// File exists and not empty → unmarshal
		if err := json.Unmarshal(data, &rawEntries); err != nil {
			return err
		}
	} else if err != nil && !os.IsNotExist(err) {
		// Some other error (permissions, etc.)
		return err
	}

	newRaw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	rawEntries = append(rawEntries, newRaw)

	// Marshal back
	newData, err := json.MarshalIndent(rawEntries, "", "  ")
	if err != nil {
		return err
	}

	// Save back to file
	if err := os.WriteFile(p.filename, newData, 0644); err != nil {
		return err
	}
	return nil
}

type PipelineStep interface {
	UnRaw() (params.RunCommand, error)
}
