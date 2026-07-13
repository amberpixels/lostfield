package fixer_test

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/amberpixels/lostfield/internal/lf/fixer"
	"github.com/expectto/be"
	"github.com/expectto/be/be_string"
	. "github.com/onsi/gomega"
)

// newFixContext returns a minimal fixer.FixContext for inferAssignment unit tests.
func newFixContext(style fixer.OutputStyle) *fixer.FixContext {
	return &fixer.FixContext{
		InVar:       "in",
		OutVar:      "out",
		InFieldVar:  "in",
		OutputStyle: style,
	}
}

// namedTypeWithGetter builds a named struct type carrying a `Get<field>() string` method.
func namedTypeWithGetter(field string) *types.Named {
	pkg := types.NewPackage("example.com/test", "test")
	named := types.NewNamed(
		types.NewTypeName(token.NoPos, pkg, "User", nil),
		types.NewStruct(nil, nil),
		nil,
	)
	results := types.NewTuple(types.NewVar(token.NoPos, pkg, "", types.Typ[types.String]))
	sig := types.NewSignatureType(
		types.NewVar(token.NoPos, pkg, "u", named), // receiver
		nil, nil,
		nil, results, false,
	)
	named.AddMethod(types.NewFunc(token.NoPos, pkg, "Get"+field, sig))
	return named
}

func TestInferAssignment(t *testing.T) {
	strType := types.Typ[types.String]
	intType := types.Typ[types.Int]
	int64Type := types.Typ[types.Int64]

	t.Run("tier 1: identical types produce direct assignment", func(t *testing.T) {
		g := NewWithT(t)

		got := fixer.InferAssignment(newFixContext(fixer.OutputStyleDotAssignment), "Name", strType, strType)
		g.Expect(got).To(be_string.ContainingSubstring("out.Name = in.Name"))
		g.Expect(got).NotTo(be_string.ContainingSubstring("TODO"))

		got = fixer.InferAssignment(newFixContext(fixer.OutputStyleCompositeLit), "Name", strType, strType)
		g.Expect(got).To(be_string.ContainingSubstring("Name: in.Name,"))
	})

	t.Run("tier 2: getter method is preferred over conversion", func(t *testing.T) {
		g := NewWithT(t)

		fixCtx := newFixContext(fixer.OutputStyleDotAssignment)
		fixCtx.InNamedType = namedTypeWithGetter("Email")

		// int -> string is not identical; the getter tier kicks in before conversion.
		got := fixer.InferAssignment(fixCtx, "Email", intType, strType)
		g.Expect(got).To(be_string.ContainingSubstring("out.Email = in.GetEmail()"))
	})

	t.Run("tier 3: convertible types get an explicit conversion", func(t *testing.T) {
		g := NewWithT(t)

		got := fixer.InferAssignment(newFixContext(fixer.OutputStyleDotAssignment), "Count", intType, int64Type)
		g.Expect(got).To(be_string.ContainingSubstring("out.Count = int64(in.Count)"))

		got = fixer.InferAssignment(newFixContext(fixer.OutputStyleCompositeLit), "Count", intType, int64Type)
		g.Expect(got).To(be_string.ContainingSubstring("Count: int64(in.Count),"))
	})

	t.Run("tier 4: incompatible types fall back to a TODO", func(t *testing.T) {
		g := NewWithT(t)

		strSlice := types.NewSlice(strType)
		chanType := types.NewChan(types.SendRecv, strType)

		got := fixer.InferAssignment(newFixContext(fixer.OutputStyleDotAssignment), "Data", strSlice, chanType)
		g.Expect(got).To(be.All(
			be_string.ContainingSubstring("TODO(lostfield): convert in.Data"),
			be_string.ContainingSubstring("_ = in.Data"),
		))

		got = fixer.InferAssignment(newFixContext(fixer.OutputStyleCompositeLit), "Data", strSlice, chanType)
		g.Expect(got).To(be_string.ContainingSubstring("// TODO(lostfield): convert in.Data to Data"))
	})

	t.Run("nil types fall back to a verify-type TODO", func(t *testing.T) {
		g := NewWithT(t)

		got := fixer.InferAssignment(newFixContext(fixer.OutputStyleDotAssignment), "X", nil, nil)
		g.Expect(got).To(be_string.ContainingSubstring("// TODO(lostfield): verify type"))
	})
}
