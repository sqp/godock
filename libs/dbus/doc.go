/* Update is an applet for Cairo-Dock to check for its new versions and do update.

Copyright : (C) 2012 by SQP
E-mail : sqp@glx-dock.org

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 3
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.
http://www.gnu.org/licenses/licenses.html#GPL */

/* 
godock/libs/dbus is a library to build cairo-dock applets with the DBus connector.

This is a part of the external applets for Cairo-Dock

Examples:
* Some of the actions on the main icon:
	demo.SetQuickInfo("OK")
	demo.SetLabel("label changed")
	demo.ShowDialog("dialog string\n with time in second", 8)
	demo.BindShortkey("<Control><Shift>Z", "<Alt>K")
	demo.SetIcon("/usr/share/icons/gnome/32x32/actions/gtk-media-pause.png")
	demo.SetEmblem("/usr/share/icons/gnome/32x32/actions/gtk-go-down.png", dock.EmblemTopRight)
	demo.ControlAppli("devhelp")

* Some of the actions to play with SubIcons:
	demo.AddSubIcon([]string{
		"icon 1", "firefox-3.0", "id1",
		"icon 2", "chromium-browser", "id2",
		"icon 3", "geany", "id3",
	})
	demo.RemoveSubIcon("id1")

	demo.Icons["id3"].SetQuickInfo("woot")
	demo.Icons["id2"].SetLabel("label changed")
	demo.Icons["id3"].Animate("fire", 3)


demo.AskText("Enter your name", "<my name>")
demo.AskValue("needed value", 0, 42)
demo.AskQuestion("why?")



*CDApplet.ShowAppli(show bool) error
*CDApplet.DemandsAttention(start bool, animation string) error
*CDApplet.PopulateMenu(items... string) error
*CDApplet.Get(property string) ([]interface{}, error) 
*CDApplet.GetAll() (*DockProperties, error)

*CDApplet.AskText(message, initialText string) error 
*CDApplet.AskValue(message string, initialValue, maxValue float64) error 
*CDApplet.AskQuestion(message string) error 


//~ 
//~ demo.AskText("Enter your name", "<my name>")
demo.AskValue("needed value", 0, 42)
//~ demo.AskQuestion("why?")


Still to do;
* Handle the config file somehow.
* Actions without effect:
	AddDataRenderer("gauge", 2, "Turbo-night-fuel")
	RenderValues([]float64{12, 32})
	PopupDialog


To improve :
* detect callback methods on given implementation and connect only those found to DBus.
Will also allow to remove many methods from the main interface.


	public abstract Variant Get (string cProperty) throws IOError;
	public abstract HashTable<string,Variant> GetAll () throws IOError;




Copyright : (C) 2012 by SQP
E-mail : sqp@glx-dock.org
*/
package documentation
