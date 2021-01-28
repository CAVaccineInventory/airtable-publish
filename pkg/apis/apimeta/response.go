package apimeta

// List represents an API response that returns a list of objects (e.g. a list of counties).
type List struct {
	Metadata Metadata
	// Content is a list of objects with key/value pairs.
	Content interface{}
}
