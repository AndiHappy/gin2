package gin

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(gin.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	return atomic.LoadInt32(&ginMode) == debugCode
}

// DebugPrintFunc indicates debug log output format.
// 可以声明和试用默认的DebugPrintFunc，
var DebugPrintFunc func(format string, values ...any)

func debugPrint(format string, values ...any) {
	if !IsDebugging() {
		return
	}

	if DebugPrintFunc != nil {
		DebugPrintFunc(format, values...)
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(DefaultWriter, "[GIN-debug] "+format, values...)
}

func debugPrintError(err error) {
	if err != nil && IsDebugging() {
		fmt.Fprintf(DefaultErrorWriter, "[GIN-debug] [ERROR] %v\n", err)
	}
}
