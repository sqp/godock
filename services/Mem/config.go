package Mem

import "github.com/sqp/godock/libs/cdtype"

const (
	defaultUpdateDelay = 3
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
	DisplayText   cdtype.InfoPosition
	DisplayValues int

	GaugeName string
	// GaugeRotate bool

	GraphType int
	// GraphColorHigh []float64
	// GraphColorLow  []float64
	// GraphColorBg   []float64
	// GraphMix bool

	UpdateDelay int
	ShowRAM     bool
	ShowSwap    bool
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
