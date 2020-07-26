package main

import (
	"fmt"
	"io/ioutil"
	"testing"
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) { return 0, nil }

var emptyByte []byte = []byte{}

// A Syncer is a spy for the Sync portion of zapcore.WriteSyncer.
type Syncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *Syncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *Syncer) Called() bool {
	return s.called
}

type Discarder struct{ Syncer }

// Write implements io.Writer.
func (d *Discarder) Write(b []byte) (int, error) {
	return ioutil.Discard.Write(b)
}

func TestT(t *testing.T) {
	s := []byte("")
	fmt.Println(len(s))
}
