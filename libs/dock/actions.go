package dock

import "github.com/sqp/godock/libs/cdtype"

type Actions struct {
	list                        []*Action
	onActionStart, onActionStop func() // Before and after main actions calls. Used to set display for threaded tasks.
}

func (actions *Actions) Add(acts ...*Action) {
	for _, act := range acts {
		actions.list = append(actions.list, act)
	}
}

// Get action details.
//
func (actions *Actions) Get(id int) *Action {
	return actions.list[id]
}

// Find id for given action name.
//
func (actions *Actions) Id(name string) int {
	for _, act := range actions.list {
		if act.Name == name {
			return act.Id
		}
	}
	return 0
}

// Execute desired action by id.
//
func (actions *Actions) Launch(id int) {
	if actions.onActionStart != nil && actions.list[id].Threaded {
		actions.onActionStart()
	}

	actions.list[id].Call()

	if actions.onActionStart != nil && actions.list[id].Threaded {
		actions.onActionStop()
	}
}

// Execute desired action by name.
//
// func (actions *Actions) LaunchName(name string) {
// 	for _, act := range actions.list {
// 		if act.Name == name {
// 			actions.Launch(act.Id)
// 			return
// 		}
// 	}
// }

func (actions *Actions) SetActionIndicators(onStart, onStop func()) {
	actions.onActionStart = onStart
	actions.onActionStop = onStop
}

type Action struct {
	Id   int
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

// Popup the menu with the given actions list.
//
func (cda *CDApplet) BuildMenu(actionIds []int) {
	var menu []string
	for _, id := range actionIds {
		act := cda.Actions.Get(id)
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
