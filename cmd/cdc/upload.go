package main

import (
	"github.com/sqp/godock/libs/files"
	"github.com/sqp/godock/libs/srvdbus"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var cmdUpload = &Command{
	UsageLine: "upload path/to/file",
	Short:     "upload files or text to one-click hosting services",
	Long: `
Upload send data (file or pipe) to a one-click hosting service.

Flags:
  -m           Multiple files enabled (safety).

Examples:
  cdc upload /path/to/file
  cdc upload < path/to/file

  cat /path/to/file | cdc upload
  echo "test online text" | cdc upload   # send raw text

Note: you must have an active instance of the NetActivity applet.
`,
}

func init() {
	cmdUpload.Run = runUpload // break init cycle
}

var uploadMultiFiles = cmdUpload.Flag.Bool("m", false, "")

func runUpload(cmd *Command, args []string) {
	txtStdin := readStdin()

	switch {
	case len(txtStdin) > 0: // Data from stdin.
		e := srvdbus.UploadString(string(txtStdin))
		logger.Err(e, "upload data from pipe")
		return

	case len(args) == 0: // Ensure we have at least one file in args.
		cmd.Usage()

	case len(args) > 1 && !*uploadMultiFiles:
		fmt.Println("use -m to enable multiple files upload (safety option).")
		exit(1)
	}

	// Ensure files exists.
	var found []string
	for _, arg := range args {
		direct := filepath.Clean(arg)
		fullpath, e := filepath.Abs(arg)
		switch {
		case files.IsExist(direct):
			found = append(found, direct)

		case logger.Err(e, "upload find file"):

		case files.IsExist(fullpath):
			found = append(found, fullpath)

		default:
			logger.NewErr("upload file not found", arg)
		}

	}
	if len(found) > 0 {
		e := srvdbus.UploadFiles(found...)
		logger.Err(e, "upload data")
	}
}

func readStdin() []byte {
	var frompipe []byte
	stat, e := os.Stdin.Stat()
	if e == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		frompipe, e = ioutil.ReadAll(os.Stdin)
		logger.Err(e)
	}
	return frompipe
}
