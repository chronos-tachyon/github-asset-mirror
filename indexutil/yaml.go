package indexutil

import (
	"bytes"
	"context"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

func FromYAML[T any](ctx context.Context, ptr *T, raw []byte) {
	var tmp T
	d := yaml.NewDecoder(bytes.NewReader(raw))
	d.KnownFields(true)
	err := d.Decode(&tmp)
	if err != nil {
		logger := zerolog.Ctx(ctx)
		logger.Fatal().
			Err(err).
			Msgf("failed to decode YAML as value of type %T", tmp)
		panic(nil)
	}
	*ptr = tmp
}

func ToYAML(ctx context.Context, value any) []byte {
	var result []byte
	WithBuffer(func(buf *bytes.Buffer) {
		e := yaml.NewEncoder(buf)
		e.SetIndent(2)
		err := e.Encode(value)
		if err2 := e.Close(); err == nil {
			err = err2
		}
		if err != nil {
			logger := zerolog.Ctx(ctx)
			logger.Fatal().
				Err(err).
				Msgf("failed to encode value of type %T as YAML", value)
			panic(nil)
		}
		tmp := buf.Bytes()
		result = make([]byte, len(tmp))
		copy(result, tmp)
	})
	return result
}
