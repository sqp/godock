// Package Audio is a pulseaudio controler applet for Cairo-Dock.
package Audio

import (
	"github.com/godbus/dbus"

	"github.com/sqp/pulseaudio"

	"github.com/sqp/godock/libs/cdtype"  // Applet types.
	"github.com/sqp/godock/libs/ternary" // Ternary operators.

	"errors"
	"os/exec"
	"strconv"
)

var log cdtype.Logger

//
//------------------------------------------------------------------[ APPLET ]--

func init() { cdtype.Applets.Register("Audio", NewApplet) }

// Applet defines a dock applet.
//
type Applet struct {
	cdtype.AppBase // Applet base and dock connection.

	conf  *appletConf
	pulse *AppPulse
}

// NewApplet creates a new applet instance.
//
func NewApplet(base cdtype.AppBase, events *cdtype.Events) cdtype.AppInstance {
	app := &Applet{AppBase: base}
	app.SetConfig(&app.conf)

	// Events.
	events.OnClick = app.onClick
	events.OnMiddleClick = app.onMiddleClick
	events.OnScroll = app.onScroll
	events.OnBuildMenu = app.onBuildMenu
	events.OnSubMiddleClick = app.onSubMiddleClick
	events.OnSubScroll = app.onSubScroll
	events.OnSubBuildMenu = app.onSubBuildMenu

	// Pulseaudio service.
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
func (app *Applet) Init(def *cdtype.Defaults, confLoaded bool) {
	// Try to get a mixer command.
	if app.conf.MixerCommand == "" {
		app.conf.MixerCommand = findMixer()
	}

	// Defaults
	def.Commands[cmdMixer] = cdtype.NewCommand(app.conf.LeftAction == 1, app.conf.MixerCommand, app.conf.MixerClass)

	// Shortkeys.
	app.conf.MixerShortkey.Call = app.openMixer
	app.conf.ShortkeyAllMute.CallE = app.pulse.ToggleMute
	app.conf.ShortkeyAllIncrease.Call = app.globalVolumeIncrease
	app.conf.ShortkeyAllDecrease.Call = app.globalVolumeDecrease

	// Config pulse.
	app.pulse.StreamIcons = app.conf.StreamIcons
	app.RemoveSubIcons()

	// Volume renderer.
	app.DataRenderer().Remove() // Remove renderer when settings changed to be sure.
	switch app.conf.DisplayValues {
	case 0:
		app.DataRenderer().Progress(1)

	case 1:
		app.DataRenderer().Gauge(1, app.conf.GaugeName)
	}

	// Text renderer.
	switch app.conf.DisplayText {
	case 0:
		app.pulse.showText = func(string) error { return nil }
	case 1:
		app.pulse.showText = app.SetQuickInfo
	case 2:
		app.pulse.showText = app.SetLabel
	}

	// Find sound card and display current volume, and maybe show subicons.
	e := app.pulse.Init()
	app.Log().Err(e, "pulseaudio init")
}

//
//-------------------------------------------------------------[ DOCK EVENTS ]--

func (app *Applet) onClick() {
	switch app.conf.LeftAction {
	case 1:
		if app.conf.MixerCommand != "" {
			app.Command().Launch(cmdMixer)
		}
	}
}

func (app *Applet) onMiddleClick() {
	switch app.conf.MiddleAction {
	case 3: // TODO: need more actions and constants to define them.
		log.Err(app.pulse.ToggleMute())
	}
}

func (app *Applet) onScroll(up bool) {
	if up {
		app.globalVolumeIncrease()
	} else {
		app.globalVolumeDecrease()
	}
}

// onBuildMenu fills the menu with device actions: mute, mixer, select device.
//
func (app *Applet) onBuildMenu(menu cdtype.Menuer) { // device actions menu: mute, mixer, select device.
	mute, _ := app.pulse.Device(app.pulse.sink).Bool("Mute")
	menu.AddCheckEntry("Mute volume", mute, app.pulse.ToggleMute)
	if app.conf.MixerCommand != "" {
		menu.AddEntry("Open mixer", "multimedia-volume-control", app.Command().Callback(cmdMixer))
	}
	app.menuAddDevices(menu, app.pulse.sink, "Managed device", app.pulse.SetSink)
}

func (app *Applet) onSubMiddleClick(icon string) {
	switch app.conf.MiddleAction {
	case 3: // TODO: need more actions and constants to define them.
		log.Debug("mute")
		toggleMute(app.pulse.Stream(dbus.ObjectPath(icon)))
	}
}

func (app *Applet) onSubScroll(icon string, up bool) {
	dev := app.pulse.Stream(dbus.ObjectPath(icon))
	values, e := dev.ListUint32("Volume")
	if log.Err(e) {
		return
	}
	delta := app.conf.VolumeStep
	if !up {
		delta = -delta
	}
	log.Err(dev.Set("Volume", VolumeDelta(values, delta)))
}

// onSubBuildMenu fills the menu with stream actions: select device.
//
func (app *Applet) onSubBuildMenu(icon string, menu cdtype.Menuer) { // stream actions menu: select device.
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

// openMixer opens the mixer if found.
//
func (app *Applet) openMixer() {
	if app.conf.MixerCommand != "" {
		app.Command().Launch(cmdMixer)
	}
}

func (app *Applet) globalVolumeIncrease() {
	e := app.pulse.SetVolumeDelta(app.conf.VolumeStep)
	app.Log().Err(e, "SetVolumeDelta")
}

func (app *Applet) globalVolumeDecrease() {
	e := app.pulse.SetVolumeDelta(-app.conf.VolumeStep)
	app.Log().Err(e, "SetVolumeDelta")
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
	ap.icon.Log().GoTry(pulse.Listen)

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
		return ap.icon.DataRenderer().Render(0)
	}

	ap.showText(VolumeToPercent(value))
	return ap.icon.DataRenderer().Render(value)

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
