package ember

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/maja42/ember/internal"
)

// Attachments represent embedded data in an executable.
type Attachments struct {
	exeFile *os.File
	toc internal.TOC
}

// Open returns the attachments of the running executable.
func Open() (*Attachments, error) {
	path, err := os.Executable()
	if err != nil {
		return nil, err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}
	return OpenExe(path)
}

// OpenExe returns the attachments of an arbitrary executable.
func OpenExe(exePath string) (*Attachments, error) {
	att := &Attachments{}

	exe, err := os.Open(exePath)
	if err != nil {
		return nil, err
	}
	att.exeFile = exe
	dontClose := false
	defer func() {
		if !dontClose {
			_ = exe.Close()
		}
	}()

	// determine TOC location
	tocOffset := internal.SeekBoundary(exe)
	if tocOffset < 0 { // No attachments found
		dontClose = true
		return att, nil
	}
	nextBoundary := internal.SeekBoundary(exe)
	if nextBoundary < 0 {
		// first boundary was found, but the next one (indicating the end of TOC data) is missing.
		return nil, newAttErr("corrupt attachment data (incomplete TOC)")
	}
	tocSize := int(nextBoundary) - internal.BoundarySize

	// read TOC
	if _, err := exe.Seek(tocOffset, io.SeekStart); err != nil {
		return nil, err
	}

	var jsonTOC = make([]byte, tocSize)
	if _, err := io.ReadFull(exe, jsonTOC); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonTOC, &att.toc); err != nil {
		return nil, newAttErr("corrupt attachment data (invalid TOC)")
	}

	dontClose = true
	return att, nil
}

// Close the executable containing the attachments.
// Close will return an error if it has already been called.
func (a *Attachments) Close() error {
	return a.exeFile.Close()
}

// List returns a list containing the names of all attachments.
func (a *Attachments) List() []string {
	if len(a.toc) == 0 { // no attachments
		return nil
	}
	l := make([]string, len(a.toc))
	for i, item := range a.toc {
		l[i] = item.Name
	}
	return l
}

// Count returns the number of attachments.
func (a *Attachments) Count() int {
	return len(a.toc)
}

func (a *Attachments) GetResource(name string) ([]byte, error) {
	for _, item := range a.toc {
		if item.Name == name {
			var raw bytes.Buffer
			var err error

			// Decode the data.
			in, err := base64.StdEncoding.DecodeString(item.Data)
			if err != nil {
				return nil, err
			}

			// Gunzip the data to the client
			gr, err := gzip.NewReader(bytes.NewBuffer(in))
			if err != nil {
				return nil, err
			}
			defer gr.Close()
			data, err := ioutil.ReadAll(gr)
			if err != nil {
				return nil, err
			}
			_, err = raw.Write(data)
			if err != nil {
				return nil, err
			}

			// Return it.
			return raw.Bytes(), nil
		}
	}

	return nil, newAttErr("could not found resource with name '%w'", name)
}
