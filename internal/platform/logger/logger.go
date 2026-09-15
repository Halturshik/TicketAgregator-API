package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	FormatText = "text"
	FormatJSON = "json"

	DefaultFormat = FormatText
	DefaultLevel  = "info"
)

type Options struct {
	Service string
	Format  string
	Level   string
	Output  io.Writer
}

func New(options Options) (*slog.Logger, error) {
	if options.Service == "" {
		return nil, fmt.Errorf("название сервиса логирования не указано")
	}
	if options.Format == "" {
		options.Format = DefaultFormat
	}
	if options.Level == "" {
		options.Level = DefaultLevel
	}
	if options.Output == nil {
		options.Output = os.Stdout
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(options.Level))); err != nil {
		return nil, fmt.Errorf("некорректный уровень логирования %q: %w", options.Level, err)
	}
	handlerOptions := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch strings.ToLower(options.Format) {
	case FormatText:
		handler = slog.NewTextHandler(options.Output, handlerOptions)
	case FormatJSON:
		handler = slog.NewJSONHandler(options.Output, handlerOptions)
	default:
		return nil, fmt.Errorf("неподдерживаемый формат логирования %q", options.Format)
	}

	handler = contextHandler{Handler: handler}
	return slog.New(handler).With(slog.String("service", options.Service)), nil
}
