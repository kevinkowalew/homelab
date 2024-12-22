package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Logger struct {
	info, debug, err, warn *log.Logger
	fields                 []any
}

func New() *Logger {
	return &Logger{
		info:   log.New(os.Stdout, "", 0),
		debug:  log.New(os.Stdout, "", 0),
		warn:   log.New(os.Stdout, "", 0),
		err:    log.New(os.Stdout, "", 0),
		fields: make([]any, 0),
	}
}

func (l *Logger) WithFields(fields ...interface{}) *Logger {
	nl := New()
	nl.fields = append(l.fields, fields)
	return nl
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(msg, "INFO", l.info, fields)
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(msg, "DEBUG", l.debug, fields)
}

func (l *Logger) Error(err error, msg string, fields ...interface{}) {
	fields = append(fields, "error")
	fields = append(fields, err.Error())
	l.log(msg, "ERROR", l.err, fields)
}

func (l *Logger) log(msg, level string, il *log.Logger, fields ...interface{}) {
	date := time.Now().Format("2006-01-02 15:04:05")

	if len(l.fields) > 0 {
		combinedFields := fmt.Sprintf(
			"%s %s",
			createFieldsString(l.fields...),
			createFieldsString(fields))
		il.Printf("[%s] [%s] \"%s\" %s", date, level, msg, combinedFields)

	} else {
		il.Printf("[%s] [%s] \"%s\" %s", date, level, msg, createFieldsString(fields))
	}
}

func createFieldsString(fields ...interface{}) string {
	ffields := flattenArgs(fields)
	if len(ffields) == 0 {
		return ""
	} else if len(ffields)%2 != 0 {
		return "odd number of log fields, can't create pairs"
	} else {
		items := []string{}
		for i := 0; i < len(ffields); i += 2 {
			item := fmt.Sprintf(
				"%v=%v",
				ffields[i],
				addQuotesIfNeeded(ffields[i+1]),
			)
			items = append(items, item)
		}
		return strings.Join(items, " ")
	}
}

func addQuotesIfNeeded(s any) any {
	switch s.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", s)
	default:
		return s
	}
}

func flattenArgs(args ...interface{}) []interface{} {
	flattened := make([]interface{}, 0)

	for _, arg := range args {
		switch v := arg.(type) {
		case []interface{}:
			// If the argument is a slice, recursively flatten it.
			flattened = append(flattened, flattenArgs(v...)...)
		default:
			// Otherwise, append the argument to the flattened slice.
			flattened = append(flattened, arg)
		}
	}

	return flattened
}
