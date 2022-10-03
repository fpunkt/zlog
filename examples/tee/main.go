package main

import (
	"fmt"

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

	log.Logger = zlog.Tee("log.json", zlog.Options{Format: zlog.FormatJson, Overwrite: true})
	log.Info().Msg("ok")

	err := zlog.NewError("Test Message").Int("n", 3).Str("file", "myfile.go")

	err.Logger().Warn().Str("message", err.Message).Msg("This was an error")
	for i := 0; i < 5; i++ {
		log.Debug().Int("i", i).Str("msg", fmt.Sprintf("msg #%d", i)).Msg("debug message")
	}
}
