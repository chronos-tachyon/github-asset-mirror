package indexutil

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"github.com/rs/zerolog"
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

func FromJSON[T any](ctx context.Context, ptr *T, raw []byte) {
	var tmp T
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.DisallowUnknownFields()
	err := d.Decode(&tmp)
	if err == nil {
		*ptr = tmp
		return
	}

	logger := zerolog.Ctx(ctx)
	logger.Fatal().
		Err(err).
		Msgf("failed to decode JSON as value of type %T", tmp)
	panic(nil)
}

func ToJSON(ctx context.Context, value any) []byte {
	var result []byte
	WithBuffer(func(buf *bytes.Buffer) {
		e := json.NewEncoder(buf)
		e.SetEscapeHTML(false)
		e.SetIndent("", "  ")
		err := e.Encode(value)
		if err == nil {
			tmp := buf.Bytes()
			result = make([]byte, len(tmp))
			copy(result, tmp)
			return
		}

		logger := zerolog.Ctx(ctx)
		logger.Fatal().
			Err(err).
			Msgf("failed to encode value of type %T as JSON", value)
		panic(nil)
	})
	return result
}
