package videodl

import (
	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/text/linesplit" // Parse command output.

	"os/exec"
	"strings"
)

const (
	cmdName = "youtube-dl"
)

//
//--------------------------------------------------------------------[ FILE ]--

// YoutubeDL defines a video file downloader around the youtube-dl command.
//
type YoutubeDL struct {
	url     string
	formats []Format

	log cdtype.Logger
}

// NewYoutubeDL creates a video file downloader.
//
func NewYoutubeDL(log cdtype.Logger, url string) Filer {
	return &YoutubeDL{
		url: url,
		log: log,
	}
}

// Title gets the title of the video.
//
func (f *YoutubeDL) Title() (string, error) {
	return f.log.ExecSync(cmdName, "--get-filename", "-i", f.url) // -i: ignore errors.
}

// DownloadCmd downloads the video file from server.
//
func (f *YoutubeDL) DownloadCmd(path, quality string) *exec.Cmd {
	args := []string{"--all-subs", "-i", f.url}
	if quality != "" {
		args = append([]string{"-f", quality}, args...)
	}
	cmd := f.log.ExecCmd(cmdName, args...)
	cmd.Dir = path

	return cmd
}

// Formats returns the list of available streams and formats for the video.
//
func (f *YoutubeDL) Formats() ([]Format, error) {
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
		if len(args) >= 3 {
			form := Format{
				Code: args[0],
				Ext:  args[1],
				Res:  args[2],
			}

			if len(args) > 3 {
				form.Note = args[3]
			}
			if len(args) > 4 {
				last := args[len(args)-1]
				if strings.HasSuffix(last, "iB") {
					if strings.HasSuffix(last, "MiB") {
						form.Size = last[:len(last)-3]
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

	e := cmd.Run()
	return f.formats, e
}
