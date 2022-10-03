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
	log.Info().Msg("ok")

	err := zlog.NewError("Test Message").Int("n", 3).Str("file", "myfile.go")

	fmt.Printf("err = %s\n", err.Error())

	if e, m := zlog.AsZerologError(err); e != nil {
		e.Error().Msg(m)
		e.Warn().Str("message", m).Msg("This was an error")
	}
}
