# Self-describing `result` payloads inside strict JSON-RPC 2.0 responses

Pipeline responses conform strictly to JSON-RPC 2.0 at the envelope level (`jsonrpc`, `id`, and either `result` or `error`), while every `result` object carries `{ "method": "<method>", "format"?: "<format>", "data"?: <payload> }` (and `error.data` carries `{ "method": "<method>" }` when known). Standard JSON-RPC 2.0 and OpenRPC clients reject or ignore non-standard top-level response fields, yet shell pipelines, `jq` filters, and FIFO consumers need to interpret each NDJSON response line without maintaining an external `id -> request` lookup table or guessing whether `config-show` returned JSON, XML, YAML, or an XPath map.

## Considered Options

- **Top-level `"method"` field on the JSON-RPC response** — flatter, but violates the JSON-RPC 2.0 Response object specification and breaks strict OpenRPC validators/clients.
- **Synthetic `"type"` slug inside `result` (`"config-show-xml"`, etc.)** — introduces a parallel type vocabulary instead of reusing the `method` and `format` terms callers already send in `params`.
- **Bare payload in `result` with `id` correlation only** — minimal wire size, but forces every stream consumer to index sent requests and leaves empty `{}` responses (`schema-load`, `config-load`) indistinguishable.
