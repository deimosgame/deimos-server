package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

const (
	colorBlack = (iota + 30)
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite
	colorGray
)

var (
	colors = map[string]string{
		"NOCOLOR": "",
		"ERROR":   colorSeq(colorRed),
		"WARN":    colorSeq(colorYellow),
		"NOTICE":  colorSeq(colorGreen),
		"DEBUG":   colorSeq(colorCyan),
		"INFO":    colorSeq(colorGray),
	}
)

// InitLogging creates a logger and returns it for further usage
func InitLogging(logfile string) *Logger {
	stdLogger := log.New(os.Stdout, "", log.Ltime)
	errLogger := log.New(os.Stderr, "", log.Ltime)

	// File opening flags for creation or appending
	openFlags := os.O_RDWR | os.O_APPEND
	if _, err := os.Stat(logfile); os.IsNotExist(err) {
		openFlags = os.O_CREATE
	} else if err != nil {
		panic("Error accessing log file!")
	}

	// Actual file opening
	file, err := os.OpenFile(logfile, openFlags, 0660)
	if err != nil {
		panic("Couldn't write to log file!")
	}
	fileLogger := log.New(file, "", log.LstdFlags)

	return &Logger{
		stdLogger:  stdLogger,
		errLogger:  errLogger,
		fileLogger: fileLogger,
		ToFile:     false,
	}
}

// Logger is a wrap-up type arround standard logger to add stuff like colors
type Logger struct {
	stdLogger  *log.Logger
	errLogger  *log.Logger
	fileLogger *log.Logger
	DebugMode  bool
	ToFile     bool
}

// Println prints the colored version of a string to a logger
func Println(out *log.Logger, color string, str ...string) {
	outBuffer := bytes.NewBuffer(nil)
	if color != "" {
		outBuffer.WriteString(color)
	}
	for _, v := range str {
		outBuffer.WriteString(v)
	}
	if color != "" {
		outBuffer.WriteString("\033[0m")
	}
	out.Println(outBuffer.String())
}

/**
 *  Custom loggers outputs
 *
 *  Currently ERROR, WARN and NOTICE are logging to stdErr + log file
 *  The others are logged into stdOut, unless user set Logger.ToFile = true
 */

// toStdOut writes a log line to a Logger.stdLogger, and possibly to a file
func (l *Logger) toStdOut(color string, str ...string) {
	Println(l.stdLogger, colors[color], str...)
	if l.ToFile {
		l.toLogFile(color, str...)
	}
}

// toStdErr writes a log line to a Logger.errLogger and to the log file
func (l *Logger) toStdErr(color string, str ...string) {
	Println(l.errLogger, colors[color], str...)
	l.toLogFile(color, str...)
}

// toLogFile writes directly a log line to the log file
func (l *Logger) toLogFile(color string, str ...string) {
	argsSlice := make([]string, len(str)+1)
	argsSlice[0] = "[" + color + "] "
	for i, v := range str {
		argsSlice[i+1] = v
	}
	Println(l.fileLogger, "", argsSlice...)
}

// Small helper functions for color

func (l *Logger) Info(str ...string) {
	l.toStdOut("INFO", str...)
}

func (l *Logger) Error(str ...string) {
	l.toStdErr("ERROR", str...)
}

func (l *Logger) Panic(str ...string) {
	l.Error(str...)
	os.Exit(1)
}

func (l *Logger) Warn(str ...string) {
	l.toStdErr("WARN", str...)
}

func (l *Logger) Notice(str ...string) {
	l.toStdErr("NOTICE", str...)
}

func (l *Logger) Debug(str ...string) {
	if l.DebugMode {
		l.toStdOut("DEBUG", str...)
	}
}

// colorSeq generates the coloring string for terminal based on a color int
func colorSeq(color int) string {
	return fmt.Sprintf("\033[%dm", color)
}
