package utils

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/FlyrInc/flyr-lib-go/config"
)

type Caller struct {
	// FilePath is the absolute file path of the caller
	FilePath string
	// LineNumber is the line number of the caller
	LineNumber int
	// FunctionName is the name of the function of the caller
	FunctionName string
	// The “namespace” within which code.function
	Namespace string
}

// GetCallerName retrieves caller information from the call stack.
//
// This function obtains the caller's function name, file path, and line number
// by inspecting the runtime call stack. The number of frames to skip is specified
// by numofSkippedFrames, allowing it to identify a specific caller level. If the
// function details are available, they are extracted; otherwise, a default empty
// string is used for the function name.
//
// Returns a Caller struct containing:
//   - FilePath: the absolute path of the file where the caller is located.
//   - LineNumber: the line number within the file.
//   - FunctionName: the name of the caller function.
func GetCallerName(numofSkippedFrames int) Caller {
	pc, filePath, line, _ := runtime.Caller(numofSkippedFrames)
	functionName := ""
	namespace := ""

	details := runtime.FuncForPC(pc)
	if details != nil {
		functionName = details.Name()

		// Split the function name to extract the namespace and function name
		split := strings.Split(functionName, ".")
		if len(split) > 1 {
			namespace = split[0]
			functionName = split[1]
		}
	}

	return Caller{
		FilePath:     filePath,
		LineNumber:   line,
		FunctionName: functionName,
		Namespace:    namespace,
	}
}

// String returns a string representation of the Caller struct.
func (c Caller) String() string {
	return fmt.Sprintf("%s:%d (%s)", c.FilePath, c.LineNumber, c.FunctionName)
}

// Custom MarshalJSON method to dynamically set JSON field names
func (c Caller) MarshalJSON() ([]byte, error) {
	// Define a map to hold the JSON structure with dynamic keys
	data := map[string]interface{}{
		config.CODE_PATH: c.FilePath,
		config.CODE_LINE: c.LineNumber,
		config.CODE_FUNC: c.FunctionName,
		config.CODE_NS:   c.Namespace,
	}
	return json.Marshal(data)
}
