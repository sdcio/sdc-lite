package visitors

import (
	"context"

	"github.com/sdcio/data-server/pkg/tree"
	"github.com/sdcio/data-server/pkg/utils"
	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
)

type XPathVisitor struct {
	tree.BaseVisitor
	descendMethod         tree.DescendMethod
	includeDefaults       bool
	onlyNewOrUpdated      bool
	includeExplicitDelete bool
	result                tree.LeafVariantSlice
}

func NewXPathVisitor() *XPathVisitor {
	return &XPathVisitor{
		descendMethod:         tree.DescendMethodActiveChilds,
		includeDefaults:       false,
		onlyNewOrUpdated:      false,
		includeExplicitDelete: false,
		result:                tree.LeafVariantSlice{},
	}
}

func (x *XPathVisitor) SetDescendMethod(d tree.DescendMethod) {
	x.descendMethod = d
}

func (x *XPathVisitor) SetIncludeDefaults(includeDefaults bool) {
	x.includeDefaults = includeDefaults
}

func (x *XPathVisitor) SetOnlyNewOrUpdated(onlyNewOrUpdated bool) {
	x.onlyNewOrUpdated = onlyNewOrUpdated
}

func (x *XPathVisitor) SetIncludeExplicitDelete(includeExplicitDelete bool) {
	x.includeExplicitDelete = includeExplicitDelete
}

func (x *XPathVisitor) DescendMethod() tree.DescendMethod {
	return x.descendMethod
}

func (x *XPathVisitor) Visit(ctx context.Context, e tree.Entry) error {
	x.result = e.GetHighestPrecedence(x.result, x.onlyNewOrUpdated, x.includeDefaults, x.includeExplicitDelete)
	return nil
}

func (x *XPathVisitor) GetResult() (interfaces.Output, error) {
	var err error
	result := map[string]any{}
	for _, r := range x.result {
		result[r.Path().ToXPath(false)], err = utils.GetJsonValue(r.Value(), false)
		if err != nil {
			return nil, err
		}
	}
	return output.NewConfigShowXPath(result), nil
}

var _ tree.EntryVisitor = (*XPathVisitor)(nil)
