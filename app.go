package main

import (
    "fmt"
    "log"
    "database/sql"
    "github.com/gorilla/mux"
    _"github.com/lib/pq"
)

type App struct {
    Router *mux.Router
    DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname, sslmode string) {
    connectionString :=
        fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbname, sslmode)
    fmt.Println(connectionString)

    var err error
    a.DB, err = sql.Open("postgres", connectionString)
    if err != nil {
        log.Fatal(err)
    }

    a.Router = mux.NewRouter()
}

func (a *App) Run(addr string) { }
