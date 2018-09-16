package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		username string
		passwd   string
		uri      string
		database string
	)

	flag.StringVar(&username, "u", os.Getenv("MYSQL_USER"), "username")
	flag.StringVar(&passwd, "p", os.Getenv("MYSQL_PASSWORD"), "password")
	flag.StringVar(&uri, "uri", os.Getenv("MYSQL_URL"), "uri")
	flag.StringVar(&database, "db", os.Getenv("MYSQL_DATABASE"), "database")

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case sig := <-c:
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			cancel()
		}
	}()

	mysql, err := NewMysql(username, passwd, uri, database)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("[INFO] %s:%s@tcp(%s)/%s?parseTime=True\n", username, "*****", uri, database)

	vmstat := NewVmstat(mysql, 1)

	vmstat.wg.Add(1)

	vmstat.Run(ctx)

	vmstat.wg.Wait()

	os.Exit(1)
}
