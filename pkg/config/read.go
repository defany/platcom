package conf

import (
	"errors"

	gocfg "github.com/dsbasko/go-cfg"
)

var (
	ErrConfigPathEmpty = errors.New("config path is empty")
)

type fileFinder struct {
	Path string `s-flag:"cfg" flag:"config" env:"CONFIG_FILE_PATH" description:"path to config file of app"`
}

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
WithFileFinder - will try to find config path in flags passed to application and in environment variables

To pass a config file use flags like: -cfg="some very cool path" --config="some very cool path with long flag" or use CONFIG_FILE_PATH env variable
*/
func (r *Reader[T]) WithFileFinder() error {
	fr := NewReader[fileFinder]()

	ff, err := fr.Read()
	if err != nil {
		return err
	}

	if ff.Path == "" {
		return ErrConfigPathEmpty
	}

	r.WithFilePath(ff.Path)

	return nil
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
