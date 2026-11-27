package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
	"github.com/sdcio/sdc-lite/pkg/types"
)

// RpcId represents a JSON-RPC 2.0 request/response identifier (integer, string, or null).
// Presence on the wire is tracked separately so Notifications (omitted "id") stay distinct
// from Requests carrying "id": null.
type RpcId struct {
	set bool
	raw json.RawMessage
}

func NewRpcIdInt32(id int32) RpcId {
	return RpcId{
		set: true,
		raw: json.RawMessage(strconv.FormatInt(int64(id), 10)),
	}
}

func NewRpcIdString(id string) RpcId {
	b, _ := json.Marshal(id)
	return RpcId{
		set: true,
		raw: b,
	}
}

func NullRpcId() RpcId {
	return RpcId{
		set: true,
		raw: json.RawMessage("null"),
	}
}

func (id RpcId) IsSet() bool {
	return id.set
}

func (id RpcId) Raw() json.RawMessage {
	return id.raw
}

func (id RpcId) MarshalJSON() ([]byte, error) {
	if len(id.raw) == 0 {
		return []byte("null"), nil
	}
	return id.raw, nil
}

func (id *RpcId) UnmarshalJSON(b []byte) error {
	trimmed := bytes.TrimSpace(b)
	switch {
	case bytes.Equal(trimmed, []byte("null")):
		id.set = true
		id.raw = json.RawMessage("null")
		return nil
	case len(trimmed) > 0 && trimmed[0] == '"':
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return fmt.Errorf("invalid string id: %w", err)
		}
		id.set = true
		id.raw = append(json.RawMessage(nil), trimmed...)
		return nil
	case len(trimmed) > 0 && (trimmed[0] == '-' || (trimmed[0] >= '0' && trimmed[0] <= '9')):
		if _, err := strconv.ParseInt(string(trimmed), 10, 64); err != nil {
			return fmt.Errorf("id number must be an integer: %w", err)
		}
		id.set = true
		id.raw = append(json.RawMessage(nil), trimmed...)
		return nil
	default:
		return fmt.Errorf("id must be a string, integer, or null")
	}
}

type JsonRpcMessage[T RpcRawParams] struct {
	JsonRpcEnvelope
	Params T `json:"params"`
}

func NewJsonRpcMessage[T RpcRawParams](method types.CommandType, id RpcId, params T) *JsonRpcMessage[T] {
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
	Id      RpcId             `json:"id,omitzero"`
}

type JsonRpcMessageRaw struct {
	JsonRpcEnvelope
	Params json.RawMessage `json:"params"`
}
