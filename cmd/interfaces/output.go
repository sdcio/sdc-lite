package interfaces

import "io"

type Output interface {
	ToString() (string, error)
	ToStringDetails() (string, error)
	ToStruct() (any, error)
	WriteToJson(w io.Writer) error
}
