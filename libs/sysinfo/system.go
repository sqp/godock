package sysinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/sqp/godock/libs/cdtype"         // Logger type.
	"github.com/sqp/godock/libs/text/linesplit" // Parse command output.

	"strings"
)

//
//-------------------------------------------------------------------[ PRINT ]--

// VideoCard defines video card and drivers informations.
//
type VideoCard struct {
	Vendor            string
	Renderer          string
	CoreProfileOpenGL string
	VersionOpenGL     string
}

// NewVideoCard gets informations about the video card.
//
func NewVideoCard(log cdtype.Logger) (*VideoCard, error) {
	vid := &VideoCard{}
	cmd := log.ExecCmd("glxinfo")
	cmd.Stdout = linesplit.NewWriter(func(line string) {

		for ptr, match := range map[*string]string{
			&vid.Vendor:            "OpenGL vendor string: ",
			&vid.Renderer:          "OpenGL renderer string: ",
			&vid.CoreProfileOpenGL: "OpenGL core profile version string: ",
			&vid.VersionOpenGL:     "OpenGL version string: ",
		} {
			if strings.HasPrefix(line, match) {
				*ptr = line[len(match):]
			}
		}
	})

	e := cmd.Run()
	return vid, e
}

//
//-------------------------------------------------------------------[ PROCESS MEM ]--

// https://stackoverflow.com/questions/31879817/golang-os-exec-realtime-memory-usage?rq=1
// by Didier Spezia

// ProcessMemory returns the PSS (Proportional Set Size) for a given PID, expressed in KB.
// If you have just started the process, you should have the rights to access the corresponding /proc file.
//
func ProcessMemory(pid int) (uint64, error) {

	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	res := uint64(0)
	pfx := []byte("Pss:")
	r := bufio.NewScanner(f)
	for r.Scan() {
		line := r.Bytes()
		if bytes.HasPrefix(line, pfx) {
			var size uint64
			_, err := fmt.Sscanf(string(line[4:]), "%d", &size)
			if err != nil {
				return 0, err
			}
			res += size
		}
	}
	if err := r.Err(); err != nil {
		return 0, err
	}

	return res, nil
}
