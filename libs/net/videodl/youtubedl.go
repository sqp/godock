package videodl

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/linesplit" // Parse command output.

	"os/exec"
	"strconv"
	"strings"
)

const (
	cmdName = "youtube-dl"
)

// YoutubeDL implements the VideoDler interface with external youtube-dl command.
//
type YoutubeDL struct{}

// New creates a filer with the youtube-dl backend for the given url.
//
func (f YoutubeDL) New(log cdtype.Logger, url string) (Filer, error) {
	return NewYoutubeDLFile(log, url)
}

// MenuQuality returns the list of available streams and formats for the video.
//
func (f YoutubeDL) MenuQuality() []Quality {
	return []Quality{QualityAsk, QualityBestFound, QualityBestPossible}
}

//
//--------------------------------------------------------------------[ FILE ]--

// YoutubeDLFile defines a video file downloader around the youtube-dl command.
//
type YoutubeDLFile struct {
	url     string
	formats []*Format

	log cdtype.Logger
	cmd *exec.Cmd
}

// NewYoutubeDLFile creates a video file downloader.
//
func NewYoutubeDLFile(log cdtype.Logger, url string) (Filer, error) {
	return &YoutubeDLFile{
		url: url,
		log: log,
	}, nil
}

// Title gets the title of the video.
//
func (f *YoutubeDLFile) Title() (string, error) {
	return f.log.ExecSync(cmdName, "--get-filename", "-i", f.url) // -i: ignore errors.
}

// DownloadCmd downloads the video file from server.
//
func (f *YoutubeDLFile) DownloadCmd(path string, format *Format, progress *Progress) func() error {
	args := []string{"--all-subs", "-i", f.url}
	// if quality < len(f.formats) {
	q := strconv.Itoa(format.Itag)
	args = append([]string{"-f", q}, args...)
	// }
	cmd := f.log.ExecCmd(cmdName, args...)
	cmd.Dir = path

	return f.runCmd(cmd)
}

// Formats returns the list of available streams and formats for the video.
//
func (f *YoutubeDLFile) Formats() ([]*Format, error) {
	init := false

	cmd := f.log.ExecCmd(cmdName, "-F", f.url)
	cmd.Stdout = linesplit.NewWriter(func(s string) { // results display formatter.

		if strings.HasPrefix(s, "format code") {
			init = true
			return
		}
		if !init {
			return
		}

		args := strings.Fields(s)
		code, e := strconv.Atoi(args[0])
		f.log.Err(e, "convert code ID")
		if len(args) >= 3 {
			form := &Format{
				Itag:       code,
				Extension:  args[1],
				Resolution: args[2],
			}

			if len(args) > 3 {
				// form.Note = args[3]
			}
			if len(args) > 4 {
				last := args[len(args)-1]
				if strings.HasSuffix(last, "iB") {
					if strings.HasSuffix(last, "MiB") {
						// form.Size = last[:len(last)-3]
						// f.log.Info("SIZE M", last[:len(last)-3])
					}
					args = args[:len(args)-1]
				}

				f.log.Info("more", len(args)-3, args[3:])
			}

			f.log.Info("ik", form)
			f.formats = append(f.formats, form)
		}
	})

	lastID := len(f.formats) - 1
	// ids := ""
	sel := []*Format{}
	for i := range f.formats {
		form := f.formats[lastID-i] // Reverse list order (best quality first).
		sel = append(sel, form)
		// if form.Note == "" || form.Note == "(best)" { // TODO: improve.
		// ids = strhelp.Separator(ids, ";", strhelp.Separator(form.Extension, ": ", form.Resolution))
		// }
	}
	f.formats = sel

	e := cmd.Run()
	return f.formats, e
}

func (f *YoutubeDLFile) runCmd(cmd *exec.Cmd) func() error {
	return func() error {
		f.cmd = cmd
		e := f.cmd.Run()
		f.cmd = nil
		return e
	}
}
