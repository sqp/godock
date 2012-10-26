package dbus

//~ import "fmt"
//~ public void set_emblem_busy () {
//~ try { this.icon.SetEmblem (this.sAppletDirectory + "/icons/emblem-important.svg", 3); }
//~ catch (Error e) {}
//~ }
//~ 
//~ public void set_emblem_none () {
//~ try { this.icon.SetEmblem ("", 3); }
//~ catch (Error e) {}
//~ }

type Actions []*Action

type Action struct {
	Id       int
	Name     string
	Call     func()
	Icon     string
	Icontype int
	Threaded bool
}


func (cda *CDApplet) SetActionIndicators(onStart, onStop func ()) {
	cda.onActionStart = onStart
	cda.onActionStop = onStop
}

func (cda *CDApplet) AddAction(actions ...*Action) {
	for _, act := range actions {
		cda.Actions = append(cda.Actions, act)
	}
}

func (cda *CDApplet) Launch(id int) {
	if cda.Actions[id].Threaded {
		if cda.onActionStart != nil {
			cda.onActionStart()
		}
		cda.Actions[id].Call()
		if cda.onActionStart != nil {
			cda.onActionStop()
		}
	} else {
		cda.Actions[id].Call()
	}
}

//~ func (a Actions) Launch(id int) {
	//~ if a[id].Threaded {
		//~ set_emblem_busy ();
		//~ go a[id].Call()
	//~ } else {
		//~ a[id].Call()
	//~ }
//~ }

//~ func (a Actions) GetCall(id int) func() {
	//~ return a[id].Call
//~ }
