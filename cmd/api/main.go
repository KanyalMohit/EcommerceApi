package main

import (
	"LetsGoFurther/internal/data"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	/* flag.StringVar(&cfg.db.dsn,"db-dsn","postgres://greenlight:password@localhost/greenlight","PostgresSQL DSB") */
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgresSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns,"db-max-open-conns",25,"PostgresSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns,"db-max-idle-cons",25,"PostgresSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime,"db-max-idle-time","15m","PostgresSQL max connection idle time")

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Printf("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting %s server on %s", cfg.env, srv.Addr)

	err = srv.ListenAndServe()

	logger.Fatal(err)
}

// The `openDB` function in Go opens a connection to a PostgreSQL database using the provided
// configuration.
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration,err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil{
		return nil,err
	}

	db.SetConnMaxIdleTime(duration)


	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	//using PingContext() to establish a new connection to the database, passing in the context we created above as a parameter.
	//if the context couldn't be established succesfully withing 5 second deadline,then this will return an error
	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil

}
