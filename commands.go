package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/zrtgzrtg/gatorcli/internal/config"
	"github.com/zrtgzrtg/gatorcli/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}
type command struct {
	name string
	args []string
}
type commands struct {
	cMap map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("no arguments found in login command!\n")
	}
	name := cmd.args[0]
	//check if user exists
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return err
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("User: %v has been set to config", name)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	val, ok := c.cMap[cmd.name]
	if !ok {
		return fmt.Errorf("command with name: %v not found!", cmd.name)
	}
	err := val(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cMap[name] = f
}
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("no name found in register command")
	}
	timeNow := time.Now()
	sqlTimeNowStructObj := sql.NullTime{timeNow, true}
	usrName := cmd.args[0]
	usrParams := database.CreateUserParams{uuid.New(), sqlTimeNowStructObj, sqlTimeNowStructObj, usrName}

	// check if err is nil --> if yes it already exists
	_, err := s.db.GetUser(context.Background(), usrName)
	if err == nil {
		fmt.Printf("user with name: %v already exists in db!\n", usrName)
		os.Exit(1)
	}
	usr, err := s.db.CreateUser(context.Background(), usrParams)
	if err != nil {
		return err
	}
	s.cfg.SetUser(usrName)
	fmt.Printf("User was created with data:\n ID: %v \n created_at: %v \n updated_at: %v\n name: %v\n", usr.ID, usr.CreatedAt, usr.UpdatedAt, usr.Name)
	return nil
}
func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
	return nil
}
func handlerUsers(s *state, cmd command) error {
	//get list of users out of db
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	printList := []string{}
	for _, usr := range users {
		if usr.Name == s.cfg.Current_user_name {
			printList = append(printList, fmt.Sprintf("  * %v (current)\n", usr.Name))
		} else {
			printList = append(printList, fmt.Sprintf("  * %v\n", usr.Name))
		}
	}
	for _, p := range printList {
		fmt.Print(p)
	}
	return nil
}
