package cdtype_test

import "github.com/sqp/godock/libs/cdtype"

type AppTest struct{ cdtype.AppBase }

func (app *AppTest) ExampleAppIcon_popupDialogListFixed() {
	values := []string{"entry 1", "entry two", "or more"}
	app.PopupDialog(cdtype.DialogData{
		Message:   "Please choose:",
		UseMarkup: true,
		Widget: cdtype.DialogWidgetList{
			Values:       values,
			InitialValue: 2, // select "or more".
		},
		Buttons: "ok;cancel",
		Callback: cdtype.DialogCallbackValidInt(func(id int) {
			app.Log().Info("ID", id, "for string", values[id])
			// ... do something with your validated value ...
		}),
	})
}

func (app *AppTest) ExampleAppIcon_popupDialogListEditable() {
	app.PopupDialog(cdtype.DialogData{
		Message:   "Please choose:",
		UseMarkup: true,
		Widget: cdtype.DialogWidgetList{
			Values:       []string{"one", "two", "three"},
			InitialValue: "four", // a custom entry not in the list.
		},
		Buttons: "ok;cancel",
		Callback: cdtype.DialogCallbackValidString(func(str string) {
			app.Log().Info("value", str)
			// ... do something with your validated value ...
		}),
	})
}

// Examples of the most common icon actions.
func (app *AppTest) ExampleAppIcon_howTo() {
	// Main icon actions.

	app.SetQuickInfo("OK")
	app.SetLabel("label changed")
	app.SetIcon("/usr/share/icons/gnome/32x32/actions/media-playback-pause.png")
	app.SetEmblem("/usr/share/icons/gnome/32x32/actions/go-down.png", cdtype.EmblemTopRight)
	app.Animate("fire", 10)
	app.DemandsAttention(true, "default")
	app.ShowDialog("dialog string\n with time in second", 8)

	app.BindShortkey(&cdtype.Shortkey{
		ConfGroup: "Actions",
		ConfKey:   "ShortkeyOneKey",
		Desc:      "Action one",
		Shortkey:  "<Alt>K",
	})

	// Renderer (gauge, graph, progress-bar).

	app.DataRenderer().Gauge(2, "Turbo-night-fuel")
	app.DataRenderer().Render(0.2, 0.7)

	// Application monitoring.

	app.Window().SetAppliClass("devhelp") // in Init.

	// on click
	if app.Window().IsOpened() { // Window opened.
		app.Window().ToggleVisibility()
	}

	app.Window().Close() // on other event...

	// Icon and dock settings.

	pos, e := app.IconProperty().ContainerPosition()
	if !app.Log().Err(e, "IconProperty ContainerPosition") {
		app.Log().Info("Container position", pos)
	}

	properties, e := app.IconProperties()
	if !app.Log().Err(e, "IconProperties") {
		app.Log().Info("Container position", properties.ContainerPosition())
		app.Log().Info("Icon width", properties.Width())
		app.Log().Info("Icon height", properties.Height())
	}

	// Add, remove and play with SubIcons:

	app.AddSubIcon(
		"icon 1", "firefox-3.0", "id1",
		"text 2", "chromium-browser", "id2",
		"1 more", "geany", "id3",
	)
	app.RemoveSubIcon("id1")

	app.SubIcon("id3").SetQuickInfo("OK")
	app.SubIcon("id2").SetLabel("label changed")
	app.SubIcon("id3").Animate("fire", 3)

	app.RemoveSubIcons()
}
