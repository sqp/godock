package DiskFree

import "github.com/sqp/godock/libs/cdtype"

// Commands references.
const (
	cmdLeft = iota
	cmdMiddle
)

//------------------------------------------------------------------[ CONFIG ]--

type appletConf struct {
	cdtype.ConfGroupIconBoth `group:"Icon"`
	groupConfiguration       `group:"Configuration"`
	groupActions             `group:"Actions"`
}

type groupConfiguration struct {
	DisplayText cdtype.InfoPosition

	GaugeName string

	UpdateDelay cdtype.Duration `default:"60"`
	AutoDetect  bool
	Partitions  []string
}

type groupActions struct {
	LeftAction    int
	LeftCommand   string
	LeftClass     string
	MiddleAction  int
	MiddleCommand string
}
