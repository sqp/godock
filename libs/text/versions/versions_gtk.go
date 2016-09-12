// +build dock gtk

package versions

import (
	"github.com/sqp/godock/libs/gldi/globals" // Dock globals.

	"fmt"
)

func init() {
	gtkX, gtkY, gtkZ := globals.GtkVersion()
	List = append(List,
		Field{"  GTK      ", fmt.Sprintf("%d.%d.%d", gtkX, gtkY, gtkZ)},
	)
}
