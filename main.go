package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	healthy int32
)

type key int

const (
	requestIDKey key = 0
)

func main() {
	var (
		username   string
		passwd     string
		uri        string
		database   string
		listenAddr string
	)

	flag.StringVar(&username, "u", os.Getenv("MYSQL_USER"), "username")
	flag.StringVar(&passwd, "p", os.Getenv("MYSQL_PASSWORD"), "password")
	flag.StringVar(&uri, "uri", os.Getenv("MYSQL_URL"), "uri")
	flag.StringVar(&database, "db", os.Getenv("MYSQL_DATABASE"), "database")
	flag.StringVar(&listenAddr, "listen-addr", ":5000", "server listen address")

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)

	done := make(chan bool)
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	router := http.NewServeMux()
	router.Handle("/", index())
	router.Handle("/healthz", healthz())

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	ctx, _ := context.WithCancel(context.Background())
	go func() {
		select {
		case sig := <-c:
			logger.Printf("Got %s signal. Aborting...\n", sig)
			logger.Println("Server is shutting down...")
			atomic.StoreInt32(&healthy, 0)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			server.SetKeepAlivesEnabled(false)
			if err := server.Shutdown(ctx); err != nil {
				logger.Printf("Could not gracefully shutdown the server: %v\n", err)
			}
			close(done)

			cancel()
		}
	}()

	mysql, err := NewMysql(username, passwd, uri, database)
	if err != nil {
		logger.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	logger.Printf("[INFO] %s:%s@tcp(%s)/%s?parseTime=True\n", username, "*****", uri, database)

	vmstat := NewVmstat(mysql, 1)

	vmstat.wg.Add(1)
	go vmstat.Run(ctx)

	logger.Println("Server is starting...")

	logger.Println("Server is ready to handle requests at", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
	vmstat.wg.Wait()

	os.Exit(1)

}

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, World!")
	})
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
