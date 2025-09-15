package main

import (
	"fmt"
	"log"

	"github.com/zrtgzrtg/gatorcli/internal/config"
)

func main() {
	confi, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	err = confi.SetUser("zrtg")
	if err != nil {
		log.Fatal(err)
	}
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(conf.Db_url, conf.Current_user_name)

}
