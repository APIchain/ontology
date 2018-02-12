package log

import (
	"testing"
)

func init() {
	Init(Path, Stdout)
	//set your debug level
	Log.SetDebugLevel(0)
}

func TestDebugPrint(t *testing.T) {
	Debug("debug testing")
}

func TestInfoPrint(t *testing.T) {
	Info("Info testing")
}

func TestWarningPrint(t *testing.T) {
	Warn("Warning testing")
}

func TestErrorPrint(t *testing.T) {
	Error("Error testing")
}

func TestFatalPrint(t *testing.T) {
	Fatal("Fatal testing")
}

func TestTracePrint(t *testing.T) {
	Trace("Trace testing")
}
