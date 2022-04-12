package golem

import (
	"encoding/json"
	"errors"
)

const (
	LevelFatal = iota
	LevelError
	LevelInfo
)

var ErrNilPublisher = errors.New("publisher is not set")

type Log struct {
	Level   int    `json:"level"`
	Service string `json:"service"`
	Text    string `json:"text"`
}

func Info(text string) error {
	return send(LevelInfo, text)
}

func Error(text string) error {
	return send(LevelError, text)
}

func Fatal(text string) error {
	return send(LevelFatal, text)
}

func send(level int, text string) error {
	if publisher == nil {
		return ErrNilPublisher
	}

	msg := Log{
		Level:   level,
		Service: publisher.service,
		Text:    text,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return publisher.Publish(data)
}
