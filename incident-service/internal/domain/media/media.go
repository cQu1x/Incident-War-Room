// Package media defines the port for storing incident attachments. The service
// depends only on this abstraction; the concrete object storage adapter lives
// in the infrastructure layer.
package media

import "context"

// File is an uploaded attachment (image, video, document, …) together with the
// metadata needed to store it.
type File struct {
	Data        []byte
	ContentType string
	Ext         string
}

// Storage persists an incident attachment under the given key and returns a
// public URL pointing at it.
type Storage interface {
	Upload(ctx context.Context, key string, file File) (string, error)
}
