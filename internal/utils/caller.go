package utils

import (
	"fmt"
	"runtime"
)

type Caller struct {
	// FilePath is the absolute file path of the caller
	FilePath string
	// LineNumber is the line number of the caller
	LineNumber int
	// FunctionName is the name of the function of the caller
	FunctionName string
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

	details := runtime.FuncForPC(pc)
	if details != nil {
		functionName = details.Name()
	}

	return Caller{
		FilePath:     filePath,
		LineNumber:   line,
		FunctionName: functionName,
	}
}

// String returns a string representation of the Caller struct.
func (c Caller) String() string {
	return fmt.Sprintf("%s:%d (%s)", c.FilePath, c.LineNumber, c.FunctionName)
}
