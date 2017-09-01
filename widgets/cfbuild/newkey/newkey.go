// Package newkey creates keys for the config file builder.
//
// If you need to use the key values before packing keys, you have to set the
// key builder with SetBuild.
package newkey

import (
	"github.com/gotk3/gotk3/gtk"

	"github.com/sqp/godock/libs/cdtype"
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
	"github.com/sqp/godock/widgets/gtk/newgtk"
)

// Call defines a named callback.
//
type Call struct {
	Label string
	Func  interface{}
	Args  []interface{}
}

// Bool creates a bool key.
//
func Bool(group, name, label string) *cftype.Key {
	key := newKey(group, name, cftype.KeyBoolButton)
	key.Text = label
	return key
}

// Custom creates a custom key.
//
func Custom(group, name, label string, makeWidget func(*cftype.Key)) *cftype.Key {
	key := newKey(group, name, cftype.KeyEmptyWidget)
	key.Text = label
	key.SetMakeWidget(makeWidget)
	return key
}

// CustomButton creates a custom key with a label button.
//
func CustomButton(group, name, label string, calls ...Call) *cftype.Key {
	return Custom(group, name, label, func(key *cftype.Key) {
		for _, cb := range calls {
			btn := newgtk.ButtonWithLabel(cb.Label)
			btn.Connect("clicked", cb.Func, cb.Args...)
			key.PackSubWidget(btn)
		}
	})
}

// EmptyFull creates an empty key at full size.
//
func EmptyFull(group, name string) *cftype.Key {
	return newKey(group, name, cftype.KeyEmptyFull)
}

// Frame creates a frame key.
//
func Frame(group, name, label, icon string) *cftype.Key {
	var values []string
	if label != "" {
		values = []string{label, icon}
	}
	key := newKey(group, name, cftype.KeyFrame)
	key.AuthorizedValues = values
	return key
}

// LaunchCommand creates a launch command key.
//
func LaunchCommand(group, name, label, command string) *cftype.Key {
	key := newKey(group, name, cftype.KeyLaunchCmdSimple)
	key.Text = label
	key.AuthorizedValues = []string{command}
	return key
}

// Link creates a link key.
//
func Link(group, name, label, linkText, url string) *cftype.Key {
	key := newKey(group, name, cftype.KeyLink)
	key.Text = label
	key.AuthorizedValues = []string{linkText, url} // url shouldn't be here, but that'll be fine for now.

	// key.ValueSet(url) // can't use that now, need SetBuilder.
	return key
}

// ListNumbered creates a bool key.
//
func ListNumbered(group, name, label string, values ...string) *cftype.Key {
	key := newKey(group, name, cftype.KeyListNumbered)
	key.Text = label
	key.AuthorizedValues = values
	return key
}

// Separator creates a separator key.
//
func Separator(group, name string) *cftype.Key {
	return newKey(group, name, cftype.KeySeparator)
}

// StringEntry creates a string entry key.
//
func StringEntry(group, name, label string) *cftype.Key {
	key := newKey(group, name, cftype.KeyStringEntry)
	key.Text = label
	return key
}

// SwitchText creates a switch key with extra togglable text.
// baseText is always visible and moreText only when active.
// if baseText is empty, the label won't be updated.
//
func SwitchText(log cdtype.Logger, getValue func() bool, setValue func(bool), baseText, moreText string) func(*cftype.Key) {
	return func(key *cftype.Key) {
		btn := newgtk.Switch()
		key.PackSubWidget(btn)

		setText := func(val bool) {
			if baseText != "" {
				disp := baseText
				if val {
					disp += moreText
				}
				key.Label().SetMarkup(disp)
			}
		}

		onClick := func() {
			val := !getValue()
			setValue(val)
			setText(val)
		}

		// Apply directly once.
		val := getValue()
		setText(val)
		btn.SetActive(val) // Button state.

		_, e := btn.Connect("notify::active", onClick)
		log.Err(e, "connect button")
	}
}

// TextArea creates a textarea key.
//
func TextArea(group, name, label string, log cdtype.Logger) *cftype.Key {
	return Custom(group, name, label, func(key *cftype.Key) {
		tv := newgtk.TextView()
		tv.SetWrapMode(gtk.WRAP_WORD)
		buf, e := tv.GetBuffer()
		if log.Err(e, "TextArea GetBuffer") {
			return
		}
		key.SetWidGetValue(func() interface{} {
			start, end := buf.GetBounds()
			str, e := buf.GetText(start, end, false)
			log.Err(e, "TextArea GetText")
			return str
		})
		key.SetWidSetValue(func(val interface{}) { buf.SetText(val.(string)) })
		val := ""
		e = key.ValueGet(&val)
		log.Err(e, "TextArea SetText")
		key.ValueSet(val)
		key.PackWidget(tv, true, true, 0)
	})
}

// TextLabel creates a text key.
//
func TextLabel(group, name, label string) *cftype.Key {
	key := newKey(group, name, cftype.KeyTextLabel)
	key.Text = label
	return key
}

func newKey(group, name string, typ cftype.KeyType) *cftype.Key {
	return &cftype.Key{
		Type:  typ,
		Group: group,
		Name:  name,
	}
}
