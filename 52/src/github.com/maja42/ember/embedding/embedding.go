package embedding

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"github.com/maja42/ember/internal"
)

const compatibleVersion = "maja42/ember/v1"

// PrintlnFunc is used for logging the embedding progress.
type PrintlnFunc func(format string, args ...interface{})

// Embed embeds the attachments into the target executable.
//
// out receives the target executable including all attachments.
//
// exe reads from the target executable that should be augmented.
//
// Embed verifies that the target executable is compatible with this version of ember
// by searching for the magic marker-string (compiled into every executable that imports ember).
// Embed fails if the executable is incompatible or already contains embedded content.
//
// attachments is a map of attachment names to the corresponding readers for the content.
//
// logger (optional) is used to report the progress during embedding.
//
// Note that all ReadSeekers are seeked to their start before usage,
// meaning the entirety of readable content is embedded. Use io.SectionReader to avoid this.
func Embed(out io.Writer, exe io.ReadSeeker, attachments map[string][]byte, logger PrintlnFunc) error {
	if logger == nil {
		logger = func(string, ...interface{}) {}
	}

	if err := verifyTargetExe(exe); err != nil {
		return fmt.Errorf("verify executable: %w", err)
	}

	toc, err := buildTOC(attachments)
	if err != nil {
		return fmt.Errorf("build TOC: %w", err)
	}
	jsonTOC, err := json.Marshal(toc)
	if err != nil {
		return fmt.Errorf("marshal TOC: %w", err)
	}

	// Executable
	logger("Writing executable")
	if _, err := io.Copy(out, exe); err != nil {
		return fmt.Errorf("copy executable: %w", err)
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		return err
	}
	// TOC
	logger("Adding TOC (%d bytes)", len(jsonTOC))
	if _, err := out.Write(jsonTOC); err != nil {
		return fmt.Errorf("write TOC: %w", err)
	}
	// Boundary
	if err := internal.WriteBoundary(out); err != nil {
		return err
	}
	return nil
}

// EmbedFiles embeds the given files into the target executable.
//
// attachments is a map of attachment names to the respective file's filepath.
//
// See Embed for more information.
func EmbedFiles(out io.Writer, exe io.ReadSeeker, attachments map[string]string, logger PrintlnFunc) error {
	reader := make(map[string][]byte, len(attachments))

	for name, path := range attachments {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("open attachment %q (%q): %w", name, path, err)
		}
		reader[name] = data
	}
	return Embed(out, exe, reader, logger)
}

// verifyTargetExe ensures that the target executable is compatible.
// The reader is seeked to the beginning afterwards.
func verifyTargetExe(exe io.ReadSeeker) error {
	// Check if the target executable is compatible.
	// Compatible executables are importing 'ember' in the correct version,
	// causing a marker-string to be present in the binary.
	// String-replace is used to ensure the marker is not present in the embedder-executable.
	marker := "~~MagicMarker for XXX~~"
	marker = strings.ReplaceAll(marker, "XXX", compatibleVersion)

	return verifyCompatibility(exe, marker)
}

// buildTOC returns the TOC (table-of-contents) for embedding the given data.
// All attachments are seeked to the beginning afterwards.
func buildTOC(attachments map[string][]byte) (internal.TOC, error) {
	toc := make(internal.TOC, 0, len(attachments))

	for name, data := range attachments {
		//gzip the data.
		var b64Buffer bytes.Buffer
		b64w  := base64.NewEncoder(base64.StdEncoding, &b64Buffer)
		gw, err := gzip.NewWriterLevel(b64w, gzip.BestCompression)
		if err != nil {
			return nil, err
		}
		_, err = gw.Write(data)
		if err != nil {
			return nil, err
		}
		gw.Close()
		b64w.Close()

		toc = append(toc, internal.Attachment{
			Name: name,
			Data: b64Buffer.String(),
		})
	}
	return toc, nil
}

// ErrAlreadyEmbedded is returned if the target executable already contains attachments.
var ErrAlreadyEmbedded = errors.New("already contains embedded content")

// verifyCompatibility ensures that the target executable is compatible and not already augmented.
// This means that the target executable contains the magic-string "marker" that is compiled into the executable,
// which can be easily done by defining it in a global variable and using it in the init() function to ensure that
// it is not optimized away by the go linker. An example can be seen in maja42/ember/marker.go
//   (Note that the calling function's application should build this marker programmatically.
//    Otherwise, it will end up in the embeder's executable as well, letting it appear compatible.)
// Returns ErrAlreadyEmbedded if the target executable already contains attachments.
// The reader is seeked to the beginning afterwards.
func verifyCompatibility(exe io.ReadSeeker, marker string) error {
	// Rewind seeker to start-of-executable (just in case)
	if _, err := exe.Seek(0, io.SeekStart); err != nil {
		return err
	}

	offset := internal.SeekPattern(exe, []byte(marker))
	if offset == -1 { // not a go executable, or does not import correct library(-version) and therefore not the correct marker
		return errors.New("incompatible (magic string not found)")
	}

	offset = internal.SeekBoundary(exe)
	if offset != -1 {
		return ErrAlreadyEmbedded
	}

	if _, err := exe.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return nil
}
