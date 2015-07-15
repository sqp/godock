// Package appgldi implements the dock applet API for go internal applets.
//
// Its goal is to connect the main Cairo-Dock Golang API, godock/libs/dock, to its parent.
package appgldi

import (
	"github.com/bradfitz/iter" // iter.N.
	"github.com/conformal/gotk3/glib"

	"github.com/sqp/godock/libs/cdtype" // Applets types.
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/dialog"
	"github.com/sqp/godock/libs/ternary"

	"errors"
	"strings"
	"unsafe"
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

// SubIcon returns the subicon object you can act on for the given key.
//
func (o *AppGldi) SubIcon(key string) cdtype.IconBase {
	return o.icons[key]
}

// RemoveSubIcons removes all subicons from the subdock.
//
func (o *AppGldi) RemoveSubIcons() {
	for icon := range o.icons { // Remove old subicons.
		o.RemoveSubIcon(icon)
	}
}

// HaveMonitor gives the state of the monitored application. See cdtype.AppIcon.
//
func (o *AppGldi) HaveMonitor() (haveApp bool, haveFocus bool) {
	win := o.Icon.Window()
	if win == nil {
		return false, false
	}
	return true, win.IsActive()
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

// PopupDialog open a dialog box . See cdtype.AppIcon.
//
func (o *AppGldi) PopupDialog(data cdtype.DialogData) error {
	addIdle(func() {
		dialog.NewDialog(o.Icon, o.Icon.GetContainer(), data)
	})
	return nil
}

// AddDataRenderer add a graphic data renderer to the icon. See cdtype.AppIcon.
//
func (o *AppGldi) AddDataRenderer(typ string, nbval int32, theme string) error {
	if nbval < 1 {
		addIdle(o.Icon.RemoveDataRenderer)
		return nil
	}

	switch typ {
	case "gauge":
		attr := gldi.NewDataRendererAttributeGauge()
		attr.Theme = theme

		// SQP hack !
		attr.RotateTheme = 1

		attr.LatencyTime = 500
		attr.NbValues = int(nbval)
		addIdle(func() { o.Icon.AddNewDataRenderer(attr) })

	case "graph":
		attr := gldi.NewDataRendererAttributeGraph()
		switch theme {
		case "Line":
			attr.Type = 0
		case "Plain":
			attr.Type = 1
		case "Bar":
			attr.Type = 2
		case "Circle":
			attr.Type = 3
		case "Plain Circle":
			attr.Type = 4
		}

		attr.HighColor = make([]float64, nbval*3)
		attr.LowColor = make([]float64, nbval*3)
		for i := range iter.N(int(nbval)) {
			attr.HighColor[3*i] = 1  // High R
			attr.LowColor[3*i+1] = 1 // Low G+B = yellow.
			attr.LowColor[3*i+2] = 1
		}

		w, _ := o.Icon.IconExtent()
		attr.MemorySize = ternary.Int(w > 1, w, 32)

		attr.LatencyTime = 500
		attr.NbValues = int(nbval)
		addIdle(func() { o.Icon.AddNewDataRenderer(attr) })

	case "progressbar":
		attr := gldi.NewDataRendererAttributeProgressBar()

		attr.LatencyTime = 500
		attr.NbValues = int(nbval)
		addIdle(func() { o.Icon.AddNewDataRenderer(attr) })

	default: // Failed to provide a valid renderer. Removing old if any.
		addIdle(o.Icon.RemoveDataRenderer)
	}

	return nil
}

// RenderValues render new values on the icon. See cdtype.AppIcon.
//
func (o *AppGldi) RenderValues(values ...float64) error {
	addIdle(func() {
		o.Icon.RenderNewData(values...)
		o.Icon.Redraw()
	})
	return nil
}

// ActOnAppli send an action on the controlled application. See cdtype.AppIcon.
//
func (o *AppGldi) ActOnAppli(action string) error {
	switch action {
	case "minimize":
		o.Icon.Window().Minimize()

	case "show":
		o.Icon.Window().Show()

	case "toggle-visibility":
		if o.Icon.Window().IsHidden() {
			o.Icon.Window().Show()
		} else {
			o.Icon.Window().Minimize()
		}

	case "maximize":
		o.Icon.Window().Maximize(true)

	case "restore":
		o.Icon.Window().Maximize(false)

	case "toggle-size":
		o.Icon.Window().Maximize(!o.Icon.Window().IsMaximized())

	case "close":
		o.Icon.Window().Close()

	case "kill":
		o.Icon.Window().Kill()

	default:
		return errors.New("ActOnAppli: invalid action=" + action)
	}

	return nil
}

// ControlAppli allow your applet to control a window.  See cdtype.AppIcon.
//
func (o *AppGldi) ControlAppli(applicationClass string) error {
	applicationClass = strings.ToLower(applicationClass)
	class := o.Icon.GetClass()
	if applicationClass == class.String() { // test if already set.
		return nil
	}

	addIdle(func() {
		if class.String() != "" {
			o.Icon.DeinhibiteClass()
		}
		if applicationClass != "" {
			o.Icon.InhibiteClass(applicationClass)
		}
		if !gldi.DockIsLoading() && o.Icon.GetContainer() != nil {
			o.Icon.Redraw()
		}
	})
	return nil
}

// ShowAppli set the visible state of the controlled application. See cdtype.AppIcon.
//
func (o *AppGldi) ShowAppli(show bool) error {
	addIdle(func() {
		if show {
			o.Icon.Window().Show()
		} else {
			o.Icon.Window().Minimize()
		}
	})
	return nil
}

// BindShortkey binds any number of keyboard shortcuts to your applet. See cdtype.Shortkey.
//
func (o *AppGldi) BindShortkey(shortkeys ...cdtype.Shortkey) error {
	addIdle(func() {
		for _, sk := range shortkeys {
			if _, ok := o.shortkeys[sk.ConfGroup+sk.ConfKey]; ok { // shortkey defined, rebind.
				println("TODO: missing - rebind shortkeys")
				// 		gldi_shortkey_rebind (pKeyBinding, cShortkey, NULL);

			} else { // new shortkey.
				o.shortkeys[sk.ConfGroup+sk.ConfKey] = o.mi.NewShortkey(sk.ConfGroup, sk.ConfKey, sk.Desc, sk.Shortkey, o.onShortkey)
			}
		}
	})

	return nil
}

// onShortkey is the shortkey callback, to forward events.
//
func (o *AppGldi) onShortkey(shortkey string) {
	o.onEvent("on_shortkey", shortkey)
}

// Get a property of the icon of your applet. Current available properties are :
//   x            int32     x position of the icon's center on the screen (starting from 0 on the left)
//   y            int32     y position of the icon's center on the screen (starting from 0 at the top of the screen)
//   width        int32     width of the icon, in pixels (this is the maximum width, when the icon is zoomed)
//   height       int32     height of the icon, in pixels (this is the maximum height, when the icon is zoomed)
//   container    uint32   type of container of the applet (DOCK, DESKLET)
//   orientation  uint32   position of the container on the screen (BOTTOM, TOP, RIGHT, LEFT). A desklet has always an orientation of BOTTOM.
//   Xid          uint64   ID of the application's window which is controlled by the applet, or 0 if none (this parameter can only be non nul if you used the method ControlAppli beforehand).
//   has_focus    bool     Whether the application's window which is controlled by the applet is the current active window (it has the focus) or not. E.g.:
//
func (o *AppGldi) Get(property string) (interface{}, error) {
	container := o.Icon.GetContainer()
	switch property {
	case "x":
		if container.IsHorizontal() {
			return float64(container.WindowPositionX()) + o.Icon.DrawX() + o.Icon.Width()*o.Icon.Scale()/2, nil
		}
		return float64(container.WindowPositionY()) + o.Icon.DrawY() + o.Icon.Height()*o.Icon.Scale()/2, nil

	case "y":
		if container.IsHorizontal() {
			return float64(container.WindowPositionY()) + o.Icon.DrawY() + o.Icon.Height()*o.Icon.Scale()/2, nil
		}
		return float64(container.WindowPositionX()) + o.Icon.DrawX() + o.Icon.Width()*o.Icon.Scale()/2, nil

	case "orientation":
	// 		CairoDockPositionType iScreenBorder = ((! pContainer->bIsHorizontal) << 1) | (! pContainer->bDirectionUp);
	// 		g_value_init (v, G_TYPE_UINT);
	// 		g_value_set_uint (v, iScreenBorder);

	case "container":
	// 		g_value_init (v, G_TYPE_UINT);
	// 		int iType = _get_container_type (pContainer);
	// 		g_value_set_uint (v, iType);

	case "width": // this is the dimension of the icon when it's hovered.
		w, _ := o.Icon.IconExtent()
		return w, nil

	case "height":
		_, h := o.Icon.IconExtent()
		return h, nil

	case "Xid":
		win := o.Icon.Window()
		if win == nil {
			return 0, nil
		}
		return uint64(uintptr(unsafe.Pointer(win))), nil // TODO: maybe fix

	case "has_focus":
		return o.Icon.Window() != nil && o.Icon.Window().IsActive(), nil

	default:
		return nil, errors.New("Get: unknown property=" + property)
	}

	return nil, nil
}

// GetAll returns all applet icon properties.
//
func (o *AppGldi) GetAll() *cdtype.DockProperties {
	props := &cdtype.DockProperties{}
	xid, _ := o.Get("Xid")
	props.Xid = xid.(uint64)

	uncastX, _ := o.Get("x")
	uncastY, _ := o.Get("y")
	props.X = uncastX.(int32)
	props.Y = uncastY.(int32)
	// props.Orientation = v.Value().(uint32)
	// props.Container = v.Value().(uint32)
	w, h := o.Icon.IconExtent()
	props.Width, props.Height = int32(w), int32(h)
	props.HasFocus = o.Icon.Window() != nil && o.Icon.Window().IsActive()

	return props
}

//
//----------------------------------------------------------------[ SUBICONS ]--

// AddSubIcon adds subicons by pack of 3 strings : label, icon, ID. See cdtype.AppIcon.
//
func (o *AppGldi) AddSubIcon(fields ...string) error {
	if len(fields)%3 != 0 {
		return errors.New("AddSubIcon: bad arguments count (must use 3 string per icon)")
	}

	list, clist := gldi.PrepareNewIcons(fields)
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
func (o *IconBase) Animate(animation string, rounds int32) error {
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
func (o *IconBase) ShowDialog(message string, duration int32) error {
	addIdle(func() {
		// Prevent stacking dialog messages.
		o.RemoveDialogs()

		dialog.DialogShowTemporaryWithIcon(message, o.Icon, o.Icon.GetContainer(), 1000*float64(duration), "same icon")
	})
	return nil
}

//
//------------------------------------------------------------[ IDLE ACTIONS ]--

var redraw []func()

// addIdle adds a function to call on the next gtk idle cycle, to safely use
// the dock with our goroutines.
//
func addIdle(call func()) {
	if redraw == nil {
		glib.IdleAdd(callIdle)
	}
	redraw = append(redraw, call)
}

func callIdle() {
	for _, call := range redraw {
		call()
	}
	// println("idle called", len(redraw))
	redraw = nil
}
