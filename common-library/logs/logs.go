package logs

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type LogEntry struct {
	Time      string `json:"time"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Env       string `json:"env"`
	Version   string `json:"version"`
	Tenant    string `json:"flyr_tenant"`
	Message   string `json:"message"`
	Caller    string `json:"caller"`
	Error     string `json:"error,omitempty"`
	TraceID   string `json:"dd.trace_id,omitempty"`
	SpanID    string `json:"dd.span_id,omitempty"`
	UserID    string `json:"userId,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// Initialize a logger with obfuscation for sensitive data
func InitLogger(service string, env string, version string, tenant string) {
	log.SetFlags(0) // Disable default timestamp; we are adding custom timestamps
}

// Log function to handle sensitive data obfuscation and JSON formatting
func Log(level string, message string, caller string, err error, traceID string, spanID string, userID string, requestID string) {
	entry := LogEntry{
		Time:      time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Service:   "ooms-service", // Replace with actual service name
		Env:       "prod",         // Replace with actual environment
		Version:   "v1.0.0",       // Replace with actual version
		Tenant:    "rx",           // Replace with actual tenant
		Message:   message,
		Caller:    caller,
		TraceID:   traceID,
		SpanID:    spanID,
		UserID:    obfuscateUserID(userID), // Obfuscate sensitive data
		RequestID: requestID,
		Error:     fmt.Sprintf("%v", err),
	}

	// Convert log entry to JSON
	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error marshalling log entry: %v", err)
		return
	}

	log.Println(string(jsonEntry))
}

// Utility function to obfuscate sensitive User IDs
func obfuscateUserID(userID string) string {
	if len(userID) > 0 {
		// Return a masked version of the user ID, e.g., hash or partial obfuscation
		return "obfuscated-user-id"
	}
	return ""
}
