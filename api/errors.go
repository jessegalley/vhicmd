package api

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gookit/color"
)

// ErrorResponse represents the standard error response from the API
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Title   string `json:"title"`
}

// FormatErrorResponse takes an error response body and returns a clean, formatted error message
func FormatErrorResponse(responseBody string) string {
	var genericMap map[string]interface{}
	if err := json.Unmarshal([]byte(responseBody), &genericMap); err == nil {
		errorKeys := []string{
			"badRequest",
			"NeutronError",
			"itemNotFound",
			"computeFault",
			"unauthorizedError",
			"notFound",
			"forbidden",
			"conflictingRequest",
			"overLimit",
			"serverCapacityUnavailable",
			"serviceUnavailable",
			"volumeBackendAPIException",
			"HTTPBadRequest",
			"internalServerError", // 500s
			"invalidInput",
			"resourceNotFound",
			"quotaExceeded",
			"imageUnacceptable",
			"connectionRefused",
			"volumeFault",
			"deploymentErrors",
			"resourceInUse",
		}

		for _, key := range errorKeys {
			if errorObj, ok := genericMap[key].(map[string]interface{}); ok {
				for _, msgKey := range []string{"message", "Message"} {
					if msg, exists := errorObj[msgKey].(string); exists {
						return color.Style{color.FgRed}.Render(cleanErrorMessage(msg))
					}
				}
			}
		}

		for _, msgKey := range []string{"message", "Message", "error", "error_message"} {
			if msg, ok := genericMap[msgKey].(string); ok {
				return color.Style{color.FgRed}.Render(cleanErrorMessage(msg))
			}
		}
	}

	var errResp ErrorResponse
	if err := json.Unmarshal([]byte(responseBody), &errResp); err == nil && errResp.Message != "" {
		return color.Style{color.FgRed}.Render(cleanErrorMessage(errResp.Message))
	}

	return color.Style{color.FgRed}.Render(cleanErrorMessage(responseBody))
}

// cleanErrorMessage removes HTML tags and cleans up common formatting issues
func cleanErrorMessage(msg string) string {
	msg = regexp.MustCompile("<[^>]*>").ReplaceAllString(msg, "")
	msg = strings.ReplaceAll(msg, "&quot;", "\"")
	msg = strings.ReplaceAll(msg, "&apos;", "'")
	msg = strings.ReplaceAll(msg, "&lt;", "<")
	msg = strings.ReplaceAll(msg, "&gt;", ">")
	msg = strings.ReplaceAll(msg, "&amp;", "&")

	msg = regexp.MustCompile(`\s+`).ReplaceAllString(msg, " ")

	msg = strings.TrimSpace(msg)

	return msg
}
