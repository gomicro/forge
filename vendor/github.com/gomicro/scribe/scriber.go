package scribe

import "bytes"

type Scriber interface {
	BeginDescribe(desc string)
	EndDescribe()
	Print(done string)
	PrintLines(buf *bytes.Buffer)
	Error(err error)
}
