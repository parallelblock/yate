package main

import (
	"github.com/fsnotify/fsnotify"
	"io"
	"time"
)

type FileWatcher interface {
	io.Closer
	Add(path string) error
}

type RawFileWatcher interface {
	FileWatcher
	Remove(path string) error
	Events() chan fsnotify.Event
	Errors() chan error
}

type Action func()

type ActionedFileWatcher interface {
	io.Closer
	Create(action Action) FileWatcher
}

type TimeDelaySupplier func() <-chan time.Time

type DelayableFileWatchMgrCfg struct {
	QuietTime TimeDelaySupplier
}

type DelayableFileWatchMgr struct {
	c DelayableFileWatchMgrCfg
	p RawFileWatcher
}

func (d *DelayableFileWatchMgr) Close() error {

}

func (d *DelayableFileWatchMgr) Create(action Action) error {

}
