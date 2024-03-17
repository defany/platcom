package swagger

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/rakyll/statik/fs"
)

type Schema struct {
	Version string `json:"swagger"`
	Info    struct {
		Title   string `json:"title"`
		Version string `json:"version"`
		Contact struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"contact"`
	} `json:"info"`
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
	Host     string   `json:"host"`
	Schemes  []string `json:"schemes"`
	Consumes []string `json:"consumes"`
	Produces []string `json:"produces"`
}

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

func (s *Serve) Middleware() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write(s.content)
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

	log.Info("opening swagger file")

	file, err := sfs.Open(s.path)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Info("reading file")

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	log.Info("unmarshalling file")

	var schema *Schema

	if err := json.Unmarshal(content, schema); err != nil {
		return err
	}

	if s.host != nil {
		log.Info("changing host in swagger", slog.String("new_base_url", *s.host), slog.String("old_base_url", schema.Host))

		schema.Host = *s.host
	}

	content, err = json.Marshal(schema)
	if err != nil {
		return err
	}

	s.content = content

	return nil
}

func ServeFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving swagger file: %s", path)

		statikFs, err := fs.New()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Open swagger file: %s", path)

		file, err := statikFs.Open(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		log.Printf("Read swagger file: %s", path)

		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Write swagger file: %s", path)

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Served swagger file: %s", path)
	}
}
