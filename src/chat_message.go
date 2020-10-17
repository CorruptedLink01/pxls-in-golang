package main

import (
	"time"
)

type ChatMessage struct {
	ID              uint
	Author          string
	Date            *time.Time
	RawMessage      string
	Badges          []Badge
	AuthorNameClass []string
	AuthorNameColor uint
	//TODO(link) add faction
}

type Badge struct {
	DisplayName string
	Tooltip     string
	Type        string
	CSSIcon     string
}
