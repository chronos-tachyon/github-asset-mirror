package indexutil

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
)

func FromJSON[T any](ctx context.Context, ptr *T, raw []byte) {
	var tmp T
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.DisallowUnknownFields()
	err := d.Decode(&tmp)
	if err != nil {
		logger := zerolog.Ctx(ctx)
		logger.Fatal().
			Err(err).
			Msgf("failed to decode JSON as value of type %T", tmp)
		panic(nil)
	}
	*ptr = tmp
}

func ToJSON(ctx context.Context, value any) []byte {
	var result []byte
	WithBuffer(func(buf *bytes.Buffer) {
		e := json.NewEncoder(buf)
		e.SetEscapeHTML(false)
		e.SetIndent("", "  ")
		err := e.Encode(value)
		if err != nil {
			logger := zerolog.Ctx(ctx)
			logger.Fatal().
				Err(err).
				Msgf("failed to encode value of type %T as JSON", value)
			panic(nil)
		}
		tmp := buf.Bytes()
		result = make([]byte, len(tmp))
		copy(result, tmp)
	})
	return result
}
