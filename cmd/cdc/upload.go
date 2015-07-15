package main

import "github.com/sqp/godock/libs/srvdbus"

var cmdUpload = &Command{
	Run:       runUpload,
	UsageLine: "upload fileorstring",
	Short:     "upload files or text to one-click hosting services",
	Long: `
Upload send data (raw text or file) to a one-click hosting service.

If your data start with / or file:// the file content will be sent.
Otherwise, the data is sent as raw text.

Note that you must have an active instance of the NetActivity applet.
`,
}

func runUpload(cmd *Command, args []string) {
	if len(args) == 0 { // Ensure we have some data.
		cmd.Usage()
	}
	e := srvdbus.Upload(args[0])
	logger.Err(e, "upload data")
}
