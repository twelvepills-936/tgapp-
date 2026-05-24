package generator

// ImageResult is the outcome of an image generation request.
type ImageResult struct {
	ImageBytes []byte
	MimeType   string
	TokensUsed int64
}
