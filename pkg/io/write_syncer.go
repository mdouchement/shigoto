package io

import "io"

// WriteSyncer is the interface that groups the basic Write, Sync and Close methods.
type WriteSyncer interface {
	io.Writer
	io.Closer
	// Sync commits the current contents of the WriteSyncer to stable storage.
	// Typically, this means flushing the file system's in-memory copy of recently written data to disk.
	Sync() error
}
