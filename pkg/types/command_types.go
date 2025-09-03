package types

type CommandType string

const (
	CommandTypeUnknown        = "unknown"
	CommandTypeConfigDiff     = "config-diff"
	CommandTypeConfigShow     = "config-show"
	CommandTypeConfigLoad     = "config-load"
	CommandTypeConfigBlame    = "config-blame"
	CommandTypeConfigValidate = "config-validate"
	CommandTypeSchemaLoad     = "schema-load"
)

func (c CommandType) String() string {
	return string(c)
}

func ParseCommandType(s string) CommandType {
	switch s {
	case CommandTypeConfigBlame:
		return CommandTypeConfigDiff
	case CommandTypeConfigDiff:
		return CommandTypeConfigDiff
	case CommandTypeConfigShow:
		return CommandTypeConfigShow
	case CommandTypeConfigValidate:
		return CommandTypeConfigValidate
	case CommandTypeConfigLoad:
		return CommandTypeConfigLoad
	case CommandTypeSchemaLoad:
		return CommandTypeSchemaLoad
	}
	return CommandTypeUnknown
}
