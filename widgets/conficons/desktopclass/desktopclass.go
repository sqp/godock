// Package desktopclass extends the launcher builder with desktop class informations.
package desktopclass

import (
	"github.com/conformal/gotk3/gtk"

	"github.com/sqp/godock/libs/text/strhelp" // String helpers.

	"github.com/sqp/godock/widgets/cfbuild/cftype"   // Types for config file builder usage.
	"github.com/sqp/godock/widgets/cfbuild/datatype" // Types for config file builder data source.
	"github.com/sqp/godock/widgets/cfbuild/newkey"   // Create config file builder keys.
	"github.com/sqp/godock/widgets/common"           // Text format gtk.
	"github.com/sqp/godock/widgets/gtk/newgtk"       // Create widgets.

	"os"
	"path/filepath"
	"strings"
)

// Tweak prepares a build tweak to add desktop class informations for a launcher.
//
func Tweak(build cftype.Grouper, source datatype.Source, selected datatype.DesktopClasser) func(cftype.Builder) {
	return func(build cftype.Builder) {
		// Create a key for the hidden existing Origin entry.
		keyOrigin := newkey.EmptyFull(cftype.DesktopEntry, "Origin")

		// Check if a launcher origin has been set (as a desktop class file location).
		keyOrigin.SetBuilder(build)
		origins := keyOrigin.Value().String()
		if origins == "" {
			return
		}

		// Prepare the desktop class widget.
		keyOrigin.SetMakeWidget(func(key *cftype.Key) {
			w := Widget(source, selected, origins)
			key.PackWidget(w, false, false, 0)
		})

		// Add the key to the builder.
		build.AddKeys(cftype.DesktopEntry,
			newkey.Frame(cftype.DesktopEntry, "FrameOrigin", "", ""), // A frame to close the expander.
			keyOrigin)
	}
}

// Widget creates a desktop class informations widget.
//
func Widget(source datatype.Source, selected datatype.DesktopClasser, origins string) gtk.IWidget {
	apps := strings.Split(origins, ";")
	if len(apps) == 0 {
		return nil
	}

	// Remove the path from the first item.
	dir := filepath.Dir(apps[0])
	apps[0] = filepath.Base(apps[0])

	// Try force select the first one (can be inactive if "do not bind appli").
	if selected.String() == "" {
		selected = source.DesktopClasser(strings.TrimSuffix(apps[0], ".desktop"))
	}

	command := selected.Command()
	desktopFile := desktopFileText(apps, dir, selected.String())

	wName := boxLabel(selected.Name())
	wIcon := boxLabel(selected.Icon())
	wCommand := boxButton(command, func() { println("need to launch", command) })
	wDesktopFiles := boxLabel(desktopFile)

	grid := newgtk.Grid()
	grid.Attach(boxLabel("Name"), 0, 0, 1, 1)
	grid.Attach(boxLabel("Icon"), 0, 1, 1, 1)
	grid.Attach(boxLabel("Command"), 0, 2, 1, 1)
	grid.Attach(boxLabel("Desktop file"), 0, 3, 1, 1)
	grid.Attach(wName, 1, 0, 1, 1)
	grid.Attach(wIcon, 1, 1, 1, 1)
	grid.Attach(wCommand, 1, 2, 1, 1)
	grid.Attach(wDesktopFiles, 1, 3, 1, 1)

	frame := newgtk.Frame("")
	label := newgtk.Label(common.Bold("Launcher origin"))
	label.SetUseMarkup(true)
	frame.SetLabelWidget(label)
	frame.Add(grid)
	frame.ShowAll()
	return frame
}

func boxButton(label string, call func()) *gtk.Box {
	btnCommand := newgtk.ButtonWithLabel(label)
	btnCommand.Connect("clicked", call)
	btnCommand.SetRelief(gtk.RELIEF_HALF)
	return boxWidget(btnCommand)
}

func boxLabel(str string) *gtk.Box {
	label := newgtk.Label(str + "\t")
	label.SetUseMarkup(true)
	return boxWidget(label)
}

func boxWidget(widget gtk.IWidget) *gtk.Box {
	box := newgtk.Box(gtk.ORIENTATION_HORIZONTAL, 0)
	box.PackStart(widget, false, false, 0)
	return box
}

func fileExists(path string) bool {
	_, e := os.Stat(path)
	return e == nil || !os.IsNotExist(e)

}

func desktopFileText(apps []string, dir, selected string) string {
	text := ""
	for _, v := range apps {
		// Remove suffix for name and highlight the active one (with link if possible).
		name := strings.TrimSuffix(v, ".desktop")
		isCurrent := name == selected

		if fileExists(filepath.Join(dir, v)) {
			name = common.URI("file://"+filepath.Join(dir, v), name)
		}

		if isCurrent {
			name = common.Bold(name)
		}
		text = strhelp.Separator(text, ", ", name)
	}
	return text
}
