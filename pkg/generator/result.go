package generator

// Result is the outcome of a text generation request.
type Result struct {
	Text       string
	TokensUsed int64
}
