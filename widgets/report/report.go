// Package report creates a widget with program and system informations for user reports.
//
// Use the config builder with a virtual setup.
//
package report

import (
	"github.com/shirou/gopsutil/cpu" // Get system info.
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"

	"github.com/sqp/godock/libs/cdtype"         // Logger type.
	"github.com/sqp/godock/libs/clipboard"      // Set clipboard content.
	"github.com/sqp/godock/libs/dock/confown"   // Dock templates.
	"github.com/sqp/godock/libs/gldi/desktops"  // Desktop and screens info.
	"github.com/sqp/godock/libs/net/websrv"     // Web server.
	"github.com/sqp/godock/libs/packages/build" // The config file builder.
	"github.com/sqp/godock/libs/sysinfo"        // Dock system information.
	"github.com/sqp/godock/libs/text/gtktext"   // Format text GTK.
	"github.com/sqp/godock/libs/text/tran"      // Translate.
	"github.com/sqp/godock/libs/text/versions"  // Crash counter.

	"github.com/sqp/godock/widgets/cfbuild"        // The config file builder.
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/newkey" // Create config file builder keys.
	"github.com/sqp/godock/widgets/pageswitch"     // Switcher for config pages.

	"fmt"
	"os"
	"runtime"
	"time"
)

const groupReport = "Report"

// Drop 2s of dock init flood.
const warmupDelay int64 = 2

var warmupGC int64

func init() {
	time.AfterFunc(time.Duration(warmupDelay)*time.Second, func() {
		warmupGC = runtime.NumCgoCall()
	})
}

// New creates a report widget with program and system informations.
//
func New(source cftype.Source, log cdtype.Logger, switcher *pageswitch.Switcher) cftype.Grouper {
	build := cfbuild.NewVirtual(source, log, "", "", "")
	build.BuildAll(switcher, PageReport(log))

	return build.BuildApply( // post build.
		cfbuild.TweakKeySetLabelSelectable(groupReport, "dev_report"),
	)
}

// PageReport creates a report group for the builder with program and system informations.
//
func PageReport(log cdtype.Logger) func(cftype.Builder) {
	return cfbuild.TweakAddGroup(groupReport, KeysReport(log)...)
}

// KeysReport prepares report keys for the config builder.
//
func KeysReport(log cdtype.Logger) cftype.ListKey {

	title := tran.Slate("Status")

	// Links.
	urlCharts := fmt.Sprintf("http://%s/%s", websrv.Service.URL(), websrv.PathCharts)
	urlPprof := fmt.Sprintf("http://%s/%s", websrv.Service.URL(), websrv.PathPprof)

	// Switch widgets.
	switchDebug := newkey.SwitchText(log, log.GetDebug, func(v bool) { log.SetDebug(v) }, "", "")

	monitorSwitch := newkey.SwitchText(log,
		websrv.Service.IsMonitored,
		websrv.Service.SetMonitored,
		"Monitoring web service",
		fmt.Sprintf(": %s and %s", gtktext.URI(urlCharts, "Charts"), gtktext.URI(urlPprof, "PProf")),
	)

	// Report text.
	txtReport := renderReport(log)
	callCopy := newkey.Call{Label: "Copy", Func: func() { clipboard.Write(txtReport) }}

	bc, ok := build.Current.Today()
	if ok {
		txtReport += "\n" + render(log, "report_counters", bc)
	}
	txtReport += "\n" + render(log, "report_todo", nil)

	return cftype.ListKey{
		newkey.Custom(groupReport, "debug", "Debug mode", switchDebug),
		newkey.Custom(groupReport, "monitor", " ", monitorSwitch),
		newkey.Separator(groupReport, "sep_title"),
		newkey.CustomButton(groupReport, "dev_copy", gtktext.Bold(gtktext.Big(title)), callCopy),
		newkey.TextLabel(groupReport, "dev_report", txtReport),
	}
}

func renderReport(log cdtype.Logger) string {
	data, e := getData(log)
	if e != nil {
		log.Err(e, "report get data system")
		return ""
	}
	return render(log, "report_system", data)
}

func getData(log cdtype.Logger) (map[string]interface{}, error) {

	hostInfo, e := host.Info()
	if e != nil {
		return nil, e
	}

	cpuInfo, e := cpu.Info()
	if e != nil {
		return nil, e
	}

	memInfo, e := mem.VirtualMemory()
	if e != nil {
		return nil, e
	}

	ps, e := process.NewProcess(int32(os.Getpid()))
	if e != nil {
		return nil, e
	}

	vid, e := sysinfo.NewVideoCard(log)
	if e != nil {
		return nil, e
	}

	memTotal, e := sysinfo.ProcessMemory(os.Getpid())
	if e != nil {
		return nil, e
	}

	psUptime, e := build.ProcessUptime()
	if e != nil {
		return nil, e
	}

	data := map[string]interface{}{
		"Host":          hostInfo,
		"Mem":           memInfo,
		"Process":       ps,
		"Video":         vid,
		"TotalRAM":      memTotal * 1024,
		"GOMAXPROCS":    runtime.GOMAXPROCS(0),
		"NumCPU":        runtime.NumCPU(),
		"AvgCgoCall":    uint32(runtime.NumCgoCall()-warmupGC) / uint32(int64(build.ProcessUptimeToSeconds(psUptime))-warmupDelay),
		"NumGoroutine":  runtime.NumGoroutine(),
		"SystemUptime":  time.Now().Add(-time.Duration(hostInfo.Uptime) * time.Second),
		"ProcessUptime": psUptime,
		"Screens":       desktops.Screens(),
		"NbDesktops":    desktops.NbDesktops(),
		"DesktopWidth":  desktops.WidthAll(),
		"DesktopHeight": desktops.HeightAll(),
		"ViewportX":     desktops.NbViewportX(),
		"ViewportY":     desktops.NbViewportY(),
	}

	log.Debugf("Cgo calls", "Total: %d  uptime sec: %d  avg/s: %d",
		(runtime.NumCgoCall() - warmupGC),
		int64(build.ProcessUptimeToSeconds(psUptime))-warmupDelay,
		data["AvgCgoCall"],
	)

	if len(cpuInfo) > 0 {
		data["CPU"] = cpuInfo[0]
	}

	vfields := versions.Fields()
	vfields[0].K = "Version"
	for _, field := range vfields {
		data[field.K] = field.V
	}
	return data, nil
}

func render(log cdtype.Logger, name string, data interface{}) string {
	str, e := confown.Current.TmplReport.ToString(name, data)
	log.Err(e, "report template "+name)
	return str
}
