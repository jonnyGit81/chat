package main

import "time"

// message represents a single message
type message struct {
	Name      string
	Message   string // auto encapsulated on readJson
	When      time.Time
	AvatarURL string
}
