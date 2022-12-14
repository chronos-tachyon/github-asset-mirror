package indexutil

import (
	"bytes"
	"sync"
)

var gPool = sync.Pool{
	New: func() any {
		buf := new(bytes.Buffer)
		buf.Grow(1 << 10) // 1 KiB
		return buf
	},
}

func WithBuffer(fn func(*bytes.Buffer)) {
	buf := gPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		gPool.Put(buf)
	}()
	fn(buf)
}
