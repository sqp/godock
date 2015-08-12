// Package newkey creates keys for the config file builder.
//
// If you need to use the key values before packing keys, you have to set the
// key builder with SetBuild.
package newkey

import (
	"github.com/sqp/godock/widgets/cfbuild/cftype" // Types for config file builder usage.
	"github.com/sqp/godock/widgets/gtk/newgtk"
)

// Custom creates a custom key.
//
func Custom(group, name, label string, makeWidget func(*cftype.Key)) *cftype.Key {
	key := &cftype.Key{
		Type:  cftype.KeyLink, // TODO: improve.
		Group: group,
		Name:  name,
		Text:  label,
	}

	key.SetMakeWidget(makeWidget)
	return key
}

// CustomButtonLabel creates a custom key with a label button.
//
func CustomButtonLabel(group, name, label, btnText string, call func(), args ...interface{}) *cftype.Key {
	return Custom(group, name, label, func(key *cftype.Key) {
		btn := newgtk.ButtonWithLabel(btnText)
		btn.Connect("clicked", call, args)
		key.PackSubWidget(btn)
	})
}

// EmptyFull creates an empty key at full size.
//
func EmptyFull(group, name string) *cftype.Key {
	return &cftype.Key{
		Type:  cftype.KeyEmptyFull,
		Group: group,
		Name:  name,
	}
}

// Frame creates a frame key.
//
func Frame(group, name, label, icon string) *cftype.Key {
	var values []string
	if label != "" {
		values = []string{label, icon}
	}
	return &cftype.Key{
		Type:             cftype.KeyFrame,
		Group:            group,
		Name:             name,
		AuthorizedValues: values,
	}
}

// LaunchCommand creates a launch command key.
//
func LaunchCommand(group, name, label, command string) *cftype.Key {
	return &cftype.Key{
		Type:             cftype.KeyLaunchCmdSimple,
		Group:            group,
		Name:             name,
		Text:             label,
		AuthorizedValues: []string{command},
	}
}

// Link creates a link key.
//
func Link(group, name, label, linkText, url string) *cftype.Key {
	key := &cftype.Key{
		Type:             cftype.KeyLink,
		Group:            group,
		Name:             name,
		Text:             label,
		AuthorizedValues: []string{linkText, url}, // url shouldn't be here, but that'll be fine for now.
	}
	// key.ValueSet(url) // can't use that now, need SetBuilder.
	return key
}

// Separator creates a separator key.
//
func Separator(group, name string) *cftype.Key {
	return &cftype.Key{
		Type:  cftype.KeySeparator,
		Group: group,
		Name:  name,
	}
}

// TextLabel creates a text key.
//
func TextLabel(group, name, label string) *cftype.Key {
	return &cftype.Key{
		Type:  cftype.KeyTextLabel,
		Group: group,
		Name:  name,
		Text:  label,
	}
}
