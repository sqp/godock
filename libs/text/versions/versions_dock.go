// +build dock

package versions

import (
	"github.com/sqp/godock/libs/gldi"
	"github.com/sqp/godock/libs/gldi/globals" // Dock globals.
	"github.com/sqp/godock/libs/ternary"      // Ternary operators.
)

func init() {
	Name = "Custom Dock"

	List = append(List,
		Field{"  gldi     ", globals.Version()},
		Field{"  OpenGL   ", ternary.String(gldi.GLBackendIsUsed(), "Yes", "No")},
	)
}
