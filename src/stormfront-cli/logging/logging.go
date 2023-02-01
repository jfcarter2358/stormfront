package logging

import (
	"fmt"
	"os"
	"stormfront-cli/ansi"
)

const (
	NONE_NAME    string = "NONE"
	FATAL_NAME   string = "FATAL"
	SUCCESS_NAME string = "SUCCESS"
	ERROR_NAME   string = "ERROR"
	WARN_NAME    string = "WARN"
	INFO_NAME    string = "INFO"
	DEBUG_NAME   string = "DEBUG"
	TRACE_NAME   string = "TRACE"
)

const (
	NONE_LEVEL    int = -1
	FATAL_LEVEL   int = 0
	ERROR_LEVEL   int = 1
	SUCCESS_LEVEL int = 2
	WARN_LEVEL    int = 3
	INFO_LEVEL    int = 4
	DEBUG_LEVEL   int = 5
	TRACE_LEVEL   int = 6
)

var Level = ERROR_LEVEL

func GetDefaults() string {
	return fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s", NONE_NAME, FATAL_NAME, SUCCESS_NAME, ERROR_NAME, WARN_NAME, INFO_NAME, DEBUG_NAME, TRACE_NAME)
}

func SetLevel(level string) error {
	switch level {
	case NONE_NAME:
		Level = NONE_LEVEL
	case FATAL_NAME:
		Level = FATAL_LEVEL
	case SUCCESS_NAME:
		Level = SUCCESS_LEVEL
	case ERROR_NAME:
		Level = ERROR_LEVEL
	case WARN_NAME:
		Level = WARN_LEVEL
	case INFO_NAME:
		Level = INFO_LEVEL
	case DEBUG_NAME:
		Level = DEBUG_LEVEL
	case TRACE_NAME:
		Level = TRACE_LEVEL
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	return nil
}

func Fatal(message string) {
	if Level >= FATAL_LEVEL {
		fmt.Printf("%s[%s]%s :: %s\n", ansi.RED, FATAL_NAME, ansi.NO_COLOR, message)
	}
	os.Exit(1)
}

func Success(message string) {
	if Level >= SUCCESS_LEVEL {
		fmt.Printf("%s[%s]%s :: %s\n", ansi.GREEN, SUCCESS_NAME, ansi.NO_COLOR, message)
	}
}

func Error(message string) {
	if Level >= ERROR_LEVEL {
		fmt.Printf("%s[%s]%s :: %s\n", ansi.RED, ERROR_NAME, ansi.NO_COLOR, message)
	}
}

func Warn(message string) {
	if Level >= WARN_LEVEL {
		fmt.Printf("%s[%s]%s :: %s\n", ansi.YELLOW, WARN_NAME, ansi.NO_COLOR, message)
	}
}

func Info(message string) {
	if Level >= INFO_LEVEL {
		fmt.Printf("%s[%s]%s :: %s\n", ansi.GREEN, INFO_NAME, ansi.NO_COLOR, message)
	}
}

func Debug(message string) {
	if Level >= DEBUG_LEVEL {
		fmt.Printf("%s[%s]%s :: %s\n", ansi.CYAN, DEBUG_NAME, ansi.NO_COLOR, message)
	}
}

func Trace(message string) {
	if Level >= TRACE_LEVEL {
		fmt.Printf("%s[%s]%s :: %s\n", ansi.BLUE, TRACE_NAME, ansi.NO_COLOR, message)
	}
}
