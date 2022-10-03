package main

import (
	"github.com/spf13/pflag"

	"github.com/fpunkt/zlog"
	"github.com/rs/zerolog/log"
)

var options = struct {
	verbose int
}{}

func main() {
	pflag.CountVarP(&options.verbose, "verbose", "v", "verbose messages")
	pflag.Parse()
	zlog.InitL(options.verbose)
	logmessages("color")

	log.Logger = zlog.New(zlog.Options{
		Level:      options.verbose,
		Format:     zlog.FormatUnicode,
		TimeFormat: "s",
	})
	logmessages("unicode with seconds")

	log.Logger = zlog.New(zlog.Options{
		Level:      options.verbose,
		Format:     zlog.FormatBW,
		TimeFormat: "highres",
	})
	logmessages("BW")

	zlog.Init()
	logmessages("Defaults")
}

func logmessages(msg string) {
	log.Info().Int("level", options.verbose).Str("Setting", msg).Msg("A bunch of messages")
	log.Log().Msg("This is a log message")
	log.Trace().Msg("This is a trace message")
	log.Debug().Msg("This is a debug message")
	log.Info().Msg("This is a Info message")
	log.Warn().Msg("This is a Warning message")
	log.Error().Msg("This is a Error message")
}
