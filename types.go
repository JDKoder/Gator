package main

import (
	"fmt"

	"github.com/JDKoder/Gator/internal/config"
	"github.com/JDKoder/Gator/internal/database"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	val, ok := c.commands[cmd.name]
	if ok {
		return val(s, cmd)
	}
	return fmt.Errorf("command not found: %s", cmd.name)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}
