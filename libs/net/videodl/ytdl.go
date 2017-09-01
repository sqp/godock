package videodl

import (
	"github.com/rylio/ytdl"

	"github.com/sqp/godock/libs/cdtype"

	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// YTDL implements the VideoDler interface with internal ytdl library.
//
type YTDL struct{}

// New creates a filer with the ytdl backend for the given url.
//
func (f YTDL) New(log cdtype.Logger, url string) (Filer, error) {
	return NewYTDLFile(log, url)
}

// MenuQuality returns the list of available streams and formats for the video.
//
func (f YTDL) MenuQuality() []Quality {
	return []Quality{QualityAsk, QualityBestFound}
}

type options struct {
	outputFile  string
	filters     []string
	downloadURL bool
	byteRange   string
	startOffset string
}

// YTDLFile defines a video file downloader around the ytdl library.
//
type YTDLFile struct {
	ytdl.VideoInfo

	formats []*Format
	log     cdtype.Logger
}

// NewYTDLFile creates a video file downloader.
//
func NewYTDLFile(log cdtype.Logger, url string) (Filer, error) {
	info, e := ytdl.GetVideoInfo(url)
	if e != nil {
		return nil, fmt.Errorf("Unable to fetch video info: %s", e.Error())
	}

	// formats := info.Formats
	// // parse filter arguments, and filter through formats
	// for _, filter := range options.filters {
	// 	filter, e := parseFilter(filter)
	// 	if !log.Err(e) {
	// 		formats = filter(formats)
	// 	}
	// }
	// if len(formats) == 0 {
	// 	return nil, fmt.Errorf("No formats available that match criteria: %s", e.Error())
	// }

	log.Info("Author", info.Author)
	log.Info("Description", info.Description)
	log.Info("ID", info.ID)
	log.Info("vid", info)

	var list []*Format
	for _, v := range info.Formats {
		nf := &Format{
			Itag:          v.Itag,
			Extension:     v.Extension,
			Resolution:    v.Resolution,
			VideoEncoding: v.VideoEncoding,
			AudioEncoding: v.AudioEncoding,
			AudioBitrate:  v.AudioBitrate,
		}

		s := v.ValueForKey("clen")
		if s != nil {
			nf.Size, e = strconv.Atoi(s.(string))
			nf.Size /= 1000000
			log.Err(e, "convert size. format=", v.Itag)
			// } else {
			// log.Info("no clen", v.Itag)
		}

		list = append(list, nf)
	}

	// pretty.Println(info)

	return &YTDLFile{
		VideoInfo: *info,
		formats:   list,
		log:       log,
	}, nil
}

// Title gets the title of the video.
//
func (f *YTDLFile) Title() (string, error) {
	return f.VideoInfo.Title, nil
}

// DownloadCmd downloads the video file from server.
//
func (f *YTDLFile) DownloadCmd(path string, locformat *Format, progress *Progress) func() error {
	var format ytdl.Format
	var found bool
	for _, q := range f.VideoInfo.Formats {
		if q.Itag == locformat.Itag {
			format = q
			found = true
		}
	}
	if !found {
		return func() error { return nil }
	}

	title := strings.Replace(f.VideoInfo.Title, "/", "-", -1)
	dest := filepath.Join(path, title+"."+format.Extension)

	// f.log.Info(dest, format)

	return func() error {
		downloadURL, e := f.GetDownloadURL(format)
		if e != nil {
			return fmt.Errorf("Unable to get download url: %s", e.Error())
		}

		// if options.startOffset != "" {
		// 	var offset time.Duration
		// 	offset, e = time.ParseDuration(options.startOffset)
		// 	query := downloadURL.Query()
		// 	query.Set("begin", fmt.Sprint(int64(offset/time.Millisecond)))
		// 	downloadURL.RawQuery = query.Encode()
		// }

		// if options.downloadURL {
		// 	fmt.Print(downloadURL.String())
		// 	// print new line character if outputing to terminal
		// 	if isatty.IsTerminal(os.Stdout.Fd()) {
		// 		fmt.Println()
		// 	}
		// 	return
		// }

		// if out == nil {
		// 	var fileName string
		// 	fileName, err = createFileName(options.outputFile, outputFileName{
		// 		Title:         sanitizeFileNamePart(info.Title),
		// 		Ext:           sanitizeFileNamePart(format.Extension),
		// 		DatePublished: sanitizeFileNamePart(info.DatePublished.Format("2006-01-02")),
		// 		Resolution:    sanitizeFileNamePart(format.Resolution),
		// 		Author:        sanitizeFileNamePart(info.Author),
		// 		Duration:      sanitizeFileNamePart(info.Duration.String()),
		// 	})
		// 	if err != nil {
		// 		err = fmt.Errorf("Unable to parse output file file name: %s", err.Error())
		// 		return
		// 	}
		// 	// Create file truncate if append flag is not set

		flags := os.O_CREATE | os.O_WRONLY

		// 	if options.append {
		// 		flags |= os.O_APPEND
		// 	} else {
		flags |= os.O_TRUNC
		// 	}

		// open as write only
		of, e := os.OpenFile(dest, flags, 0644)
		if e != nil {
			return fmt.Errorf("Unable to open output file: %s", e.Error())
		}
		defer of.Close()

		f.log.Info("Downloading to", of.Name()) // out.(*os.File).Name())

		req, e := http.NewRequest("GET", downloadURL.String(), nil)

		// if byte range flag is set, use http range header option
		// if options.byteRange != "" || options.append {
		// 	if options.byteRange == "" && out != os.Stdout {
		// 		if stat, err := out.(*os.File).Stat(); err != nil {
		// 			options.byteRange = strconv.FormatInt(stat.Size(), 10) + "-"
		// 		}
		// 	}
		// 	if options.byteRange != "" {
		// 		req.Header.Set("Range", "bytes="+options.byteRange)
		// 	}
		// }
		resp, e := http.DefaultClient.Do(req)
		if e != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if e == nil {
				e = fmt.Errorf("Received status code %d from download url", resp.StatusCode)
			}
			return fmt.Errorf("Unable to start download: %s", e.Error())
		}
		defer resp.Body.Close()
		// if we aren't in silent mode or the no progress flag wasn't set,
		// initialize progress bar

		// progressBar := pb.New64(resp.ContentLength)
		// progressBar.SetUnits(pb.U_BYTES)
		// progressBar.ShowTimeLeft = true
		// progressBar.ShowSpeed = true
		// //	progressBar.RefreshRate = time.Millisecond * 1
		// progressBar.Output = logOut
		// progressBar.Start()
		// defer progressBar.Finish()

		io.MultiWriter(of, progress)

		progress.SetMax(resp.ContentLength)
		_, e = io.Copy(of, resp.Body)

		f.log.Info("Finished")
		return e
	}
}

// Formats returns the list of available streams and formats for the video.
//
func (f *YTDLFile) Formats() ([]*Format, error) {
	return f.formats, nil
}

//
//
//
//
//
//
//
//
//
//
//
//

// func main() {
// 	app := cli.NewApp()
// 	app.Name = "ytdl"
// 	app.HelpName = "ytdl"
// 	// Set our own custom args usage
// 	app.ArgsUsage = "[youtube url or video id]"
// 	app.Usage = "Download youtube videos"
// 	app.HideHelp = true
// 	app.Version = "0.5.0"

// 	app.Flags = []cli.Flag{
// 		cli.HelpFlag,
// 		cli.StringFlag{
// 			Name:  "output, o",
// 			Usage: "Write output to a file, passing - outputs to stdout",
// 			Value: "{{.Title}}.{{.Ext}}",
// 		},

// 		cli.StringSliceFlag{
// 			Name:  "filter, f",
// 			Usage: "Filter available formats, syntax: [format_key]:val1,val2",
// 		},
// 		cli.StringFlag{
// 			Name:  "range, r",
// 			Usage: "Download a specific range of bytes of the video, [start]-[end]",
// 		},
// 		cli.BoolFlag{
// 			Name:  "download-url, u",
// 			Usage: "Prints download url to stdout",
// 		},

// 		cli.StringFlag{
// 			Name:  "start-offset",
// 			Usage: "Offset the start of the video by time",
// 		},
// 	}

// 	app.Action = func(c *cli.Context) {
// 		identifier := c.Args().First()
// 		if identifier == "" || c.Bool("help") {
// 			cli.ShowAppHelp(c)
// 		} else {
// 			options := options{
// 				outputFile:  c.String("output"),
// 				filters:     c.StringSlice("filter"),
// 				downloadURL: c.Bool("download-url"),
// 				byteRange:   c.String("range"),
// 				startOffset: c.String("start-offset"),
// 			}
// 			if len(options.filters) == 0 {
// 				options.filters = cli.StringSlice{
// 					fmt.Sprintf("%s:mp4", ytdl.FormatExtensionKey),
// 					fmt.Sprintf("!%s:", ytdl.FormatVideoEncodingKey),
// 					fmt.Sprintf("!%s:", ytdl.FormatAudioEncodingKey),
// 					fmt.Sprint("best"),
// 				}
// 			}
// 			handler(identifier, options)
// 		}
// 	}
// 	app.Run(os.Args)
// }

func handler(identifier string, options options, log cdtype.Logger) error {
	info, err := ytdl.GetVideoInfo(identifier)
	if err != nil {
		return fmt.Errorf("Unable to fetch video info: %s", err.Error())
	}

	formats := info.Formats
	// parse filter arguments, and filter through formats
	for _, filter := range options.filters {
		filter, err := parseFilter(filter)
		if err == nil {
			formats = filter(formats)
		}
	}
	if len(formats) == 0 {
		return fmt.Errorf("No formats available that match criteria")

	}
	format := formats[0]
	downloadURL, err := info.GetDownloadURL(format)
	if err != nil {
		return fmt.Errorf("Unable to get download url: %s", err.Error())
	}
	if options.startOffset != "" {
		var offset time.Duration
		offset, err = time.ParseDuration(options.startOffset)
		query := downloadURL.Query()
		query.Set("begin", fmt.Sprint(int64(offset/time.Millisecond)))
		downloadURL.RawQuery = query.Encode()
	}

	// log.DEV("test", )

	// pretty.Println(info)

	// if options.downloadURL {
	// 	fmt.Print(downloadURL.String())
	// 	// print new line character if outputing to terminal
	// 	if isatty.IsTerminal(os.Stdout.Fd()) {
	// 		fmt.Println()
	// 	}
	// 	return
	// }

	// if out == nil {
	// 	var fileName string
	// 	fileName, err = createFileName(options.outputFile, outputFileName{
	// 		Title:         sanitizeFileNamePart(info.Title),
	// 		Ext:           sanitizeFileNamePart(format.Extension),
	// 		DatePublished: sanitizeFileNamePart(info.DatePublished.Format("2006-01-02")),
	// 		Resolution:    sanitizeFileNamePart(format.Resolution),
	// 		Author:        sanitizeFileNamePart(info.Author),
	// 		Duration:      sanitizeFileNamePart(info.Duration.String()),
	// 	})
	// 	if err != nil {
	// 		err = fmt.Errorf("Unable to parse output file file name: %s", err.Error())
	// 		return
	// 	}
	// 	// Create file truncate if append flag is not set
	// 	flags := os.O_CREATE | os.O_WRONLY
	// 	if options.append {
	// 		flags |= os.O_APPEND
	// 	} else {
	// 		flags |= os.O_TRUNC
	// 	}
	// 	var f *os.File
	// 	// open as write only
	// 	f, err = os.OpenFile(fileName, flags, 0666)
	// 	if err != nil {
	// 		err = fmt.Errorf("Unable to open output file: %s", err.Error())
	// 		return
	// 	}
	// 	defer f.Close()
	// 	out = f
	// }

	// log.Info("Downloading to ", out.(*os.File).Name())
	// var req *http.Request
	// req, err = http.NewRequest("GET", downloadURL.String(), nil)
	// // if byte range flag is set, use http range header option
	// if options.byteRange != "" || options.append {
	// 	if options.byteRange == "" && out != os.Stdout {
	// 		if stat, err := out.(*os.File).Stat(); err != nil {
	// 			options.byteRange = strconv.FormatInt(stat.Size(), 10) + "-"
	// 		}
	// 	}
	// 	if options.byteRange != "" {
	// 		req.Header.Set("Range", "bytes="+options.byteRange)
	// 	}
	// }
	// resp, err := http.DefaultClient.Do(req)
	// if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
	// 	if err == nil {
	// 		err = fmt.Errorf("Received status code %d from download url", resp.StatusCode)
	// 	}
	// 	err = fmt.Errorf("Unable to start download: %s", err.Error())
	// 	return
	// }
	// defer resp.Body.Close()
	// // if we aren't in silent mode or the no progress flag wasn't set,
	// // initialize progress bar
	// // if !silent && !options.noProgress {
	// // 	progressBar := pb.New64(resp.ContentLength)
	// // 	progressBar.SetUnits(pb.U_BYTES)
	// // 	progressBar.ShowTimeLeft = true
	// // 	progressBar.ShowSpeed = true
	// // 	//	progressBar.RefreshRate = time.Millisecond * 1
	// // 	progressBar.Output = logOut
	// // 	progressBar.Start()
	// // 	defer progressBar.Finish()
	// // 	out = io.MultiWriter(out, progressBar)
	// // }
	// _, err = io.Copy(out, resp.Body)
	return err
}

//
//
//
//
//
//
//
//
//

//
//
//
//
//
//
//
//
//
//
//

func parseFilter(filterString string) (func(ytdl.FormatList) ytdl.FormatList, error) {

	filterString = strings.TrimSpace(filterString)
	switch filterString {
	case "best", "worst":
		return func(formats ytdl.FormatList) ytdl.FormatList {
			return formats.Extremes(ytdl.FormatResolutionKey, filterString == "best").Extremes(ytdl.FormatAudioBitrateKey, filterString == "best")
		}, nil
	case "best-video", "worst-video":
		return func(formats ytdl.FormatList) ytdl.FormatList {
			return formats.Extremes(ytdl.FormatResolutionKey, strings.HasPrefix(filterString, "best"))
		}, nil
	case "best-audio", "worst-audio":
		return func(formats ytdl.FormatList) ytdl.FormatList {
			return formats.Extremes(ytdl.FormatAudioBitrateKey, strings.HasPrefix(filterString, "best"))
		}, nil
	}
	err := fmt.Errorf("Invalid filter")
	split := strings.SplitN(filterString, ":", 2)
	if len(split) != 2 {
		return nil, err
	}
	key := ytdl.FormatKey(split[0])
	exclude := key[0] == '!'
	if exclude {
		key = key[1:]
	}
	value := strings.TrimSpace(split[1])
	if value == "best" || value == "worst" {
		return func(formats ytdl.FormatList) ytdl.FormatList {
			f := formats.Extremes(key, value == "best")
			if exclude {
				f = formats.Subtract(f)
			}
			return f
		}, nil
	}
	vals := strings.Split(value, ",")
	values := make([]interface{}, len(vals))
	for i, v := range vals {
		values[i] = strings.TrimSpace(v)
	}
	return func(formats ytdl.FormatList) ytdl.FormatList {
		f := formats.Filter(key, values)
		if exclude {
			f = formats.Subtract(f)
		}
		return f
	}, nil
}

type outputFileName struct {
	Title         string
	Author        string
	Ext           string
	DatePublished string
	Resolution    string
	Duration      string
}

var fileNameTemplate = template.New("OutputFileName")

func createFileName(template string, values outputFileName) (string, error) {
	t, err := fileNameTemplate.Parse(template)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = t.Execute(buf, values)
	if err != nil {
		return "", err
	}
	return string(buf.String()), nil
}

var illegalFileNameCharacters = regexp.MustCompile(`[^[a-zA-Z0-9]-_]`)

func sanitizeFileNamePart(part string) string {
	part = strings.Replace(part, "/", "-", -1)
	part = illegalFileNameCharacters.ReplaceAllString(part, "")
	return part
}
