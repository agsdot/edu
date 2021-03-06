package watch

import (
	"time"
)

// Watcher is an interface that watches duh...
// jk lol
type Watcher interface {
	Watch() error
}

// WatcherFunc is a function that implements the
// Watcher interface.
type WatcherFunc func() error

// Watch will run the watch function
func (wf WatcherFunc) Watch() error {
	return wf()
}

// Watch is a class that holds config data
// that defines an action to be watched.
type Watch struct {
	// CRNs is a list of courses to be
	// checked by crn
	CRNs []int
	// Courses is a list of courses
	// to be checked by name
	Names []string
	// Duration is the time spent between checks
	Duration time.Duration // 12h
	Term     string
	Year     int
}
