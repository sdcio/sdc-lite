package output

import (
	"encoding/json"
	"io"

	"github.com/beevik/etree"
	"github.com/sdcio/data-server/pkg/tree"
	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigShowXmlOutput struct {
	root                   tree.Entry
	onlyNewOrUpdated       bool
	honorNamespace         bool
	operationWithNamespace bool
	useOperationRemove     bool
}

var _ interfaces.Output = (*ConfigShowXmlOutput)(nil)

func NewConfigShowXmlOutput(root tree.Entry) *ConfigShowXmlOutput {
	return &ConfigShowXmlOutput{
		root:                   root,
		onlyNewOrUpdated:       false,
		honorNamespace:         true,
		operationWithNamespace: true,
		useOperationRemove:     false,
	}
}

func (c *ConfigShowXmlOutput) SetOnlyNewOrUpdated(v bool) {
	c.onlyNewOrUpdated = v
}

func (c *ConfigShowXmlOutput) SetHonorNamespace(v bool) {
	c.honorNamespace = v
}

func (c *ConfigShowXmlOutput) SetOperationWithNamespace(v bool) {
	c.operationWithNamespace = v
}

func (c *ConfigShowXmlOutput) SetUseOperationRemove(v bool) {
	c.useOperationRemove = v
}

func (o *ConfigShowXmlOutput) ToString() (string, error) {
	xmlDoc, err := o.root.ToXML(o.onlyNewOrUpdated, o.honorNamespace, o.operationWithNamespace, o.useOperationRemove)
	if err != nil {
		return "", err
	}
	// make sure we have a root element
	// Create a new root <config>
	root := etree.NewElement("config")

	// Move all top-level elements under <config>
	for _, el := range xmlDoc.ChildElements() {
		root.AddChild(el)
	}

	// Reset document root
	xmlDoc.SetRoot(root)

	xmlDoc.Indent(2)
	return xmlDoc.WriteToString()
}
func (o *ConfigShowXmlOutput) ToStringDetails() (string, error) {
	return o.ToString()
}
func (o *ConfigShowXmlOutput) ToStruct() (any, error) {
	etre, err := o.root.ToXML(o.onlyNewOrUpdated, o.honorNamespace, o.operationWithNamespace, o.useOperationRemove)
	if err != nil {
		return nil, err
	}
	xmlString, err := etre.WriteToString()
	if err != nil {
		return nil, err
	}
	return struct{ xml string }{xml: xmlString}, nil
}
func (o *ConfigShowXmlOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)
	jVal, err := o.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
