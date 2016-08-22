package DiskActivity

import "github.com/sqp/godock/libs/cdtype"

const (
	// BlockSize is the disk block size.
	BlockSize = 512
)

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
	DisplayText   int
	DisplayValues int

	GaugeName string
	// GaugeRotate bool

	GraphType cdtype.RendererGraphType
	// GraphColorHigh []float64
	// GraphColorLow  []float64
	// GraphColorBg   []float64
	// GraphMix bool

	UpdateDelay cdtype.Duration `default:"3"`
	Disks       []string
}

type groupActions struct {
	LeftAction    int
	LeftCommand   string
	LeftClass     string
	MiddleAction  int
	MiddleCommand string
}

/*
Dropped from conf as there is no way to implement it from an external applet ATM.

#l+[No;With dock orientation;Yes] Rotate applet theme :
GaugeRotate=1

#c+ High value's colour :
#{It's the colour of the graphic for high rate values.}
GraphColorHigh=1;0;0;

#c+ Low value's colour :
#{Graph colour for low rate vaues:}
GraphColorLow=1;1;0;

#C+ Background colour of the graphic :
GraphColorBg=0.61438925764858476;0.61438925764858476;0.64814221408407724;0.7686274509803922;
*/
