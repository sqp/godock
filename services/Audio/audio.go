// Package Audio is a pulseaudio controler applet for Cairo-Dock.
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

var log cdtype.Logger

//
//------------------------------------------------------------------[ APPLET ]--

// Applet data and controlers.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf  *appletConf
	pulse *AppPulse
}

// NewApplet create a new applet instance.
//
func NewApplet() cdtype.AppInstance {
	app := &Applet{AppBase: dock.NewCDApplet()} // Icon controler and interface to cairo-dock.
	log = app.Log()
	var e error
	app.pulse, e = NewAppPulse(app)
	if log.Err(e, "pulseaudio dbus") {
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

	app.SetDefaults(cdtype.Defaults{
		Icon:  app.conf.Icon,
		Label: app.conf.Name,
		Commands: cdtype.Commands{
			"mixer": cdtype.NewCommand(app.conf.LeftAction == 1, app.conf.MixerCommand, app.conf.MixerClass)},
		Shortkeys: []cdtype.Shortkey{
			{"Actions", "MixerShortkey", "Open volume mixer", app.conf.MixerShortkey},
		},
		Debug: app.conf.Debug})

	// pulse config.
	app.pulse.StreamIcons = app.conf.StreamIcons
	app.RemoveSubIcons()

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
	log.Err(app.pulse.Init(), "pulseaudio init")
}

//
//-------------------------------------------------------------[ DOCK EVENTS ]--

// OnClick tries to launch the configured action (left click).
//
func (app *Applet) OnClick() {
	switch app.conf.LeftAction {
	case 1:
		if app.conf.MixerCommand != "" {
			app.CommandLaunch("mixer")
		}
	}
}

// OnMiddleClick tries to launch the configured action.
//
func (app *Applet) OnMiddleClick() {
	switch app.conf.MiddleAction {
	case 3: // TODO: need more actions and constants to define them.
		log.Err(app.pulse.ToggleMute())
	}
}

// OnScroll tries to launch the configured action (mouse wheel).
//
func (app *Applet) OnScroll(up bool) {
	delta := int64(ternary.Int(up, app.conf.VolumeStep, -app.conf.VolumeStep))
	e := app.pulse.SetVolumeDelta(delta)
	log.Err(e, "SetVolumeDelta")
}

// OnBuildMenu fills the menu with device actions: mute, mixer, select device (right click).
//
func (app *Applet) OnBuildMenu(menu cdtype.Menuer) {
	mute, _ := app.pulse.Device(app.pulse.sink).Bool("Mute")
	menu.AddCheckEntry("Mute volume", mute, app.pulse.ToggleMute)
	if app.conf.MixerCommand != "" {
		menu.AddEntry("Open mixer", "multimedia-volume-control", func() { app.CommandLaunch("mixer") })
	}
	app.menuAddDevices(menu, app.pulse.sink, "Managed device", app.pulse.SetSink)
}

// OnShortkey opens the mixer if found.
//
func (app *Applet) OnShortkey(string) {
	if app.conf.MixerCommand != "" {
		app.CommandLaunch("mixer")
	}
}

// OnSubMiddleClick tries to launch the configured action.
//
func (app *Applet) OnSubMiddleClick(icon string) {
	switch app.conf.MiddleAction {
	case 3: // TODO: need more actions and constants to define them.
		log.Debug("mute")
		toggleMute(app.pulse.Stream(dbus.ObjectPath(icon)))
	}
}

// OnSubScroll tries to launch the configured action (mouse wheel).
//
func (app *Applet) OnSubScroll(icon string, up bool) {
	dev := app.pulse.Stream(dbus.ObjectPath(icon))
	values, e := dev.ListUint32("Volume")
	if log.Err(e) {
		return
	}
	delta := int64(ternary.Int(up, app.conf.VolumeStep, -app.conf.VolumeStep))
	log.Err(dev.Set("Volume", VolumeDelta(values, delta)))
}

// OnSubBuildMenu fills the menu with stream actions: select device (right click).
//
func (app *Applet) OnSubBuildMenu(icon string, menu cdtype.Menuer) {
	dev := app.pulse.Stream(dbus.ObjectPath(icon))

	mute, _ := dev.Bool("Mute")
	menu.AddCheckEntry("Mute volume", mute, func() {
		toggleMute(dev)
	})

	sel, es := dev.ObjectPath("Device")
	if log.Err(es) {
		return
	}
	app.menuAddDevices(menu, sel, "Output", func(sink dbus.ObjectPath) error {
		return dev.Call("Move", 0, sink).Err
	})

	// Kill works but seem to leave the client app into a bugged state (same for stream or client kill).
	// app.menu.Append("Kill", func() {
	// 	// log.Err(dev.Call("Kill", 0).Err, "Kill") // kill stream.
	// client, ec := dev.ObjectPath("Client")
	// if ec != nil {
	// 	return
	// }
	// app.pulse.Client.Client(client).Call("Kill", 0) // kill client.
	// })
}

func (app *Applet) menuAddDevices(menu cdtype.Menuer, selected dbus.ObjectPath, title string, call func(dbus.ObjectPath) error) {
	sinks, _ := app.pulse.Core().ListPath("Sinks")
	if len(sinks) < 2 { // Only show the sinks list if we have at least 2 devices to switch between.
		return
	}
	menu.AddSeparator()
	menu.AddEntry(title, "audio-card", nil)
	menu.AddSeparator()
	for _, sink := range sinks {
		dev := app.pulse.Device(sink)
		sink := sink // make static reference of sink for the callback (we're in a range).

		v, e := dev.MapString("PropertyList")
		name := ternary.String(e == nil, v["device.description"], "unknown")
		menu.AddCheckEntry(name, sink == selected, func() { log.Err(call(sink)) })
	}
}

//
//------------------------------------------------------------[ PULSE CLIENT ]--

// AppPulse connects the pulseaudio service to the dock icon.
//
type AppPulse struct {
	pulseaudio.Client                    // Parent API. Allow direct access to control methods.
	icon              *Applet            // cdtype.RenderSimple // Dock icon renderer. To display updates on the icon.
	sink              dbus.ObjectPath    // Selected sound card.
	StreamIcons       bool               // whether we need to manage subicons for streams.
	showText          func(string) error // Volume display callback.
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
	}

	for _, e := range pulse.Register(ap) {
		log.Err(e, "register signal")
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

// SetSink sets the sink (device) to monitor.
//
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
	log.Debug("stream added", name)
	log.Err(ap.icon.AddSubIcon(name, icon, string(path)))

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

	ap.icon.SubIcon(name).SetEmblem(emblem, EmblemMuted)

	return ap.icon.SubIcon(name).SetQuickInfo(label)
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
	log.Info("NewSink", path)
	if ap.sink == "" {
		log.Info("autoselected sink, need to check.")
		ap.sink = path
	}
}

// SinkRemoved receives a lost device information.
//
func (ap *AppPulse) SinkRemoved(path dbus.ObjectPath) {
	log.Info("SinkRemoved", path)
	if ap.sink == path {
		log.Info("selected sink removed, need to check the reselect.")
		ap.sink = ""
		log.Err(ap.Init(), "SinkRemoved")
	}
}

// DeviceVolumeUpdated receives a device volume update.
//
func (ap *AppPulse) DeviceVolumeUpdated(path dbus.ObjectPath, values []uint32) {
	ap.DisplayVolume(values)
}

// DeviceMuteUpdated receives a device mute update.
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
	log.Err(ap.icon.RemoveSubIcon(string(path)))
	log.Debug("stream removed")
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
