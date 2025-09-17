package params

import (
	"encoding/json"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type OutFormat string

var (
	OutFormatUnknown  OutFormat = ""
	OutFormatString   OutFormat = "string"
	OutFormatDetailed OutFormat = "detailed"
	OutFormatJson     OutFormat = "json"
)

func ParseOutFormat(s string) OutFormat {
	switch strings.ToLower(s) {
	case string(OutFormatString):
		return OutFormatString
	case string(OutFormatDetailed):
		return OutFormatDetailed
	case string(OutFormatJson):
		return OutFormatJson
	default:
		return OutFormatUnknown
	}
}

func (o OutFormat) String() string {
	return string(o)
}

type JsonRpcResult struct {
	id      int
	jsonrpc string
	err     error
	result  interfaces.Output
}

func NewJsonRpcResult(id int, err error, result interfaces.Output) *JsonRpcResult {
	return &JsonRpcResult{
		id:      id,
		err:     err,
		result:  result,
		jsonrpc: "2.0",
	}
}

func (j *JsonRpcResult) JsonMarshall(outFormat OutFormat) ([]byte, error) {
	var err error
	result := struct {
		Id      int    `json:"id"`
		JsonRpc string `json:"jsonrpc"`
		Error   error  `json:"error,omitempty"`
		Result  any    `json:"result,omitempty"`
	}{
		Id:      j.id,
		JsonRpc: j.jsonrpc,
		Error:   j.err,
	}

	if j.result != nil {
		switch outFormat {
		case OutFormatJson:
			result.Result, err = j.result.ToStruct()
			if err != nil {
				return nil, err
			}
		case OutFormatString:
			result.Result, err = j.result.ToString()
			if err != nil {
				return nil, err
			}
		case OutFormatDetailed:
			result.Result, err = j.result.ToStringDetails()
			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(result)
}
