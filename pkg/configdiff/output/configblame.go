package output

import (
	"context"
	"fmt"
	"io"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/encoding/protojson"
)

type BlameResultOutput struct {
	*sdcpb.BlameTreeElement
}

var _ interfaces.Output = (*BlameResultOutput)(nil)

func NewBlameResultOutput(e *sdcpb.BlameTreeElement) *BlameResultOutput {
	return &BlameResultOutput{
		BlameTreeElement: e,
	}
}

func (b *BlameResultOutput) ToString(_ context.Context) (string, error) {
	return fmt.Sprintln(b.BlameTreeElement.ToString()), nil
}

func (b *BlameResultOutput) ToStringDetails(ctx context.Context) (string, error) {
	return b.ToString(ctx)
}

func (b *BlameResultOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	marshaler := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
	jsonBytes, err := marshaler.Marshal(b.BlameTreeElement)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

func (b *BlameResultOutput) ToStruct(_ context.Context) (any, error) {
	return b, nil
}
