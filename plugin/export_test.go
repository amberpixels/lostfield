package plugin

import (
	"github.com/golangci/plugin-module-register/register"

	"github.com/amberpixels/lostfield"
)

// Internals exposed for black-box tests.
var NewPlugin = newPlugin

// ConfigOf returns the config a built plugin holds.
func ConfigOf(p register.LinterPlugin) *lostfield.Config { return p.(*plugin).cfg }
