package dock

import "github.com/sqp/godock/libs/cdtype"

//
//-----------------------------------------------------------------[ ACTIONS ]--

// Actions manages applet internal actions list.
//
type Actions struct {
	list                        []*cdtype.Action // Actions defined.
	onActionStart, onActionStop func()           // Before and after main actions calls. Used to set display for threaded tasks.
	Max                         int              // Maximum number of concurrent actions (simultaneous).
	Current                     int              // Current number of active actions.
}

// ActionAdd adds actions to the list.
//
func (o *Actions) ActionAdd(acts ...*cdtype.Action) {
	for _, act := range acts {
		o.list = append(o.list, act)
	}
}

// ActionID finds the ID matching given action name.
//
func (o *Actions) ActionID(name string) int {
	for _, act := range o.list {
		if act.Name == name {
			return act.ID
		}
	}
	return 0
}

// ActionLaunch starts the desired action by ID.
//
func (o *Actions) ActionLaunch(ID int) {
	if o.list[ID].Call == nil || (o.Max > 0 && o.Current >= o.Max) {
		return
	}

	o.Current++
	if o.onActionStart != nil && o.list[ID].Threaded {
		o.onActionStart()
	}

	o.list[ID].Call()

	if o.onActionStart != nil && o.list[ID].Threaded {
		o.onActionStop()
	}
	o.Current--
}

// ActionCallback returns a callback to the given action ID.
//
func (o *Actions) ActionCallback(ID int) func() {
	return func() { o.ActionLaunch(ID) }
}

// ActionCount returns the number of started actions.
//
func (o *Actions) ActionCount() int {
	return o.Current
}

// ActionSetMax sets the maximum number of actions that can be started at the same time.
//
func (o *Actions) ActionSetMax(max int) {
	o.Max = max
}

// ActionSetBool sets the pointer to the boolean value for a checkentry menu field.
//
func (o *Actions) ActionSetBool(ID int, boolPointer *bool) {
	o.list[ID].Bool = boolPointer
}

// ActionSetIndicators set the pre and post action callbacks.
//
func (o *Actions) ActionSetIndicators(onStart, onStop func()) {
	o.onActionStart = onStart
	o.onActionStop = onStop
}

//
//---------------------------------------------------------------[ BUILDMENU ]--

// BuildMenu fills the menu with the given actions list.
//
func (o *Actions) BuildMenu(menu cdtype.Menuer, actionIds []int) {
	for _, ID := range actionIds {
		act := o.list[ID]
		var entry cdtype.MenuWidgeter
		switch act.Menu {
		case cdtype.MenuEntry:
			entry = menu.AddEntry(act.Name, act.Icon, o.ActionCallback(act.ID))

		case cdtype.MenuSeparator:
			menu.Separator()

		case cdtype.MenuCheckBox:
			entry = menu.AddCheckEntry(act.Name, *act.Bool, o.ActionCallback(act.ID))
			if act.Call == nil {
				act.Call = func() {
					*act.Bool = !*act.Bool
				}
			}

		case cdtype.MenuRadioButton:
			entry = menu.AddRadioEntry(act.Name, *act.Bool, act.Group, o.ActionCallback(act.ID))

			// case cdtype.MenuSubMenu:
		}

		if entry != nil && act.Tooltip != "" {
			entry.SetTooltipText(act.Tooltip)
		}
	}
}

// BuildMenuCallback provides a fill menu callback with the given actions list.
//
func (o *Actions) BuildMenuCallback(actionIds []int) func(menu cdtype.Menuer) {
	return func(menu cdtype.Menuer) { o.BuildMenu(menu, actionIds) }
}
