# Initialization for zerolog console output

Initialization for zerolog console loggers. Especially fixes colors
for terminals with dark background, where colors are not readable
with the zerolog default colors.

It will also allow colored output under windows

Usage:

```go
var log = zlog.New()
// 2021-04-17 14:30:57 INF Creating file file=hosts

var log = zlog.New(zlog.Options{TimeFormat: "s"})
// [0000] INF Creating file file=hosts

var log = zlog.New(zlog.Options{})
// INF Creating file file=hosts



--> 2022-02-06 12:34:56 This is a log message
--> 2022-02-06 12:34:56 TRC This is a trace message
--> 2022-02-06 12:34:56 DBG This is a debug message
--> 2022-02-06 12:34:56 INF This is a Info message
--> 2022-02-06 12:34:56 WRN This is a Warning message
--> 2022-02-06 12:34:56 ERR This is a Error message

log.Logger = zlog.New(zlog.Options{
	Level:      options.verbose,
	Format:     zlog.FormatUnicode,
	TimeFormat: "s",
})
--> [0000] 🔷 This is a trace message
--> [0000] 🔵 This is a debug message
--> [0000] 🟢 This is a Info message
--> [0000] 🔶 This is a Warning message
--> [0000] ❌ This is a Error message


log.Logger = zlog.New(zlog.Options{
	Level:      options.verbose,
	Format:     zlog.FormatBW,
	TimeFormat: "highres",
})
--> 2022-02-06 12:34:56.170 TRC This is a trace message
--> 2022-02-06 12:34:56.170 DBG This is a debug message
--> 2022-02-06 12:34:56.170 INF This is a Info message
--> 2022-02-06 12:34:56.170 WRN This is a Warning message
--> 2022-02-06 12:34:56.170 ERR This is a Error message

log.Logger = zlog.New()
logmessages("Defaults")
```
