package src

const (
	defaultUpdateDelay = 3

	BlockSize       = 512
	KernelDiskStats = "/proc/diskstats"
)

var DockGraphType = []string{"Line", "Plain", "Bar", "Circle", "Plain Circle"}

// typedef enum _CDDisksDisplayType {
// 	CD_DISKS_GAUGE=0,
// 	CD_DISKS_GRAPH,
// 	CD_DISKS_NB_TYPES
// 	} CDDisksDispl

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
	DisplayText   int
	DisplayValues int

	GaugeName string
	// GaugeRotate bool

	GraphType int
	// GraphColorHigh []float64
	// GraphColorLow  []float64
	// GraphColorBg   []float64
	// GraphMix bool

	UpdateDelay int
	Disks       []string

	// MonitorName    string
	// MonitorEnabled bool
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
