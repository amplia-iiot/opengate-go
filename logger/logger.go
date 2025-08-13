package logger

import (
	"fmt"
	"log"
)

// para el control de la jerarquia de trazas. No se usa el objeto traceLevel porque tiene espacios en WARN e INFO
type LogLevel string

var (
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
)

var logHierarchy map[LogLevel]int = map[LogLevel]int{
	ERROR: 0,
	WARN:  1,
	INFO:  2,
	DEBUG: 3,
}

type LoggerI interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Error(v ...interface{})
	Warn(v ...interface{})
}

var writter LoggerI

func NewConf(w LoggerI) {
	writter = w
}
func Info(v ...interface{}) {
	var info2print string
	for _, e := range v {
		info2print = info2print + fmt.Sprintf("%v", e)
	}
	if writter != nil {
		writter.Info(info2print)
	} else {
		log.Println(info2print)
	}
}
func Debug(v ...interface{}) {
	var info2print string
	for _, e := range v {
		info2print = info2print + fmt.Sprintf("%v", e)
	}
	if writter != nil {
		writter.Debug(info2print)
	} else {
		log.Println(info2print)
	}
}

func Error(v ...interface{}) {
	var info2print string
	for _, e := range v {
		info2print = info2print + fmt.Sprintf("%v", e)
	}
	if writter != nil {
		writter.Error(info2print)
	} else {
		log.Println(info2print)
	}
}
func Warn(v ...interface{}) {
	var info2print string
	for _, e := range v {
		info2print = info2print + fmt.Sprintf("%v", e)
	}
	if writter != nil {
		writter.Warn(info2print)
	} else {
		log.Println(info2print)
	}
}
