package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/zrtgzrtg/gatorcli/internal/config"
	"github.com/zrtgzrtg/gatorcli/internal/database"
)

func main() {
	confi, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	//open Db connection

	db, err := sql.Open("postgres", confi.Db_url)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	statePtr := &state{dbQueries, &confi}

	commands := commands{
		cMap: make(map[string]func(*state, command) error),
	}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)

	args := os.Args
	if len(args) < 2 {
		log.Fatal("not enough arguments given!")
	}
	commandName := args[1]
	if len(args) > 2 {
		args = args[2:]
	} else {
		args = []string{}
	}

	err = commands.run(statePtr, command{commandName, args})
	if err != nil {
		log.Fatal(err)
	}

}
