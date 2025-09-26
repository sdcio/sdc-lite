package types

import (
	"fmt"
	"strings"
)

type ConfigFormat string

const (
	ConfigFormatUnknown  ConfigFormat = "unknown"
	ConfigFormatJson     ConfigFormat = "json"
	ConfigFormatJsonIetf ConfigFormat = "json_ietf"
	ConfigFormatXml      ConfigFormat = "xml"
	ConfigFormatYaml     ConfigFormat = "yaml"
	ConfigFormatSdc      ConfigFormat = "sdc"
	ConfigFormatXPath    ConfigFormat = "xpath"
)

func (c ConfigFormat) String() string {
	return string(c)
}

// ConfigFormatsList List of all the known config formats
var ConfigFormatsList = ConfigFormats{ConfigFormatJson, ConfigFormatJsonIetf, ConfigFormatXml, ConfigFormatSdc, ConfigFormatYaml, ConfigFormatXPath}

func ParseConfigFormat(s string) (ConfigFormat, error) {
	for _, n := range ConfigFormatsList {
		if strings.EqualFold(string(n), s) {
			return n, nil
		}
	}
	return "", fmt.Errorf("unknown config format: %q", s)
}

type ConfigFormats []ConfigFormat

func (c ConfigFormats) StringSlice() []string {
	result := make([]string, 0, len(c))
	for _, x := range c {
		result = append(result, string(x))
	}
	return result
}
