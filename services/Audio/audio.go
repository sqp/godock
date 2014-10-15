// Package Audio is a pulseaudio controler applet for the Cairo-Dock project.
package Audio

import (
	"github.com/godbus/dbus"

	"github.com/sqp/pulseaudio"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/libs/dock"    // Connection to cairo-dock.
	"github.com/sqp/godock/libs/ternary" // Ternary operators.

	"errors"
	"os/exec"
	"strconv"
)

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	*dock.CDApplet
	conf  *appletConf
	pulse *AppPulse
	menu  dock.Menu
}

// NewApplet create a new applet instance.
//
func NewApplet() dock.AppletInstance {
	app := &Applet{CDApplet: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.

	var e error
	app.pulse, e = NewAppPulse(app)
	if app.Log.Err(e, "pulseaudio dbus") {
		return nil
	}
	return app
}

// Init load user configuration if needed and initialise applet.
//
func (app *Applet) Init(loadConf bool) {
	app.LoadConfig(loadConf, &app.conf) // Load config will crash if fail. Expected.

	if app.conf.MixerCommand == "" {
		app.conf.MixerCommand = findMixer()
	}

	app.SetDefaults(dock.Defaults{
		Icon:  app.conf.Icon,
		Label: ternary.String(app.conf.Name != "", app.conf.Name, app.AppletName),
		Commands: dock.Commands{
			"mixer": dock.NewCommand(app.conf.LeftAction == 1, app.conf.MixerCommand, app.conf.MixerClass)},
		Shortkeys: []string{app.conf.MixerShortkey},
		Debug:     app.conf.Debug})

	// pulse config.
	app.pulse.StreamIcons = app.conf.StreamIcons

	for icon := range app.Icons { // Remove old subicons.
		app.RemoveSubIcon(icon)
	}

	app.AddDataRenderer("", 0, "") // Remove renderer when settings changed to be sure.
	switch app.conf.DisplayValues {
	case 0:
		app.AddDataRenderer("progressbar", 1, "")
	case 1:
		app.AddDataRenderer("gauge", 1, app.conf.GaugeName)
	}

	switch app.conf.DisplayText {
	case 0:
		app.pulse.showText = func(string) error { return nil }
	case 1:
		app.pulse.showText = app.SetQuickInfo
	case 2:
		app.pulse.showText = app.SetLabel
	}

	// find sound card and display current volume, and maybe show subicons.
	app.Log.Err(app.pulse.Init(), "pulseaudio init")
}

//
//-------------------------------------------------------------[ DOCK EVENTS ]--

// OnClick tries to launch the configured action (left click).
//
func (app *Applet) OnClick() {
	switch app.conf.LeftAction {
	case 1:
		if app.conf.MixerCommand != "" {
			app.LaunchCommand("mixer")
		}
	}
}

// OnMiddleClick tries to launch the configured action.
//
func (app *Applet) OnMiddleClick() {
	switch app.conf.MiddleAction {
	case 3: // TODO: need more actions and constants to define them.
		app.Log.Err(app.pulse.ToggleMute())
	}
}

// OnScroll tries to launch the configured action (mouse wheel).
//
func (app *Applet) OnScroll(up bool) {
	delta := int64(ternary.Int(up, app.conf.VolumeStep, -app.conf.VolumeStep))
	e := app.pulse.SetVolumeDelta(delta)
	app.Log.Err(e, "SetVolumeDelta")
}

// OnBuildMenu fills the menu with device actions: mute, mixer, select device (right click).
//
func (app *Applet) OnBuildMenu() {
	app.menu.Clear()

	app.menu.Append("Mute volume", func() { app.pulse.ToggleMute() })

	if app.conf.MixerCommand != "" {
		app.menu.Append("Open mixer", func() { app.LaunchCommand("mixer") })
	}

	sinks, _ := app.pulse.Core().ListPath("Sinks")
	if len(sinks) > 1 { // Only show the sinks list if we have at least 2 devices to switch between.
		app.menu.Separator()
		for _, sink := range sinks {
			dev := app.pulse.Device(sink)
			prefix := ternary.String(sink == app.pulse.sink, "* ", "  ")

			v, e := dev.MapString("PropertyList")
			name := ternary.String(e == nil, v["device.description"], "")

			app.menuAppend(prefix+name, sink) // use a non closure func so it will make a static reference to sink (fuck range).
		}
	}
	app.PopulateMenu(app.menu.Names...)
}

func (app *Applet) menuAppend(name string, sink dbus.ObjectPath) {
	app.menu.Append(name, func() { app.pulse.SetSink(sink) })
}

// OnMenuSelect launch the selected menu action (from main or sub icons).
//
func (app *Applet) OnMenuSelect(i int32) {
	app.menu.Launch(i)
}

// OnShortkey opens the mixer if found.
//
func (app *Applet) OnShortkey(string) {
	if app.conf.MixerCommand != "" {
		app.LaunchCommand("mixer")
	}
}

// OnSubMiddleClick tries to launch the configured action.
//
func (app *Applet) OnSubMiddleClick(name string) {
	switch app.conf.MiddleAction {
	case 3: // TODO: need more actions and constants to define them.
		toggleMute(app.pulse.Stream(dbus.ObjectPath(name)))
	}
}

// OnSubScroll tries to launch the configured action (mouse wheel).
//
func (app *Applet) OnSubScroll(up bool, name string) {
	dev := app.pulse.Stream(dbus.ObjectPath(name))
	values, e := dev.ListUint32("Volume")
	if app.Log.Err(e) {
		return
	}
	delta := int64(ternary.Int(up, app.conf.VolumeStep, -app.conf.VolumeStep))
	app.Log.Err(dev.Set("Volume", VolumeDelta(values, delta)))
}

// OnSubBuildMenu fills the menu with stream actions: select device (right click).
//
func (app *Applet) OnSubBuildMenu(name string) {
	dev := app.pulse.Stream(dbus.ObjectPath(name))
	sel, es := dev.ObjectPath("Device")
	if app.Log.Err(es) {
		return
	}

	app.menu.Clear()
	sinks, _ := app.pulse.Core().ListPath("Sinks")

	if len(sinks) > 1 { // Only show the sinks list if we have at least 2 devices to switch between.
		for _, sink := range sinks {
			prefix := ternary.String(sink == sel, "* ", "  ")
			v, e := app.pulse.Device(sink).MapString("PropertyList")
			name := ternary.String(e == nil, v["device.description"], "")

			sink := sink // make local reference for the call as we are in a range.

			app.menu.Append(prefix+name, func() { app.Log.Err(dev.Call("Move", 0, sink).Err) })
		}
	}

	// Kill works but seem to leave the client app into a bugged state (same for stream or client kill).
	// app.menu.Append("Kill", func() {
	// 	// app.Log.Err(dev.Call("Kill", 0).Err, "Kill") // kill stream.
	// client, ec := dev.ObjectPath("Client")
	// if ec != nil {
	// 	return
	// }
	// app.pulse.Client.Client(client).Call("Kill", 0) // kill client.
	// })

	if len(app.menu.Names) > 0 {
		app.PopulateMenu(app.menu.Names...)
	}

}

//
//------------------------------------------------------------[ PULSE CLIENT ]--

// AppPulse connects the pulseaudio service to the dock icon.
//
type AppPulse struct {
	pulseaudio.Client               // Parent API. Allow direct access to control methods.
	icon              *Applet       // dock.RenderSimple // Dock icon renderer. To display updates on the icon.
	log               cdtype.Logger // Dock logger for the flood.

	sink        dbus.ObjectPath // Selected sound card.
	StreamIcons bool            // whether we need to manage subicons for streams.
	showText    func(string) error
}

// NewAppPulse creates a pulseaudio dbus service.
//
func NewAppPulse(obj interface{}) (*AppPulse, error) {
	pulse, e := pulseaudio.New()
	if e != nil {
		return nil, e
	}

	ap := &AppPulse{
		icon:   obj.(*Applet),
		Client: *pulse,
		log:    obj.(*Applet).Log,
	}

	for _, e := range pulse.Register(ap) {
		ap.log.Err(e, "register signal")
	}

	go pulse.Listen()

	return ap, nil
}

// Init finds the default sink to display the current volume on icon.
//
func (ap *AppPulse) Init() error {
	sink, _ := ap.Core().ObjectPath("FallbackSink") // get default sink.
	if sink == "" {
		sinks, _ := ap.Core().ListPath("Sinks")
		if len(sinks) == 0 {
			return errors.New("no sound card found")
		}
		sink = sinks[0] // then fallback to the first found.
	}

	if ap.StreamIcons {
		streams, _ := ap.Core().ListPath("PlaybackStreams")
		for _, stream := range streams {
			ap.addStream(stream)
		}
	}

	return ap.SetSink(sink)
}

func (ap *AppPulse) SetSink(sink dbus.ObjectPath) error {
	ap.sink = sink
	values, e := ap.Volume()
	if e != nil {
		return e
	}
	ap.DisplayVolume(values)
	return nil
}

// Volume returns the selected device current volume.
//
func (ap *AppPulse) Volume() ([]uint32, error) {
	if ap.sink == "" {
		return nil, errors.New("get volume: no sound card selected")
	}

	mute, e := ap.Device(ap.sink).Bool("Mute")
	if e != nil {
		return nil, e // no mute, check if need to send values??
	}
	if mute {
		return []uint32{0}, nil
	}
	values, e := ap.Device(ap.sink).ListUint32("Volume")
	if e != nil {
		return nil, e
	}
	return values, nil

}

// SetVolumeDelta changes the device volume by a relative amount.
//
func (ap *AppPulse) SetVolumeDelta(delta int64) error {
	if ap.sink == "" {
		return errors.New("set volume: no sound card selected")
	}

	dev := ap.Device(ap.sink)
	values, e := dev.ListUint32("Volume")
	if e != nil {
		return e
	}
	return dev.Set("Volume", VolumeDelta(values, delta))
}

// DisplayVolume renders the given device volume on the icon.
//
func (ap *AppPulse) DisplayVolume(values []uint32) error {
	if ap.sink == "" {
		return errors.New("get volume: no sound card selected")
	}
	mute, e := ap.Device(ap.sink).Bool("Mute")
	if e != nil {
		return e // no mute, check if need to send values??
	}

	value := VolumeToFloat(values)

	if mute {
		ap.showText(VolumeToPercent(value) + " - muted")
		return ap.icon.RenderValues(0)
	}

	ap.showText(VolumeToPercent(value))
	return ap.icon.RenderValues(value)

}

// ToggleMute changes the muted state of the selected device.
//
func (ap *AppPulse) ToggleMute() error {
	if ap.sink == "" {
		return errors.New("toggle mute: no sound card selected")
	}
	return toggleMute(ap.Device(ap.sink))
}

//
//-----------------------------------------------------------------[ STREAMS ]--

func (ap *AppPulse) addStream(path dbus.ObjectPath) {
	name, icon := ap.StreamInfo(path)
	ap.log.Err(ap.icon.AddSubIcon(name, icon, string(path)))

	values, e := ap.Stream(path).ListUint32("Volume")
	if e == nil {
		ap.DisplayStreamVolume(string(path), values)
	}
}

// DisplayStreamVolume renders the given stream volume on the subicon.
//
func (ap *AppPulse) DisplayStreamVolume(name string, values []uint32) error {
	mute, e := ap.Stream(dbus.ObjectPath(name)).Bool("Mute")
	if e != nil {
		return e // no mute, check if need to send values??
	}
	label := VolumeToPercent(VolumeToFloat(values))

	emblem := ""
	if mute {
		// label += " [M]"
		emblem = ap.icon.FileLocation("img", DefaultIconMuted)
	}

	ap.icon.Icons[name].SetEmblem(emblem, EmblemMuted)

	return ap.icon.Icons[name].SetQuickInfo(label)
}

// StreamInfo gives the name and icon of the application source of the stream.
//
func (ap *AppPulse) StreamInfo(path dbus.ObjectPath) (name, icon string) {
	client, ec := ap.Stream(path).ObjectPath("Client")
	if ec != nil {
		return "", ""
	}

	v, el := ap.Client.Client(client).MapString("PropertyList")
	if el == nil {
		name = v["application.name"]
		icon = v["application.icon_name"]
	}

	return name, icon
}

//
//---------------------------------------------------------[ PULSE CALLBACKS ]--

// NewSink receives a new device information.
//
func (ap *AppPulse) NewSink(path dbus.ObjectPath) {
	ap.log.Info("NewSink", path)
	if ap.sink == "" {
		ap.log.Info("autoselected sink, need to check.")
		ap.sink = path
	}
}

// SinkRemoved receives a lost device information.
//
func (ap *AppPulse) SinkRemoved(path dbus.ObjectPath) {
	ap.log.Info("SinkRemoved", path)
	if ap.sink == path {
		ap.log.Info("selected sink removed, need to check the reselect.")
		ap.sink = ""
		ap.log.Err(ap.Init(), "SinkRemoved")
	}
}

// DeviceVolumeUpdated receives a device volume update.
//
func (ap *AppPulse) DeviceVolumeUpdated(path dbus.ObjectPath, values []uint32) {
	ap.DisplayVolume(values)
}

// DeviceVolumeUpdated receives a device mute update.
//
func (ap *AppPulse) DeviceMuteUpdated(dbus.ObjectPath, bool) {
	values, e := ap.Device(ap.sink).ListUint32("Volume")
	if e != nil {
		return
	}
	ap.DisplayVolume(values)
}

// NewPlaybackStream receives a new stream information.
//
func (ap *AppPulse) NewPlaybackStream(path dbus.ObjectPath) {
	if ap.StreamIcons {
		ap.addStream(path)
	}
}

// PlaybackStreamRemoved receives a lost stream information.
//
func (ap *AppPulse) PlaybackStreamRemoved(path dbus.ObjectPath) {
	ap.log.Err(ap.icon.RemoveSubIcon(string(path)))
}

// StreamVolumeUpdated receives a stream volume update.
//
func (ap *AppPulse) StreamVolumeUpdated(path dbus.ObjectPath, values []uint32) {
	if ap.StreamIcons {
		ap.DisplayStreamVolume(string(path), values)
	}
}

// StreamMuteUpdated receives a stream mute update.
//
func (ap *AppPulse) StreamMuteUpdated(path dbus.ObjectPath, mute bool) {
	if !ap.StreamIcons {
		return
	}
	values, e := ap.Stream(path).ListUint32("Volume")
	if e != nil {
		return
	}
	ap.DisplayStreamVolume(string(path), values)
}

//
//-----------------------------------------------------------------[ COMMON ]--

// VolumeDelta change the volume values provided by delta percent (relative to max).
//
func VolumeDelta(values []uint32, delta int64) []uint32 {
	delta = delta * VolumeMax / 100

	for i := range values {
		newval := int64(values[i]) + delta
		if newval > VolumeMax {
			values[i] = VolumeMax
		} else if newval < 0 {
			values[i] = 0
		} else {
			values[i] = uint32(newval)
		}
	}
	return values
}

// VolumeToFloat converts Dbus volume info into an average usable value.
//
func VolumeToFloat(values []uint32) float64 {
	if len(values) == 0 {
		return 0
	}

	var val uint32
	for _, i := range values {
		val += i
	}
	val /= uint32(len(values))
	return float64(val) / VolumeMax
}

// VolumeToPercent formats the volume text (input range 0..1).
//
func VolumeToPercent(value float64) string {
	return strconv.Itoa(int(value*100)) + "%"
}

func toggleMute(dev *pulseaudio.Object) error {
	mute, e := dev.Bool("Mute")
	if e != nil {
		return e
	}
	return dev.Set("Mute", !mute)
}

func findMixer() string {
	cmd, args := findCommand(map[string]string{
		"gnome-control-center": "sound",
		"gnome-volume-control": "-p applications",
		"cinnamon-settings":    "sound",
		"pavucontrol":          "",
	})
	if cmd != "" {
		return cmd + " " + args
	}
	return ""
}

func findCommand(list map[string]string) (string, string) {
	for cmd, args := range list {
		_, e := exec.LookPath(cmd)
		if e == nil {
			return cmd, args
		}
	}
	return "", ""
}
