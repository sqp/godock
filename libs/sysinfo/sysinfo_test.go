package sysinfo_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/sysinfo"

	"os"
	"testing"
)

var logger = log.NewLog(log.Logs)

func TestSystem(t *testing.T) {
	vc, e := sysinfo.NewVideoCard(logger)
	assert.NoError(t, e, "sysinfo.NewVideoCard")

	assert.NotEmpty(t, vc.CoreProfileOpenGL, "CoreProfileOpenGL")
	assert.NotEmpty(t, vc.Renderer, "Renderer")
	assert.NotEmpty(t, vc.Vendor, "Vendor")
	assert.NotEmpty(t, vc.VersionOpenGL, "VersionOpenGL")

	mem, e := sysinfo.ProcessMemory(os.Getpid())
	assert.NoError(t, e, "sysinfo.ProcessMemory")
	assert.True(t, mem > 1000, "sysinfo.ProcessMemory")
}
