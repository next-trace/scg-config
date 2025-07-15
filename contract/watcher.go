// Package contract defines the public interfaces and shared types used across the
// configuration system.
package contract

// Watcher watches files for changes and invokes callbacks on updates.
type Watcher interface {
	AddFile(path string, callback func()) error
	Watch(callback func())
	Close() error
}
