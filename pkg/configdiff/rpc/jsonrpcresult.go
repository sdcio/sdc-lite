package rpc

import (
	"context"
	"encoding/json"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/types"
)

const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeServerError    = -32000
)

// RpcError is a JSON-RPC 2.0 Error object. When the failing Step's Method is known,
// Data carries {"method": "<method>"} so stream consumers can attribute errors without id lookup.
type RpcError struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    *RpcErrorData `json:"data,omitempty"`
}

type RpcErrorData struct {
	Method types.CommandType `json:"method,omitempty"`
}

// JsonRpcResult formats a Step outcome as a JSON-RPC 2.0 Response object whose inner
// result payload carries {method, format?, data?}.
type JsonRpcResult struct {
	id      RpcId
	jsonrpc string
	method  types.CommandType
	format  types.ConfigFormat
	errCode int
	err     error
	result  interfaces.Output
}

func NewJsonRpcSuccess(id RpcId, method types.CommandType, format types.ConfigFormat, result interfaces.Output) *JsonRpcResult {
	return &JsonRpcResult{
		id:      id,
		jsonrpc: "2.0",
		method:  method,
		format:  format,
		result:  result,
	}
}

func NewJsonRpcError(id RpcId, method types.CommandType, code int, err error) *JsonRpcResult {
	if code == 0 {
		code = ErrCodeServerError
	}
	return &JsonRpcResult{
		id:      id,
		jsonrpc: "2.0",
		method:  method,
		errCode: code,
		err:     err,
	}
}

type emptyOutputChecker interface {
	IsEmptyOutput() bool
}

type resultNoData struct {
	Method types.CommandType  `json:"method"`
	Format types.ConfigFormat `json:"format,omitempty"`
}

type resultWithData struct {
	Method types.CommandType  `json:"method"`
	Format types.ConfigFormat `json:"format,omitempty"`
	Data   any                `json:"data"`
}

type responseEnvelope struct {
	JsonRpc string    `json:"jsonrpc"`
	Id      RpcId     `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RpcError `json:"error,omitempty"`
}

func (j *JsonRpcResult) JsonMarshall(ctx context.Context) ([]byte, error) {
	env := responseEnvelope{
		JsonRpc: j.jsonrpc,
		Id:      j.id,
	}

	if j.err != nil {
		rpcErr := &RpcError{
			Code:    j.errCode,
			Message: j.err.Error(),
		}
		if j.method != "" && j.method != types.CommandTypeUnknown {
			rpcErr.Data = &RpcErrorData{Method: j.method}
		}
		env.Error = rpcErr
		return json.Marshal(env)
	}

	hasData := j.result != nil
	if ec, ok := j.result.(emptyOutputChecker); ok && ec.IsEmptyOutput() {
		hasData = false
	}

	if !hasData {
		env.Result = resultNoData{
			Method: j.method,
			Format: j.format,
		}
		return json.Marshal(env)
	}

	data, err := j.result.ToStruct(ctx)
	if err != nil {
		return nil, err
	}
	env.Result = resultWithData{
		Method: j.method,
		Format: j.format,
		Data:   data,
	}
	return json.Marshal(env)
}
