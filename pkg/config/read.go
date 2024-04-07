package conf

import (
	"errors"
	"fmt"
	"log/slog"

	gocfg "github.com/dsbasko/go-cfg"
)

var (
	ErrConfigPathEmpty = errors.New("config path is empty")
)

type fileFinder struct {
	Path string `s-flag:"c" flag:"config" env:"CONFIG_FILE_PATH" description:"path to config file of app"`
}

type Reader[T any] struct {
	filePath string

	logger *slog.Logger
}

// NewReader creates a new structure with a type of your config.
func NewReader[T any]() *Reader[T] {
	return &Reader[T]{
		logger: slog.Default(),
	}
}

func (r *Reader[T]) WithLogger(logger *slog.Logger) *Reader[T] {
	r.logger = logger

	return r
}

// WithFilePath sets a path to your file-config
func (r *Reader[T]) WithFilePath(path string) *Reader[T] {
	r.filePath = path

	return r
}

/*
WithFileFinder - will try to find config path in flags passed to application and in environment variables

To pass a config file use flags like: -c="some very cool path to config" --config="some very cool path with long flag" or use CONFIG_FILE_PATH env variable
*/
func (r *Reader[T]) WithFileFinder() error {
	r.logger.Info("finding config file...")

	fr := NewReader[fileFinder]()

	fr.WithLogger(r.logger)

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

	r.logger.Info("reading flags...")

	if err := gocfg.ReadFlag(cfg); err != nil {
		return *cfg, err
	}

	if r.filePath != "" {
		r.logger.Info("reading config file...", slog.String("path", r.filePath))

		if err := gocfg.ReadFile(r.filePath, cfg); err != nil {
			return *cfg, err
		}
	} else {
		r.logger.Info("config file is missing, skipping...")
	}

	r.logger.Info("reading env variables...")

	if err := gocfg.ReadEnv(cfg); err != nil {
		return *cfg, err
	}

	r.logger.Info("config read successfully", slog.String("cfg_type", fmt.Sprintf("%T", *cfg)))

	return *cfg, nil
}
