package main

import (
  "github.com/Lafriakh/env"
)

func main() {
    env.New()
    a := App {}
    a.Initialize(
        env.Get("APP_DB_USERNAME").(string),
        env.Get("APP_DB_PASSWORD").(string),
        env.Get("APP_DB_NAME").(string),
        env.Get("APP_DB_SSLMODE").(string))

    a.Run(":8000")
}
