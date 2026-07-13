package lf_test

import (
	"testing"

	"github.com/amberpixels/lostfield/internal/lf"
	"github.com/expectto/be"
	"github.com/expectto/be/be_math"
	. "github.com/onsi/gomega"
)

func TestTypeNameSimilarity(t *testing.T) {
	g := NewWithT(t)

	// Identical names (case-insensitive) are fully similar.
	g.Expect(lf.TypeNameSimilarity("User", "user")).To(be.Eq(1.0))

	// A real-world false-positive class: an API params-struct vs a result-struct
	// sharing a substring must land below the recommended 0.6 threshold.
	g.Expect(lf.TypeNameSimilarity("MessageNewParams", "Message")).To(be_math.InRange(0.4, true, 0.6, false))

	// Genuine model->DTO pairs stay above the recommended threshold.
	g.Expect(lf.TypeNameSimilarity("UserModel", "UserModelDTO")).To(be_math.Gte(0.6))
	g.Expect(lf.TypeNameSimilarity("RoleGrantDTO", "RoleGrant")).To(be_math.Gte(0.6))

	// Unrelated names score low.
	g.Expect(lf.TypeNameSimilarity("Apple", "Zebra")).To(be_math.Lt(0.2))

	// Names shorter than a bigram cannot be compared: only equality matches.
	g.Expect(lf.TypeNameSimilarity("A", "AB")).To(be.Eq(0.0))
	g.Expect(lf.TypeNameSimilarity("A", "a")).To(be.Eq(1.0))
}

func TestIsFieldTagIgnored(t *testing.T) {
	g := NewWithT(t)

	// key:"value" form: exact value must match.
	g.Expect(lf.IsFieldTagIgnored(`lostfield:"ignore"`, []string{`lostfield:"ignore"`})).To(be.True())
	g.Expect(lf.IsFieldTagIgnored(`lostfield:"keep"`, []string{`lostfield:"ignore"`})).To(be.False())

	// Unquoted value in the config entry works too.
	g.Expect(lf.IsFieldTagIgnored(`lostfield:"ignore"`, []string{`lostfield:ignore`})).To(be.True())

	// Bare key: any value matches.
	g.Expect(lf.IsFieldTagIgnored(`internal:"true"`, []string{"internal"})).To(be.True())
	g.Expect(lf.IsFieldTagIgnored(`json:"-"`, []string{`json:"-"`})).To(be.True())

	// json:"-" entry must not match a regular json name.
	g.Expect(lf.IsFieldTagIgnored(`json:"name"`, []string{`json:"-"`})).To(be.False())

	// No tag / no entries.
	g.Expect(lf.IsFieldTagIgnored("", []string{"internal"})).To(be.False())
	g.Expect(lf.IsFieldTagIgnored(`json:"name"`, nil)).To(be.False())
}

func TestFieldExclusion(t *testing.T) {
	g := NewWithT(t)

	patterns := lf.CompileFieldPatterns([]string{"CreatedAt", `^ID$`, `Meta\.Internal`})
	g.Expect(patterns).To(be.HaveLength(3))

	// Leaf-name match.
	g.Expect(lf.IsFieldExcluded("CreatedAt", "CreatedAt", patterns)).To(be.True())
	// Anchored pattern: ID excluded, RoleID kept.
	g.Expect(lf.IsFieldExcluded("ID", "ID", patterns)).To(be.True())
	g.Expect(lf.IsFieldExcluded("RoleID", "RoleID", patterns)).To(be.False())
	// Full-path match for nested fields.
	g.Expect(lf.IsFieldExcluded("Internal", "Meta.Internal", patterns)).To(be.True())
	g.Expect(lf.IsFieldExcluded("Internal", "Other.Internal", patterns)).To(be.False())

	// Invalid and empty patterns are skipped, not fatal (Config.Validate rejects
	// them upfront; this is defense in depth).
	g.Expect(lf.CompileFieldPatterns([]string{"[invalid", "", "Valid"})).To(be.HaveLength(1))
	g.Expect(lf.CompileFieldPatterns(nil)).To(be.Empty())
}
