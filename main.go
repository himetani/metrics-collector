package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
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

	vmstat := &Vmstat{
		db:     &Mysql{},
		ticker: 1,
	}
	vmstat.wg.Add(1)

	vmstat.Run(ctx)

	vmstat.wg.Wait()

	os.Exit(1)
}
