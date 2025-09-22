package main

import "github.com/zrtgzrtg/gatorcli/internal/database"

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return nil
}
