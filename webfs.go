package webfs

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ecnepsnai/logtic"
	"github.com/gofrs/flock"
)

type httpHandle struct{}

var cwd string
var tmpDir string
var log *logtic.Source

// Start start the webfs server
func Start(dir, address string) error {
	cwd = dir
	log = logtic.Connect("webfs")

	t, err := ioutil.TempDir("", "webfs")
	if err != nil {
		log.Error("Error creating temporary directory: %s", err.Error())
		return err
	}
	tmpDir = t

	log.Info("Starting webfs server at '%s' for '%s'", address, dir)

	return http.ListenAndServe(address, httpHandle{})
}

func stripURL(in string) string {
	// Remove "/.." or "../" but not ".." (in case a file has two periods in the name)
	upPattern := regexp.MustCompile("(\\/\\.\\.|\\.\\.\\/)")
	out := upPattern.ReplaceAllString(in, "")
	// Remove "~" for home dir
	homePattern := regexp.MustCompile("\\~\\/?")
	out = homePattern.ReplaceAllString(out, "")
	return out
}

func pathJoin(url string) string {
	return path.Join(cwd, stripURL(url))
}

func getFileName(filePath string) string {
	components := strings.Split(filePath, "/")
	return components[len(components)-1]
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func getRealIP(r *http.Request) string {
	// If the server is behind a reverse proxy, get the real IP
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Otherwise get the ip from the host (which includes the port, so strip that)
	components := strings.Split(r.Host, ":")
	last := len(components) - 1
	host := strings.Join(components[0:last], ":")
	// Remove the [] wrap for IPv6 addresses
	host = regexp.MustCompile("[\\[\\]]").ReplaceAllString(host, "")

	return host
}

func (h httpHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileName := pathJoin(r.URL.Path)
	lock := flock.New(fileName)

	log.Info("%s %s %s", getRealIP(r), r.Method, fileName)

	// Get a file
	if r.Method == "GET" {
		// Have to check if the file exists before locking it, otherwise it will create the file (when we don't want it to)
		if !fileExists(fileName) {
			log.Warn("File '%s' does not exist", fileName)
			w.WriteHeader(404)
			return
		}

		if err := lock.RLock(); err != nil {
			log.Error("Error flocking file '%s': %s", fileName, err.Error())
			w.WriteHeader(500)
			return
		}
		defer lock.Unlock()

		file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
		if err != nil {
			log.Error("Error opening file '%s': %s", fileName, err.Error())
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
		if _, err := io.CopyBuffer(w, file, nil); err != nil {
			log.Error("Error copying buffer: %s", err.Error())
			w.WriteHeader(500)
			return
		}
		file.Close()
		return
	}

	// Create/Edit a file
	if r.Method == "PUT" || r.Method == "POST" || r.Method == "PATCH" {
		if err := lock.Lock(); err != nil {
			log.Error("Error flocking file '%s': %s", fileName, err.Error())
			w.WriteHeader(500)
			return
		}
		defer lock.Unlock()

		tmpFile := path.Join(tmpDir, getFileName(stripURL(r.URL.Path)))
		file, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Error("Error opening file '%s': %s", fileName, err.Error())
			w.WriteHeader(400)
			return
		}
		if _, err := io.CopyBuffer(file, r.Body, nil); err != nil {
			log.Error("Error copying buffer: %s", err.Error())
			w.WriteHeader(500)
			return
		}

		if err := os.Rename(tmpFile, fileName); err != nil {
			log.Error("Error renaming file '%s' to '%s': %s", tmpFile, fileName, err.Error())
			w.WriteHeader(500)
			return
		}

		file.Seek(0, io.SeekStart)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
		if _, err := io.CopyBuffer(w, file, nil); err != nil {
			log.Error("Error copying buffer: %s", err.Error())
			w.WriteHeader(500)
			return
		}
		file.Close()
	}

	// Delete a file
	if r.Method == "DELETE" {
		if err := lock.Lock(); err != nil {
			log.Error("Error flocking file '%s': %s", fileName, err.Error())
			w.WriteHeader(500)
			return
		}
		defer lock.Unlock()

		if err := os.RemoveAll(fileName); err != nil {
			log.Error("Error removing file '%s': %s", fileName, err.Error())
			w.WriteHeader(400)
			return
		}
	}
}
