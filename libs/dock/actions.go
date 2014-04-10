package dock

import "github.com/sqp/godock/libs/cdtype"

// Actions manages applet internal actions list.
//
type Actions struct {
	list                        []*Action // Actions defined.
	onActionStart, onActionStop func()    // Before and after main actions calls. Used to set display for threaded tasks.
	Max                         int       // Maximum number of concurrent actions (simultaneous).
	Current                     int       // Current number of active actions.
}

// Add action(s) to the list.
//
func (actions *Actions) Add(acts ...*Action) {
	for _, act := range acts {
		actions.list = append(actions.list, act)
	}
}

// Get action details by ID.
//
func (actions *Actions) Get(ID int) *Action {
	return actions.list[ID]
}

// ID finds the ID matching given action name.
//
func (actions *Actions) ID(name string) int {
	for _, act := range actions.list {
		if act.Name == name {
			return act.ID
		}
	}
	return 0
}

// Launch desired action by ID.
//
func (actions *Actions) Launch(ID int) {
	if actions.list[ID].Call == nil || (actions.Max > 0 && actions.Current >= actions.Max) {
		return
	}

	actions.Current++
	if actions.onActionStart != nil && actions.list[ID].Threaded {
		actions.onActionStart()
	}

	actions.list[ID].Call()

	if actions.onActionStart != nil && actions.list[ID].Threaded {
		actions.onActionStop()
	}
	actions.Current--
}

// Execute desired action by name.
//
// func (actions *Actions) LaunchName(name string) {
// 	for _, act := range actions.list {
// 		if act.Name == name {
// 			actions.Launch(act.ID)
// 			return
// 		}
// 	}
// }

// SetActionIndicators set the pre and post action callbacks.
//
func (actions *Actions) SetActionIndicators(onStart, onStop func()) {
	actions.onActionStart = onStart
	actions.onActionStop = onStop
}

// Action is an applet internal actions that can be used for callbacks or menu.
//
type Action struct {
	ID   int
	Name string
	Call func()
	Icon string
	Menu cdtype.MenuItemType

	// in fact all actions are threaded in the go version, but we could certainly
	// use this as a "add to actions queue" to prevent problems with settings
	// changed while working, or double launch.
	//
	Threaded bool
}

// BuildMenu construct and popup the menu with the given actions list.
//
func (cda *CDApplet) BuildMenu(actionIds []int) {
	var menu []string
	for _, ID := range actionIds {
		act := cda.Actions.Get(ID)
		switch act.Menu {
		case cdtype.MenuEntry:
			menu = append(menu, act.Name)
		// case cdtype.MenuSubMenu:
		case cdtype.MenuSeparator:
			menu = append(menu, "")
			// case cdtype.MenuCheckBox:
			// case cdtype.MenuRadioButton:
		}
	}
	cda.PopulateMenu(menu...)
}
