package main

import (
	"fmt"
	"github.com/alexeynavarkin/technopark_db_proj/internal/app"
	"time"
)

func main() {
	fmt.Println("APP: delay 10s.")
	time.Sleep(10 * time.Second)
	fmt.Println("APP: Starting app.")
	app.Start()
}
