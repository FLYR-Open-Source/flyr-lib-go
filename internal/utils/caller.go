package utils // import "github.com/FlyrInc/flyr-lib-go/internal/utils"

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/FlyrInc/flyr-lib-go/internal/config"
	otel_attribute "go.opentelemetry.io/otel/attribute"
)

// Caller represents the caller information of a function.
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
		namespace, functionName = splitFunctionName(details.Name())
	}

	return Caller{
		FilePath:     filePath,
		LineNumber:   line,
		FunctionName: functionName,
		Namespace:    namespace,
	}
}

// splitFunctionName splits full function name into namespace and function name
// if the passed function name does not contain a namespace, then it returns an empty string for the namespace
// and the passed function name.
func splitFunctionName(function string) (namespace, functionName string) {
	split := strings.Split(function, ".")
	if len(split) > 1 {
		return split[0], split[1]
	}
	return "", function
}

// String returns a string representation of the Caller struct.
func (c Caller) String() string {
	return fmt.Sprintf("%s:%d (%s)", c.FilePath, c.LineNumber, c.FunctionName)
}

// Custom MarshalJSON method to dynamically set JSON field names
func (c Caller) MarshalJSON() ([]byte, error) {

	// Define a map to hold the JSON structure with dynamic keys
	data := map[string]interface{}{
		config.FILE_PATH:             c.FilePath,
		config.LINE_NUMBER:           c.LineNumber,
		config.FUNCTION_NAME:         c.FunctionName,
		config.FUNCTION_PACKAGE_NAME: c.Namespace,
	}
	return json.Marshal(data)
}

// LogAttributes returns the caller attributes in a structured slog format.
func (c Caller) LogAttributes() []slog.Attr {
	codeFilePath := slog.String(config.FILE_PATH, c.FilePath)
	codeLineNumber := slog.Int(config.LINE_NUMBER, c.LineNumber)
	codeFunctionName := slog.String(config.FUNCTION_NAME, c.FunctionName)
	codeNamespace := slog.String(config.FUNCTION_PACKAGE_NAME, c.Namespace)

	return []slog.Attr{codeFilePath, codeLineNumber, codeFunctionName, codeNamespace}
}

// SpanAttributes returns the caller attributes in a structured OpenTelemetry format.
func (c Caller) SpanAttributes() []otel_attribute.KeyValue {
	codeFilePath := otel_attribute.String(config.FILE_PATH, c.FilePath)
	codeLineNumber := otel_attribute.Int(config.LINE_NUMBER, c.LineNumber)
	codeFunctionName := otel_attribute.String(config.FUNCTION_NAME, c.FunctionName)
	codeNamespace := otel_attribute.String(config.FUNCTION_PACKAGE_NAME, c.Namespace)

	return []otel_attribute.KeyValue{codeFilePath, codeLineNumber, codeFunctionName, codeNamespace}
}
