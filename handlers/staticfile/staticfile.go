package staticfile

import (
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/aaronriekenberg/go-api/config"
)

// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
func containsDotFile(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// dotFileHidingFile is the http.File use in dotFileHidingFileSystem.
// It is used to wrap the Readdir method of http.File so that we can
// remove files and directories that start with a period from its output.
type dotFileHidingFile struct {
	http.File
}

// Readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f dotFileHidingFile) Readdir(n int) (fis []fs.FileInfo, err error) {
	files, err := f.File.Readdir(n)
	for _, file := range files { // Filters out the dot files
		if !strings.HasPrefix(file.Name(), ".") {
			fis = append(fis, file)
		}
	}
	return
}

// dotFileHidingFileSystem is an http.FileSystem that hides
// hidden "dot files" from being served.
type dotFileHidingFileSystem struct {
	http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that serves a 403 permission error when name has a file or directory
// with whose name starts with a period in its path.
func (fsys dotFileHidingFileSystem) Open(name string) (http.File, error) {
	if containsDotFile(name) { // If dot file, return 403 response
		return nil, fs.ErrPermission
	}

	file, err := fsys.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return dotFileHidingFile{file}, err
}

const CacheControl = "Cache-Control"

func StaticFileHandler(
	staticFileConfiguraton config.StaticFileConfiguration,
) http.Handler {

	fileServer := http.FileServer(
		dotFileHidingFileSystem{
			FileSystem: http.Dir(staticFileConfiguraton.RootPath),
		},
	)

	// TODO: make regexes configurable
	vnstatPNGRegex := regexp.MustCompile(`^/?vnstat/.*\.png$`)

	aaronrHostRegex := regexp.MustCompile(`^aaronr.digital|.*\.aaronr.digital$`)

	cacheControlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.Default().With(
			"urlPath", r.URL.Path,
		)

		logger.Debug("in cacheControlHandler")

		switch {
		case vnstatPNGRegex.MatchString(r.URL.Path):
			logger.Debug("vnstatPNGRegex matches")
			w.Header().Set(CacheControl, "public, max-age=60")

		case aaronrHostRegex.MatchString(r.Host):
			logger.Debug("aaronrHostRegex matches")
			w.Header().Set(CacheControl, "public, max-age=60")

		default:
			logger.Debug("default case")
			w.Header().Set(CacheControl, "public, no-cache")
		}

		fileServer.ServeHTTP(w, r)
	})

	return cacheControlHandler
}