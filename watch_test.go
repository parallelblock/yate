package main

import (
	"github.com/fsnotify/fsnotify"
	"testing"
	"time"
)

func callCounter() (func(), *int) {
	i := 0
	return func() {
		i++
	}, &i
}

type mockFileWatcher struct {
	addError, closeError, removeError error
	events                            chan fsnotify.Event
	errors                            chan error
	watched                           []string
	closed                            bool
}

func (m *mockFileWatcher) Close() error {
	if closed {
		panic("double closed!")
	}
	closed = true
	return m.closeError
}

func (m *mockFileWatcher) Add(path string) error {
	for _, v := range m.watched {
		if v == path {
			panic("double add")
		}
	}

	m.watched = append(m.watched, path)
	return m.addError
}

func (m *mockFileWatcher) Events() chan fsnotify.Event {
	return m.events
}

func (m *mockFileWatcher) Errors() chan error {
	return m.errors
}

func mockFW() *mockFileWatcher {
	return &mockFileWatcher{
		nil, nil, nil,
		make(chan fsnotify.Event),
		make(chan error),
		[]string{},
		false,
	}
}

func makeDelayer() (<-chan time.Time, chan<- time.Time)

func TestWatcherNoDoubleAdd(t *testing.T) {
	a, aCount := callCounter()

	watcher := mockFW()

}
