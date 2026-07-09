package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(debug bool) {
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}

	output := zerolog.ConsoleWriter{
		Out:         os.Stdout,
		TimeFormat:  "2006-01-02 15:04:05.000",
		NoColor:     false,
		FormatLevel: formatLevel,
		FormatFieldName: func(i any) string {
			return ""
		},
		FormatFieldValue: func(i any) string {
			return fmt.Sprint(i)
		},
		FormatMessage: func(i any) string {
			return fmt.Sprint(i)
		},
	}

	Log = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	zerolog.TimestampFieldName = "time"
	zerolog.TimeFieldFormat = time.DateTime
}
