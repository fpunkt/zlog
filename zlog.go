package zlog

// Initialization for zerolog console loggers. Especially fixes colors
// for terminals with dark background, where colors are not readable
// with the zerolog default colors
//
// Colored output under windows is supported.
//
// Note that Debug and Trace levels are treated slightly different that in zerolog to allow
// convenient init with pflag/cobra CountVar:
//  level 0 corresponds to Info
//  level 1 corresponds to Debug
//  level 2 corresponds to Trace
//
// Function Tee() copies logging output to a file (using JSON format, use logdump to reformat)
//
// Usage:
/*

package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/fpunkt/zlog"
)

func main() {
	pflag.CountVarP(&options.verbose, "verbose", "v", "verbose messages")
	pflag.Parse()
	zlog.InitV(options.Verbose)

	log.Info().Msg("Info message")
	log.Debug().Msg("Debug message")
}

*/

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	SupportColors = true
)

// Init new console logger with given level. This will set the global zerolog.log.Logger.
// Use zlog.New() if you need more flexibility
func InitL(level int) { log.Logger = New(Options{Level: level}) }

// Init new console logger with level 0. This will set the global zerolog.log.Logger
// Use zlog.New() if you need more flexibility
func Init() { InitL(0) }

// Options that can be used to initialize the logger
type Options struct {
	// The level for the logger. NOTE: the convention here is 0 is info-level, positive numbers increase the verbosity,
	// negative numbers decrease (e.g. -1 is warnings only, -2 errors only.). This allows
	// convenient use of a verbose level commandline argument with pflag.CountVar()
	Level int

	// Timeformat for the output. This can be one of
	//          "s": use relative time in seconds as timestamp, format is like [0004] (for 4 seconds)
	//  	 "none": supress time field in output
	//    "default": Use "2006-01-02 15:04:05"
	//           "": (empty string): same as default
	//    "highres": Use "2006-01-02 15:04:05.000"
	//        other: Any golang time format string
	TimeFormat string

	// This is the output format used with the Tee logger. Default is the same as the console logger
	// but you might want to set this to Format: zlog.FormatJson
	Format LogOutputFormat // used in Tee

	// Option for Tee logger, whether any existing logfile is overwritten. Default is to append to
	// an existing logfile
	Overwrite bool // used in Tee
}

type LogOutputFormat = int

const (
	FormatColor LogOutputFormat = iota
	FormatBW
	FormatJson
	FormatUnicode
)

const (
	// see
	// https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html
	// for color codes
	// check https://en.wikipedia.org/wiki/ANSI_escape_code for nice colors
	// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797

	// 	ESC[ 38:5:⟨n⟩ m Select foreground color
	// ESC[ 48:5:⟨n⟩ m Select background color

	_intro = "\033[38;5;"
	// RedColor resets the color to default
	ResetColor = "\033[0m"
	// Yellow is the escape sequence to select Yellow color
	Yellow = _intro + "226m"
	// Orange is the escape sequence to select Orange color
	Orange = _intro + "208m"
	// Gray is the escape sequence to select Gray color
	Gray = _intro + "246m"
	// Green is the escape sequence to select Green color
	Green = _intro + "10m"
	// Red is the escape sequence to select Red color
	Red = _intro + "196m"
	// Cyan is the escape sequence to select Cyan color
	Cyan = _intro + "45m"
	// Magenta is the escape sequence to select Magenta color
	Magenta = _intro + "207m"
	// Blue is the escape sequence to select Blue color
	Blue = _intro + "33m"
)

// Prefix control sequence to string to colorize the output. Color-reset sequence is appended to the end of the string.
func Colorize(colorcontrolsequence, text string) string {
	return colorcontrolsequence + text + ResetColor
}

var colormap = map[string]string{
	"yellow":  Yellow,
	"orange":  Orange,
	"gray":    Gray,
	"green":   Green,
	"red":     Red,
	"cyan":    Cyan,
	"magenta": Magenta,
	"blue":    Blue,
}

// NamedColorize will add control sequences to for color output of the given string
func NamedColorize(colorname, text string) string {
	return colormap[colorname] + text + ResetColor
}

func formatLevelUnicode(i interface{}) string {
	if ll, ok := i.(string); ok {
		switch ll {
		case "trace":
			return "🔹" // 🔷🔹
		case "debug":
			return "🔷" // ◻ 🔳⚪🔲🔵🟪
		case "info":
			return "🟢" // ✅❎🔷
		case "warn":
			return "🔶"
		case "error":
			return "❌" // "🔻"
		case "fatal":
			return "❌"
		case "panic":
			return "❌"
		case "log":
			return "LOG"
		case "":
			return "nolevel"
		default:
			fmt.Printf("ups, unexpeced level %q\n", ll)
		}
	}
	if i == nil {
		return ""
	}
	return strings.ToUpper(fmt.Sprintf("%s", i))

}

// Hard coded switch to avoid mallocs, especially for the colored version
func formatLevelBW(i interface{}) string {
	if ll, ok := i.(string); ok {
		switch ll {
		case "trace":
			return "TRC"
		case "debug":
			return "DBG"
		case "info":
			return "INF"
		case "warn":
			return "WRN"
		case "error":
			return "ERR"
		case "fatal":
			return "FTL"
		case "panic":
			return "PNC"
		case "log":
			return "LOG"
		case "":
			return "nolevel"
		default:
			fmt.Printf("ups, unexpeced level %q\n", ll)
		}
	}
	if i == nil {
		return ""
	}
	return strings.ToUpper(fmt.Sprintf("%s", i))
}

func formatLevelColor(i interface{}) string {
	if !SupportColors {
		return formatLevelBW(i)
	}
	if ll, ok := i.(string); ok {
		switch ll {
		case "trace":
			return Gray + "TRC" + ResetColor
		case "debug":
			return Gray + "DBG" + ResetColor
		case "info":
			return Green + "INF" + ResetColor
		case "warn":
			return Orange + "WRN" + ResetColor
		case "error":
			return Red + "ERR" + ResetColor
		case "fatal":
			return Red + "FTL" + ResetColor
		case "panic":
			return Red + "PNC" + ResetColor
		case "log":
			return "LOG"
		case "":
			return "nolevel"
		default:
			fmt.Printf("ups, unexpeced level %q\n", ll)
		}
	}
	if i == nil {
		return ""
	}
	return strings.ToUpper(fmt.Sprintf("%s", i))
}

func getFormatter(format LogOutputFormat) func(interface{}) string {
	switch format {
	case FormatBW:
		return formatLevelBW
	case FormatUnicode:
		return formatLevelUnicode
	case FormatColor:
		return formatLevelColor
	default:
		return formatLevelBW
	}
}

var zerologStartup = time.Now()

// Defines how many stack frames are dropped from the stack traces.
var ZlogDropStack = 8

// Returns nil or a short stack dump like "main.go:68 | proc.go:225 | asm_amd64.s:1371"
func ZMarshalStack(err error) interface{} {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	sterr, ok := err.(stackTracer)
	if !ok {
		return nil
	}
	st := sterr.StackTrace()
	nlevels := len(st)
	if nlevels > ZlogDropStack {
		nlevels -= ZlogDropStack
	}
	b := []byte{}
	for i, frame := range st {
		if i > nlevels {
			return string(b)
		}
		pc := uintptr(frame) - 1
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			// huch..
			return "Internal zerologformat.306: cannot get fn"
		}
		if i > 0 {
			b = append(b, []byte(" | ")...)
		}
		file, line := fn.FileLine(pc)
		b = append(b, []byte(path.Base(file))...)
		b = append(b, []byte(":")...)
		b = append(b, strconv.Itoa(line)...)
	}
	return string(b)
}

func zconsoleWriter(o Options) zerolog.ConsoleWriter {
	zlogOptions = o
	var timestampFormat zerolog.Formatter
	switch o.TimeFormat {
	case "s":
		timestampFormat = func(i interface{}) string { return fmt.Sprintf("[%04d]", time.Since(zerologStartup)/time.Second) }
		o.TimeFormat = zerolog.TimeFormatUnix
	case "none":
		timestampFormat = func(i interface{}) string { return "" }
	case "", "default":
		//o.TimeFormat = "2006-01-02 15:04:05"
		// for some strange reason some programs dump the timestamp with +2h offset
		// when the standard formatter is used.
		// (First seen when using the NATS library, but not clear whether there is a correlation)
		timestampFormat = func(i interface{}) string {
			//fmt.Printf("i = %t / %v\n", i, i) // unix time in seconds (probably), a json.number
			return time.Now().Format("2006-01-02 15:04:05")
		}
	case "highres":
		o.TimeFormat = "2006-01-02 15:04:05.000"
		timestampFormat = func(i interface{}) string {
			return time.Now().Format("2006-01-02 15:04:05.000")
		}
	default:
		//panic(fmt.Sprintf("Bad timeformat %q", o.TimeFormat))
		// provided by user as regular golang timeformat template
	}
	zerolog.ErrorStackMarshaler = ZMarshalStack
	zerolog.TimeFieldFormat = o.TimeFormat
	zerolog.TimestampFieldName = "_zts"
	zerolog.LevelFieldName = "_zl"
	zerolog.MessageFieldName = "_zm"

	//o.PartsOrder = nil

	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: o.TimeFormat}
	output.FormatLevel = getFormatter(o.Format)

	// patch colors to be more readable
	if (o.Format == FormatColor || o.Format == FormatUnicode) && SupportColors {
		output.FormatFieldName = func(i interface{}) string {
			return Cyan + fmt.Sprint(i) + "=" + ResetColor
		}

		// use red color for error messages
		output.FormatErrFieldName = func(i interface{}) string {
			return Red + "error=" + ResetColor
		}
		output.FormatErrFieldValue = func(i interface{}) string {
			//return fmt.Sprint(i) + ResetColor
			return fmt.Sprint(i)
		}
	} else {
		output.FormatFieldName = func(i interface{}) string { return fmt.Sprint(i) + "=" }
		output.FormatErrFieldName = func(i interface{}) string { return "error=" }
		output.FormatErrFieldValue = func(i interface{}) string { return fmt.Sprint(i) }
	}

	if timestampFormat != nil {
		output.FormatTimestamp = timestampFormat
	}
	return output
}

// Store last options here for tlog (needs to create new loggers with Tee and others)
var zlogOptions Options

// Returns a new zerolog console logger instance with given options
func New(o Options) zerolog.Logger {
	output := zconsoleWriter(o)
	zlog := zerolog.New(output).With()
	if o.TimeFormat != "none" {
		zlog = zlog.Timestamp()
	}
	return setlevel(zlog.Logger(), o.Level)
}

var _zlog = zerolog.Nop()

// DisabledLogger is a logger that will never output anything
var DisabledLogger = _zlog

var loglevel int

// Return a new logger with given level Logl
func setlevel(logger zerolog.Logger, level int) zerolog.Logger {
	loglevel = level
	if level < -3 {
		level = -3
	}
	switch level {
	case -3:
		return logger.Level(zerolog.FatalLevel)
	case -2:
		return logger.Level(zerolog.ErrorLevel)
	case -1:
		return logger.Level(zerolog.WarnLevel)
	case 0:
		return logger.Level(zerolog.InfoLevel)
	case 1:
		return logger.Level(zerolog.DebugLevel)
	default:
		return logger.Level(zerolog.TraceLevel)
	}
}

// SetLevel defines the minimum log level for the global Logger
func SetLevel(level int) { log.Logger = setlevel(log.Logger, level) }

// Logl returns a disable logger if level > loglevel that has been set with SetLevel()
func Logl(level int) *zerolog.Event {
	//fmt.Printf("zlog(%d), v=%d -> %t\n", level, Options.Verbose, level > Options.Verbose)
	if level > loglevel {
		return DisabledLogger.Trace()
	}
	return log.Trace()
}

// Tee duplicates logging output to given file
func Tee(fname string, options ...Options) zerolog.Logger {
	var o Options
	if len(options) == 0 {
		o = Options{
			Overwrite: false,
			Format:    FormatBW,
		}
	} else {
		o = options[0]
	}
	var flag int = os.O_CREATE | os.O_WRONLY
	if o.Overwrite {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	fd, err := os.OpenFile(fname, flag, 0666)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot tee output")
	}
	console := zconsoleWriter(zlogOptions)

	var multi zerolog.LevelWriter
	// TODO: could share code with New()?
	switch o.Format {
	case FormatJson:
		multi = zerolog.MultiLevelWriter(console, fd)
	default:
		sc := SupportColors
		if o.Format == FormatBW {
			SupportColors = false
		}
		file := zconsoleWriter(zlogOptions)
		SupportColors = sc
		file.Out = fd
		file.FormatLevel = formatLevelBW
		if o.Format == FormatBW {
			file.NoColor = true
		}
		multi = zerolog.MultiLevelWriter(console, file)
	}

	m := zerolog.New(multi).With().Timestamp().Logger()
	// restore level

	return setlevel(m, loglevel)
	//SetLevel(loglevel)
	//return m
}

// Provide errors that can be returned as standard golang errors.
// To convert this back to a zerolog error use AsZerologError(). See NewError() for more info.
type Error struct {
	Message string
	C       zerolog.Context
	augment int
}

func AsZerologError(e error) (*zerolog.Logger, string) {
	if ee, ok := e.(*Error); ok {
		l := ee.C.Logger()
		return &l, ee.Message
	}
	return nil, ""
}

func (e *Error) Logger() *zerolog.Logger {
	l := e.C.Logger()
	return &l
}

// Augment error by another error
func (e *Error) Augment(s string) *Error {
	e.augment++
	e.C = e.C.Str(fmt.Sprintf("nested#%d", e.augment), s)
	return e
}

func (e *Error) Str(name, value string) *Error {
	e.C = e.C.Str(name, value)
	return e
}

func (e *Error) Int(name string, value int) *Error {
	e.C = e.C.Int(name, value)
	return e
}

func (e *Error) Err(err error) *Error {
	e.C = e.C.Str("nested", err.Error())
	return e
}

func (e *Error) Error() string {
	var buf bytes.Buffer
	output := zconsoleWriter(Options{TimeFormat: "none"})
	output.Out = &buf
	l := e.C.Logger().Output(output)
	l.Log().Msg(e.Message)
	return strings.TrimSpace(buf.String())
}

func NewError(msg string) *Error {
	return &Error{
		Message: msg,
		C:       log.With(),
	}
}
