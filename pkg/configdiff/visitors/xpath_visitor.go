package visitors

import (
	"context"

	"github.com/sdcio/data-server/pkg/tree"
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

func (x *XPathVisitor) GetResult() string {
	// sb := &strings.Builder{}

	// for _, r := range x.result {
	// 	r.GetEntry().SdcpbPath()
	// }
	return ""
}

var _ tree.EntryVisitor = (*XPathVisitor)(nil)
