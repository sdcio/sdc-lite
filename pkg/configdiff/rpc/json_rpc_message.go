package rpc

import (
	"encoding/json"

	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
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

type RpcRawParams interface {
	GetMethod() types.CommandType
	UnRaw() (executor.RunCommand, error)
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
