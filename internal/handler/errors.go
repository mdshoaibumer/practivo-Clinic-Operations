package handler

import (
	"errors"
	"log/slog"
	"runtime"

	"clinmitra/internal/utils"
)

// safeError ensures only AppError types are returned to the frontend.
// Internal errors are logged with caller context and replaced with a generic
// message to prevent leaking implementation details through the Wails binding layer.
func safeError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Capture caller info for debugging
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	fnName := "unknown"
	if fn != nil {
		fnName = fn.Name()
	}

	slog.Error("internal error",
		"error", err.Error(),
		"caller", fnName,
		"file", file,
		"line", line,
	)
	return utils.NewError("INTERNAL_ERROR", "An unexpected error occurred. Please try again.")
}
