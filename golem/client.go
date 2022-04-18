package golem

import (
	"encoding/json"
	"errors"
	"log"
	"runtime/debug"
)

const (
	LevelFatal = iota
	LevelError
	LevelInfo
	CodeFatal        = 500
	CodePanic        = 555
	TextUnknownPanic = "unknown panic"
)

var ErrNilPublisher = errors.New("publisher is not set")

type Log struct {
	Level   uint32 `json:"level"`
	Project string `json:"project"`
	Service string `json:"service"`
	Code    uint32 `json:"code"`
	Text    string `json:"text"`
}

func Info(text string, code uint32) error {
	return send(text, LevelInfo, code)
}

func Error(text string, code uint32) error {
	return send(text, LevelError, code)
}

func Fatal(text string) error {
	return send(text, LevelFatal, CodeFatal)
}

func send(text string, level, code uint32) error {
	if publisher == nil {
		return ErrNilPublisher
	}

	msg := Log{
		Level:   level,
		Project: publisher.project,
		Service: publisher.service,
		Text:    text,
		Code:    code,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return publisher.Publish(data)
}

func Recover() {
	if r := recover(); r != nil {
		var text string

		switch err := r.(type) {
		case string:
			text = err
		case error:
			text = err.Error()
		default:
			log.Println(TextUnknownPanic, err)
			text = TextUnknownPanic
		}

		_ = send(text+":\n"+string(debug.Stack()), LevelFatal, CodePanic)
	}
}
