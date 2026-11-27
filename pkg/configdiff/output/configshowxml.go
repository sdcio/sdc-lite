package output

import (
	"context"
	"encoding/json"
	"io"

	"github.com/beevik/etree"
	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigShowXmlOutput struct {
	tree                   TreeToXML
	onlyNewOrUpdated       bool
	honorNamespace         bool
	operationWithNamespace bool
	useOperationRemove     bool
}

var _ interfaces.Output = (*ConfigShowXmlOutput)(nil)

func NewConfigShowXmlOutput(root TreeToXML) *ConfigShowXmlOutput {
	return &ConfigShowXmlOutput{
		tree:                   root,
		onlyNewOrUpdated:       false,
		honorNamespace:         true,
		operationWithNamespace: true,
		useOperationRemove:     false,
	}
}

func (c *ConfigShowXmlOutput) SetOnlyNewOrUpdated(v bool) *ConfigShowXmlOutput {
	c.onlyNewOrUpdated = v
	return c
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

func (o *ConfigShowXmlOutput) ToString(ctx context.Context) (string, error) {
	xmlDoc, err := o.tree.ToXML(ctx, o.onlyNewOrUpdated, o.honorNamespace, o.operationWithNamespace, o.useOperationRemove)
	if err != nil {
		return "", err
	}

	wrapInConfig(xmlDoc)

	return xmlDoc.WriteToString()
}
func (o *ConfigShowXmlOutput) ToStringDetails(ctx context.Context) (string, error) {
	return o.ToString(ctx)
}

func wrapInConfig(xmlDoc *etree.Document) {
	// make sure we have a root element
	// Create a new root <config>
	root := etree.NewElement("config")

	// Move all top-level elements under <config>
	for _, el := range xmlDoc.ChildElements() {
		root.AddChild(el)
	}

	// Reset document root
	xmlDoc.SetRoot(root)
	// set indent
	xmlDoc.Indent(2)
}

func (o *ConfigShowXmlOutput) ToStruct(ctx context.Context) (any, error) {
	return o.ToString(ctx)
}
func (o *ConfigShowXmlOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jenc := json.NewEncoder(w)
	jVal, err := o.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}

type TreeToXML interface {
	ToXML(ctx context.Context, onlyNewOrUpdated bool, honorNamespace bool, operationWithNamespace bool, useOperationRemove bool) (*etree.Document, error)
}
