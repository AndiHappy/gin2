package gin

import (
	"flag"
	"io"
	"os"
	"sync/atomic"
)

// DefaultWriter is the default io.Writer used by Gin for debug output and
// middleware output like Logger() or Recovery().
// Note that both Logger and Recovery provides custom ways to configure their
// output io.Writer.
// To support coloring in Windows use:
//
//	import "github.com/mattn/go-colorable"
//	gin.DefaultWriter = colorable.NewColorableStdout()
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Gin to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

// EnvGinMode indicates environment name for gin mode.
const EnvGinMode = "GIN_MODE"

const (
	debugCode = iota // debugCode indicates that the framework is running in debug mode.
	releaseCode
	testCode
)

const (
	// DebugMode indicates gin mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates gin mode is release.
	ReleaseMode = "release"
	// TestMode indicates gin mode is test.
	TestMode = "test"
)

// 初始化的时候执行
func init() {
	mode := os.Getenv(EnvGinMode)
	SetMode(mode)
}

// SetMode sets gin mode according to input string.
func SetMode(value string) {
	if value == "" {
		if flag.Lookup("test.v") != nil {
			value = TestMode
		} else {
			value = DebugMode
		}
	}

	switch value {
	case DebugMode, "":
		atomic.StoreInt32(&ginMode, debugCode)
	case ReleaseMode:
		atomic.StoreInt32(&ginMode, releaseCode)
	case TestMode:
		atomic.StoreInt32(&ginMode, testCode)
	default:
		panic("gin mode unknown: " + value + " (available mode: debug release test)")
	}
	modeName.Store(value)
}

var (
	ginMode  int32 = debugCode
	modeName atomic.Value
)
