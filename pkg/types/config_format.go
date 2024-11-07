package types

type ConfigFormat string

const (
	ConfigFormatJson     ConfigFormat = "json"
	ConfigFormatJsonIetf ConfigFormat = "json_ietf"
	ConfigFormatXml      ConfigFormat = "xml"
)
