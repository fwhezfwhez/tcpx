package tcpx

// header const key
var (
	HEADER_ROUTER_KEY   = "Router-Type"          // value ranged [MESSAGE_ID, URL_PATTERN]
	HEADER_ROUTER_VALUE = "Router-Pattern-Value" // value ranged [MESSAGE_ID, URL_PATTERN]

	HEADER_PACK_TYPE = "Pack-Content-Type" // value ranged [JSON, PROTOBUF, TOML, YAML, NONE]
)
