package conf

import (
	gocfg "github.com/dsbasko/go-cfg"
)

type Reader[T any] struct {
	filePath string
}

// NewReader creates a new structure with a type of your config.
func NewReader[T any]() *Reader[T] {
	return &Reader[T]{}
}

// WithFilePath sets a path to your file-config
func (r *Reader[T]) WithFilePath(path string) *Reader[T] {
	r.filePath = path

	return r
}

/*
Read - this method reads environment variables, flags and file by path (to pass use method Reader.WithFilePath)

# Firstly

	Reader will check flags

# Secondly

	Reader will check file if its exists

# Thirdly

	Reader will check env variables
*/
func (r *Reader[T]) Read() (T, error) {
	cfg := new(T)

	if err := gocfg.ReadFlag(cfg); err != nil {
		return *cfg, err
	}

	if r.filePath != "" {
		if err := gocfg.ReadFile(r.filePath, cfg); err != nil {
			return *cfg, err
		}
	}

	if err := gocfg.ReadEnv(cfg); err != nil {
		return *cfg, err
	}

	return *cfg, nil
}
