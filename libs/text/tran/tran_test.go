package tran_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/text/tran"

	"os"
	"testing"
)

func TestTranslate(t *testing.T) {
	os.Setenv("LANGUAGE", "fr")

	// Translate dock strings.
	for in, out := range map[string]string{
		"Add":    "Ajouter",
		"Edit":   "Éditer",
		"Window": "Fenêtre",
	} {
		ret := tran.Slate(in)
		assert.Equal(t, ret, out, "translate %s != %s", ret, out)
	}

	// Translate plugins directly.
	for in, out := range map[string]string{
		"Animation when music changes:": "Animation au changement de musique",
		"Saturday":                      "Samedi",
	} {
		ret := tran.Splug(in)
		assert.Equal(t, ret, out, "translate %s != %s", ret, out)
	}

	// Translate plugins with domain name.
	for in, out := range map[string]string{
		"Lock position?":    "Verrouiller la position ?",
		"Image filename:":   "Nom du fichier de l'image :",
		"Font:":             "Police",
		"Use a custom font": "Utiliser une police personnalisée",
	} {
		ret := tran.Sloc("cairo-dock-plugins", in)
		assert.Equal(t, ret, out, "translate %s != %s", ret, out)
	}

	// Other string unchanged.
	for _, str := range []string{"unknown", "unchanged"} {
		ret := tran.Slate(str)
		assert.Equal(t, ret, str, "untranslated %s != %s", ret, str)
	}
}
