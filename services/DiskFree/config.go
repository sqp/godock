package DiskFree

import "github.com/sqp/godock/libs/cdtype"

const (
	defaultUpdateDelay = 60 // every minute.
)

// Commands references.
const (
	cmdLeft = iota
	cmdMiddle
)

//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	groupIcon          `group:"Icon"`
	groupConfiguration `group:"Configuration"`
	groupActions       `group:"Actions"`
}

type groupIcon struct {
	Name string `conf:"name"`
}

type groupConfiguration struct {
	DisplayText cdtype.InfoPosition

	GaugeName string

	UpdateDelay int
	AutoDetect  bool
	Partitions  []string
}

type groupActions struct {
	LeftAction    int
	LeftCommand   string
	LeftClass     string
	MiddleAction  int
	MiddleCommand string

	// Still hidden.
	Debug bool
}
