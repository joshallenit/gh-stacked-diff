package execute

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"

	"github.com/fatih/color"
)

// Options for [PrettyHandler].
type PrettyHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

// Handler for [slog] that uses diffdrent ANSI colors for each level (DEBUG, INFO, etc.).
//
// Modified from https://betterstack.com/community/guides/logging/logging-in-go/
type PrettyHandler struct {
	slog.Handler
	l *log.Logger
}

// Verify that [PrettyHandler] implements [slog.Handler].
var _ slog.Handler = new(PrettyHandler)

// Print a log record.
func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	msg := color.CyanString(r.Message)

	if len(fields) > 0 {
		h.l.Println(level, msg, color.WhiteString(string(b)))
	} else {
		h.l.Println(level, msg)
	}
	return nil
}

// Create a new [PrettyHandler].
func NewPrettyHandler(
	out io.Writer,
	opts PrettyHandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, &opts.SlogOpts),
		l:       log.New(out, "", 0),
	}

	return h
}
