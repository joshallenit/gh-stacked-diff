package testinginit

import (
	"bytes"
	"io"
	"os"
)

/*
An [io.Writer] that outputs to a string and also to Stdout.
Useful for testing so that log output can still be seen in the output of
the test if the test failed, and the output of the program can be asserted
against.
*/
type WriteRecorder struct {
	out    io.Writer
	buffer *bytes.Buffer
}

var _ io.Writer = new(WriteRecorder)

func NewWriteRecorder() *WriteRecorder {
	recorder := new(WriteRecorder)
	recorder.out = os.Stdout
	recorder.buffer = new(bytes.Buffer)
	return recorder
}

func (r *WriteRecorder) Write(p []byte) (n int, err error) {
	r.buffer.Write(p)
	return r.out.Write(p)
}

func (r *WriteRecorder) String() string {
	return r.buffer.String()
}
