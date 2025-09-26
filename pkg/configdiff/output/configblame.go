package output

import (
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

func (b *BlameResultOutput) ToString() (string, error) {
	return fmt.Sprintln(b.BlameTreeElement.ToString()), nil
}

func (b *BlameResultOutput) ToStringDetails() (string, error) {
	return b.ToString()
}

func (b *BlameResultOutput) WriteToJson(w io.Writer) error {
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

func (b *BlameResultOutput) ToStruct() (any, error) {
	return b, nil
}
