package lf

// Internals exposed for black-box tests.
var (
	TypeNameSimilarity   = typeNameSimilarity
	IsFieldTagIgnored    = isFieldTagIgnored
	IsFieldExcluded      = isFieldExcluded
	CompileFieldPatterns = compileFieldPatterns
)
