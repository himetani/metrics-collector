package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/himetani/metrics-collector/stat"
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

	logger := log.New(os.Stdout, "Info: ", log.LstdFlags)

	ctx, _ := context.WithCancel(context.Background())
	go func() {
		select {
		case sig := <-c:
			logger.Printf("Got %s signal. Aborting...\n", sig)

			_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
		}
	}()

	mysql, err := stat.NewMysql(username, passwd, uri, database)
	if err != nil {
		logger.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	logger.Printf("%s:%s@tcp(%s)/%s?parseTime=True\n", username, "*****", uri, database)

	vmstat := stat.NewVmstat(mysql, 1)

	vmstat.Add(1)
	vmstat.Run(ctx)

	vmstat.Wait()

	os.Exit(1)
}
