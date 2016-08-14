// Package appgldi implements the dock applet API for go internal applets.
//
// Its goal is to connect the main Cairo-Dock Golang applet object,
// godock/libs/cdapplet, to its parent, the dock.
package appgldi

import (
	"github.com/gotk3/gotk3/glib"

	"github.com/sqp/godock/libs/cdtype" // Applets types.
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/dialog"
	"github.com/sqp/godock/libs/ternary"

	"errors"
	"strings"
	"unsafe"

	"sync"
)

//
//----------------------------------------------------------[ API APPACTIONS ]--

// AppGldi is an applet connection to Cairo-Dock using the gldi backend.
//
type AppGldi struct {
	*IconBase // extend subIcon methods and dock Icon object to benefit from all its magic.

	mi        *gldi.ModuleInstance              // dock ModuleInstance object.
	icons     map[string]*IconBase              // SubIcons index (by ID).
	onEvent   func(string, ...interface{}) bool // Callback to dock.OnEvent to forward.
	shortkeys map[string]*gldi.Shortkey         // Shortkeys list. Index is group+key (no separator).
}

// New creates a AppGldi connection.
//
func New(mi *gldi.ModuleInstance) *AppGldi {
	return &AppGldi{
		IconBase:  &IconBase{Icon: mi.Icon()},
		mi:        mi,
		icons:     make(map[string]*IconBase),
		shortkeys: make(map[string]*gldi.Shortkey),
	}
}

// SetOnEvent sets the OnEvent callback to forwards events.
//
func (o *AppGldi) SetOnEvent(onEvent func(string, ...interface{}) bool) {
	o.onEvent = onEvent
}

//
//------------------------------------------------------------[ ICON ACTIONS ]--

// DemandsAttention is an endless Animate method. See cdtype.AppIcon.
//
func (o *AppGldi) DemandsAttention(start bool, animation string) error {
	addIdle(func() {
		switch {
		case start && gldi.ObjectIsDock(o.Icon.GetContainer()):
			o.Icon.RequestAttention(animation, 0) // endless.

		case !start && o.IsDemandingAttention():
			o.Icon.StopAttention()
		}
	})
	return nil
}

// PopupDialog opens a dialog box. See cdtype.AppIcon.
//
func (o *AppGldi) PopupDialog(data cdtype.DialogData) error {
	addIdle(func() {
		dialog.NewDialog(o.Icon, o.Icon.GetContainer(), data)
	})
	return nil
}

//
//---------------------------------------------------------------[ SHORTKEYS ]--

// BindShortkey binds any number of keyboard shortcuts to your applet. See cdtype.Shortkey.
//
func (o *AppGldi) BindShortkey(shortkeys ...cdtype.Shortkey) error {
	addIdle(func() {
		for _, sk := range shortkeys {
			if _, ok := o.shortkeys[sk.ConfGroup+sk.ConfKey]; ok { // shortkey defined, rebind.
				println("TODO: missing - rebind shortkeys")
				// 		gldi_shortkey_rebind (pKeyBinding, cShortkey, NULL);

			} else { // new shortkey.
				o.shortkeys[sk.ConfGroup+sk.ConfKey] = o.mi.NewShortkey(
					sk.ConfGroup,
					sk.ConfKey,
					sk.Desc,
					sk.Shortkey,
					func(shortkey string) { o.onEvent("on_shortkey", shortkey) },
				)
			}
		}
	})

	return nil
}

//
//-----------------------------------------------------------[ DATA RENDERER ]--

// DataRenderer manages the graphic data renderer of the icon.
//
func (o *AppGldi) DataRenderer() cdtype.IconRenderer {
	return &dataRend{icon: o}
}

// datarend implements cdtype.IconRenderer.
//
type dataRend struct {
	icon *AppGldi
}

func (o *dataRend) test(nbval int, call func() gldi.DataRendererAttributer) error {
	if nbval < 1 {
		addIdle(o.icon.RemoveDataRenderer)
	} else {
		attr := call()
		addIdle(func() { o.icon.AddNewDataRenderer(attr) })
	}
	return nil
}

func (o *dataRend) Gauge(nbval int, themeName string) error {
	return o.test(nbval, func() gldi.DataRendererAttributer {
		attr := gldi.NewDataRendererAttributeGauge()
		attr.Theme = themeName

		// SQP hack !
		attr.RotateTheme = 1

		attr.LatencyTime = 500
		attr.NbValues = int(nbval)
		return attr
	})
}

func (o *dataRend) Graph(nbval int, typ cdtype.RendererGraphType) error {
	if nbval < 1 {
		addIdle(o.icon.RemoveDataRenderer)
		return nil
	}

	attr := gldi.NewDataRendererAttributeGraph()
	attr.Type = typ
	attr.HighColor = make([]float64, nbval*3)
	attr.LowColor = make([]float64, nbval*3)
	for i := 0; i < nbval; i++ {
		attr.HighColor[3*i] = 1  // High Red.
		attr.LowColor[3*i+1] = 1 // Low  Green.
	}

	w, _ := o.icon.IconExtent()
	attr.MemorySize = ternary.Int(w > 1, w, 32)

	attr.LatencyTime = 500
	attr.NbValues = int(nbval)

	addIdle(func() {
		o.icon.AddNewDataRenderer(attr)
		// o.icon.AddDataRendererWithText(attr, o.icon.DataRendererTextPercent)
	})

	return nil
}

func (o *dataRend) Progress(nbval int) error {
	return o.test(nbval, func() gldi.DataRendererAttributer {
		attr := gldi.NewDataRendererAttributeProgressBar()

		attr.LatencyTime = 500
		attr.NbValues = int(nbval)
		return attr
	})
}

func (o *dataRend) Remove() error {
	addIdle(o.icon.RemoveDataRenderer)
	return nil
}

func (o *dataRend) Render(values ...float64) error {
	addIdle(func() {
		o.icon.Render(values...)
	})
	return nil
}

func (o *dataRend) GraphLine(nb int) error        { return o.Graph(nb, cdtype.RendererGraphLine) }
func (o *dataRend) GraphPlain(nb int) error       { return o.Graph(nb, cdtype.RendererGraphPlain) }
func (o *dataRend) GraphBar(nb int) error         { return o.Graph(nb, cdtype.RendererGraphBar) }
func (o *dataRend) GraphCircle(nb int) error      { return o.Graph(nb, cdtype.RendererGraphCircle) }
func (o *dataRend) GraphPlainCircle(nb int) error { return o.Graph(nb, cdtype.RendererGraphPlainCircle) }

//
//----------------------------------------------------------[ WINDOW ACTIONS ]--

// Window gives access to actions on the controlled window.
//
func (o *AppGldi) Window() cdtype.IconWindow { return &winAction{icon: o.Icon} }

// winAction implements cdtype.IconWindow
//
type winAction struct {
	icon *gldi.Icon
}

func (o *winAction) SetAppliClass(applicationClass string) error {
	applicationClass = strings.ToLower(applicationClass)
	class := o.icon.GetClass()
	if applicationClass == class.String() { // test if already set.
		return nil
	}

	addIdle(func() {
		if class.String() != "" {
			o.icon.DeinhibiteClass()
		}
		if applicationClass != "" {
			o.icon.InhibiteClass(applicationClass)
		}
		if !gldi.DockIsLoading() && o.icon.GetContainer() != nil {
			o.icon.Redraw()
		}
	})
	return nil
}

// act sends an action to the application controlled by the icon.
//
func (o *winAction) act(call func(*gldi.WindowActor)) error {
	if !o.icon.IsAppli() {
		return errors.New("no application")
	}
	addIdle(func() {
		call(o.icon.Window())
	})
	return nil
}

func (o *winAction) IsOpened() bool             { return o.icon.Window() != nil }
func (o *winAction) Minimize() error            { return o.act((*gldi.WindowActor).Minimize) }
func (o *winAction) Show() error                { return o.act((*gldi.WindowActor).Show) }
func (o *winAction) SetVisibility(b bool) error { return o.act(callVisibility(b)) }
func (o *winAction) ToggleVisibility() error    { return o.act((*gldi.WindowActor).ToggleVisibility) }
func (o *winAction) Maximize() error            { return o.act(winMaximize) }
func (o *winAction) Restore() error             { return o.act(winRestore) }
func (o *winAction) ToggleSize() error          { return o.act(winToggleSize) }
func (o *winAction) Close() error               { return o.act((*gldi.WindowActor).Close) }
func (o *winAction) Kill() error                { return o.act((*gldi.WindowActor).Kill) }

func winMaximize(win *gldi.WindowActor)   { win.Maximize(true) }
func winRestore(win *gldi.WindowActor)    { win.Maximize(false) }
func winToggleSize(win *gldi.WindowActor) { win.Maximize(!win.IsMaximized()) }

func callVisibility(show bool) func(*gldi.WindowActor) {
	return func(win *gldi.WindowActor) { win.SetVisibility(show) }
}

//
//---------------------------------------------------------[ SINGLE PROPERTY ]--

// IconProperty gets applet icon properties one by one.
//
func (o *AppGldi) IconProperty() cdtype.IconProperty {
	props, _ := o.IconProperties()
	return &iconProp{p: props}
}

// iconProp returns icon properties one by one, implements cdtype.IconProperty
type iconProp struct {
	p cdtype.IconProperties
}

func (o *iconProp) X() (int, error)                              { return o.p.X(), nil }
func (o *iconProp) Y() (int, error)                              { return o.p.Y(), nil }
func (o *iconProp) Width() (int, error)                          { return o.p.Width(), nil }
func (o *iconProp) Height() (int, error)                         { return o.p.Height(), nil }
func (o *iconProp) Xid() (uint64, error)                         { return o.p.Xid(), nil }
func (o *iconProp) HasFocus() (bool, error)                      { return o.p.HasFocus(), nil }
func (o *iconProp) ContainerType() (cdtype.ContainerType, error) { return o.p.ContainerType(), nil }

func (o *iconProp) ContainerPosition() (cdtype.ContainerPosition, error) {
	return o.p.ContainerPosition(), nil
}

//
//----------------------------------------------------------[ ALL PROPERTIES ]--

// IconProperties gets all applet icon properties at once.
//
func (o *AppGldi) IconProperties() (cdtype.IconProperties, error) {
	return &iconProps{icon: o.mi.Icon()}, nil
}

// iconProps returns all icon properties at once, implements cdtype.IconProperties
//
type iconProps struct {
	icon *gldi.Icon
}

func (o *iconProps) X() int {
	container := o.icon.GetContainer()
	if container.IsHorizontal() {
		return int(float64(container.WindowPositionX()) + o.icon.DrawX() + o.icon.Width()*o.icon.Scale()/2)
	}
	return int(float64(container.WindowPositionY()) + o.icon.DrawY() + o.icon.Height()*o.icon.Scale()/2)
}

func (o *iconProps) Y() int {
	container := o.icon.GetContainer()
	if container.IsHorizontal() {
		return int(float64(container.WindowPositionY()) + o.icon.DrawY() + o.icon.Height()*o.icon.Scale()/2)
	}
	return int(float64(container.WindowPositionX()) + o.icon.DrawX() + o.icon.Width()*o.icon.Scale()/2)
}

func (o *iconProps) Width() int {
	w, _ := o.icon.IconExtent()
	return w
}

func (o *iconProps) Height() int {
	_, h := o.icon.IconExtent()
	return h
}

func (o *iconProps) ContainerPosition() cdtype.ContainerPosition {
	return o.icon.GetContainer().ScreenBorder()
}

func (o *iconProps) ContainerType() cdtype.ContainerType {
	return o.icon.GetContainer().Type()
}

func (o *iconProps) Xid() uint64 {
	win := o.icon.Window()
	if win == nil {
		return 0
	}
	return uint64(uintptr(unsafe.Pointer(win))) // TODO: maybe fix
}

func (o *iconProps) HasFocus() bool {
	return o.icon.Window() != nil && o.icon.Window().IsActive()
}

//
//----------------------------------------------------------------[ SUBICONS ]--

// SubIcon returns the subicon object you can act on for the given key.
//
func (o *AppGldi) SubIcon(key string) cdtype.IconBase {
	return o.icons[key]
}

// AddSubIcon adds subicons by pack of 3 strings : label, icon, ID. See cdtype.AppIcon.
//
func (o *AppGldi) AddSubIcon(fields ...string) error {
	if len(fields)%3 != 0 {
		return errors.New("AddSubIcon: bad arguments count (must use 3 string per icon)")
	}

	list, clist := o.ModuleInstance().PrepareNewIcons(fields)
	for id, icon := range list {
		o.icons[id] = &IconBase{icon, id}
	}

	addIdle(func() {
		o.mi.InsertIcons(clist, "", "Panel")
	})

	return nil
}

// RemoveSubIcon only need the ID to remove the SubIcon. See cdtype.AppIcon.
//
func (o *AppGldi) RemoveSubIcon(id string) error {
	subicon, ok := o.icons[id]
	if !ok && id != "any" {
		return errors.New("RemoveSubIcon Icon missing: " + id)
	}

	if ok { // should allow the user to use the "any" key anyway.
		delete(o.icons, id)

		gldi.ObjectUnref(subicon.Icon)
		o.Icon.RemoveSubdockEmpty() // special hack, needs to be confirmed.
		return nil
	}

	o.icons = make(map[string]*IconBase)
	o.mi.RemoveAllIcons()
	return nil
}

// RemoveSubIcons removes all subicons from the subdock.
//
func (o *AppGldi) RemoveSubIcons() {
	for icon := range o.icons { // Remove old subicons.
		o.RemoveSubIcon(icon)
	}
}

//
//----------------------------------------------------------------[ ICONBASE ]--

// IconBase defines common actions for icons and subdock icons.
//
type IconBase struct {
	*gldi.Icon
	id string
}

// SetQuickInfo change the quickinfo text displayed on the subicon.
//
func (o *IconBase) SetQuickInfo(info string) error {
	addIdle(func() {
		o.Icon.SetQuickInfo(info)
		// o.Icon.Redraw()
	})
	return nil
}

// SetLabel change the text label next to the subicon.
//
func (o *IconBase) SetLabel(label string) error {
	addIdle(func() {
		o.Icon.SetLabel(label)
		o.Icon.Redraw()
	})
	return nil
}

// SetIcon set the image of the icon, overwriting the previous one.  See cdtype.AppIcon.
//
func (o *IconBase) SetIcon(icon string) error {
	addIdle(func() {
		o.Icon.SetIcon(icon)
		o.Icon.Redraw()
	})
	return nil
}

// SetEmblem set an emblem image on the icon. See cdtype.AppIcon.
//
func (o *IconBase) SetEmblem(iconPath string, position cdtype.EmblemPosition) error {
	addIdle(func() {
		switch {
		case iconPath == "" || iconPath == "none":
			//iPosition < CAIRO_OVERLAY_NB_POSITIONS ? iPosition : iPosition - CAIRO_OVERLAY_NB_POSITIONS // for ease of use, handle both case similarily.
			o.Icon.RemoveOverlayAtPosition(position)

		case position >= cdtype.EmblemCount: // [N; 2N-1] => print the overlay
			// cairo_dock_print_overlay_on_icon_from_image (pIcon, cImage, iPosition - CAIRO_OVERLAY_NB_POSITIONS);

		default: // [0, N-1] => add it
			o.Icon.AddOverlayFromImage(iconPath, position)
		}

		o.Icon.Redraw()
	})
	return nil
}

// Animate animates the icon for a given number of rounds.
//
func (o *IconBase) Animate(animation string, rounds int) error {
	if !gldi.ObjectIsDock(o.Icon.GetContainer()) {
		return errors.New("container is not a dock")
	}
	if animation == "" {
		return errors.New("animation text empty")
	}
	addIdle(func() {
		o.Icon.RequestAnimation(animation, int(rounds))
	})
	return nil
}

// ShowDialog pops up a simple dialog bubble on the icon. See cdtype.AppIcon.
//
func (o *IconBase) ShowDialog(message string, duration int) error {
	addIdle(func() {
		// Prevent stacking dialog messages.
		o.RemoveDialogs()

		dialog.ShowTemporaryWithIcon(message, o.Icon, o.Icon.GetContainer(), 1000*float64(duration), "same icon")
	})
	return nil
}

//
//------------------------------------------------------[ INTERNAL APPLET API ]--

// ... gives access to the underlying icon for a gldi applet.
//
// Can be expanded, there is currently no limit for a go internal applet.
//

//
//------------------------------------------------------------[ IDLE ACTIONS ]--

var idleMu = &sync.Mutex{} // protects idleDraw and idleRun.
var idleDraw []func()      // List of functions to run in the glib main loop.
var idleRun bool           // Tells if the idle flusher is running or not.

// addIdle adds a function to call on the next gtk idle cycle, to safely use
// the dock with our goroutines.
// It will also start the callIdle flush if it's not running.
//
func addIdle(call func()) {
	idleMu.Lock()
	idleDraw = append(idleDraw, call)
	if !idleRun {
		idleRun = true
		glib.IdleAdd(callIdle)
	}
	idleMu.Unlock()
}

// callIdle flushes the idleDraw list by calling them all.
// It will also process calls received while running, and stops only when
// idleDraw is really empty.
//
func callIdle() {
	for idleDraw != nil {
		idleMu.Lock()
		todraw := idleDraw
		idleDraw = nil
		idleMu.Unlock()

		for _, call := range todraw {
			call()
		}
	}

	idleMu.Lock()
	idleRun = false
	idleMu.Unlock()
}
