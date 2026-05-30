package core

import "time"

type Game interface {
	Update(delta time.Duration) error
	Render() string
	HandleInput(key string)
	Name() string
	Description() string
	IsPaused() bool
}

type Position struct {
	X, Y int
}

type Config struct {
	Ghost      bool
	StartLevel int
}