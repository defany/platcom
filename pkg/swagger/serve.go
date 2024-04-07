package swagger

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/rakyll/statik/fs"
)

type Serve struct {
	log *slog.Logger

	path string

	host *string

	content []byte
}

func NewServe(path string) *Serve {
	s := &Serve{
		log:  slog.Default(),
		path: path,
	}

	return s
}

func (s *Serve) WithLogger(log *slog.Logger) *Serve {
	s.log = log

	return s
}

func (s *Serve) WithHost(host string) *Serve {
	s.host = &host

	return s
}

func (s *Serve) Middleware(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		_, err := w.Write(s.content[path])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Serve) Setup() error {
	log := s.log.With(
		slog.String("path", s.path),
	)

	sfs, err := fs.New()
	if err != nil {
		return err
	}

	log.Info("opening swagger file swagger")

	file, err := sfs.Open(s.path)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Info("reading file swagger")

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	log.Info("unmarshalling file swagger")

	var schema map[string]any

	if err := json.Unmarshal(content, &schema); err != nil {
		log.Info(err.Error())

		return err
	}

	if s.host != nil {
		log.Info("changing host in swagger file", slog.String("new_base_url", *s.host), slog.String("old_base_url", schema["host"].(string)))

		schema["host"] = *s.host
	}

	content, err = json.Marshal(schema)
	if err != nil {
		return err
	}

	log.Info("successfully setup swagger")

	s.content = content

	return nil
}
