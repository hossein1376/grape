package main

import (
	"log"
	"os"
)

// logger embeds standard log.Logger, and it also implements grape.Logger.
// Any other logging packages can be used as well.
type logger struct {
	*log.Logger
}

func newLogger() logger {
	return logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l logger) Debug(msg string, args ...any) {
	l.log("Debug", msg, args...)
}

func (l logger) Info(msg string, args ...any) {
	l.log("Info", msg, args...)
}

func (l logger) Warn(msg string, args ...any) {
	l.log("Warn", msg, args...)
}

func (l logger) Error(msg string, args ...any) {
	l.log("Error", msg, args...)
}

func (l logger) log(level, msg string, args ...any) {
	m := append([]any{level + ":", msg}, args...)
	l.Println(m...)
}
