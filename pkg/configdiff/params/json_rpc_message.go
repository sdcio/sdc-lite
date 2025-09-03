package params

import (
	"context"
	"encoding/json"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/types"
)

type JsonRpcMessage[T RpcRawParams] struct {
	JsonRpcEnvelope
	Params T `json:"params"`
}

func NewJsonRpcMessage[T RpcRawParams](method types.CommandType, id int, params T) *JsonRpcMessage[T] {
	return &JsonRpcMessage[T]{
		JsonRpcEnvelope: JsonRpcEnvelope{
			Method:  method,
			Id:      id,
			JsonRpc: "2.0",
		},
		Params: params,
	}
}

func (j *JsonRpcMessage[T]) Execute(ctx context.Context, cde Executor) (interfaces.Output, error) {
	cmd, err := j.Params.UnRaw()
	if err != nil {
		return nil, err
	}
	return cmd.Run(ctx, cde)
}

type RpcRawParams interface {
	GetMethod() types.CommandType
	UnRaw() (RunCommand, error)
}

type JsonRpcEnvelope struct {
	Method  types.CommandType `json:"method"`
	JsonRpc string            `json:"jsonrpc"`
	Id      int               `json:"id"`
}

type JsonRpcMessageRaw struct {
	JsonRpcEnvelope
	Params json.RawMessage `json:"params"`
}
