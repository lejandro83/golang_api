package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/jwt"
)

type App struct {
    Router *mux.Router
    DB     *sql.DB
}

var dummyPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAx4fm7dngEmOULNmAs1IGZ9Apfzh+BkaQ1dzkmbUgpcoghucE
DZRnAGd2aPyB6skGMXUytWQvNYav0WTR00wFtX1ohWTfv68HGXJ8QXCpyoSKSSFY
fuP9X36wBSkSX9J5DVgiuzD5VBdzUISSmapjKm+DcbRALjz6OUIPEWi1Tjl6p5RK
1w41qdbmt7E5/kGhKLDuT7+M83g4VWhgIvaAXtnhklDAggilPPa8ZJ1IFe31lNlr
k4DRk38nc6sEutdf3RL7QoH7FBusI7uXV03DC6dwN1kP4GE7bjJhcRb/7jYt7CQ9
/E9Exz3c0yAp0yrTg0Fwh+qxfH9dKwN52S7SBwIDAQABAoIBAQCaCs26K07WY5Jt
3a2Cw3y2gPrIgTCqX6hJs7O5ByEhXZ8nBwsWANBUe4vrGaajQHdLj5OKfsIDrOvn
2NI1MqflqeAbu/kR32q3tq8/Rl+PPiwUsW3E6Pcf1orGMSNCXxeducF2iySySzh3
nSIhCG5uwJDWI7a4+9KiieFgK1pt/Iv30q1SQS8IEntTfXYwANQrfKUVMmVF9aIK
6/WZE2yd5+q3wVVIJ6jsmTzoDCX6QQkkJICIYwCkglmVy5AeTckOVwcXL0jqw5Kf
5/soZJQwLEyBoQq7Kbpa26QHq+CJONetPP8Ssy8MJJXBT+u/bSseMb3Zsr5cr43e
DJOhwsThAoGBAPY6rPKl2NT/K7XfRCGm1sbWjUQyDShscwuWJ5+kD0yudnT/ZEJ1
M3+KS/iOOAoHDdEDi9crRvMl0UfNa8MAcDKHflzxg2jg/QI+fTBjPP5GOX0lkZ9g
z6VePoVoQw2gpPFVNPPTxKfk27tEzbaffvOLGBEih0Kb7HTINkW8rIlzAoGBAM9y
1yr+jvfS1cGFtNU+Gotoihw2eMKtIqR03Yn3n0PK1nVCDKqwdUqCypz4+ml6cxRK
J8+Pfdh7D+ZJd4LEG6Y4QRDLuv5OA700tUoSHxMSNn3q9As4+T3MUyYxWKvTeu3U
f2NWP9ePU0lV8ttk7YlpVRaPQmc1qwooBA/z/8AdAoGAW9x0HWqmRICWTBnpjyxx
QGlW9rQ9mHEtUotIaRSJ6K/F3cxSGUEkX1a3FRnp6kPLcckC6NlqdNgNBd6rb2rA
cPl/uSkZP42Als+9YMoFPU/xrrDPbUhu72EDrj3Bllnyb168jKLa4VBOccUvggxr
Dm08I1hgYgdN5huzs7y6GeUCgYEAj+AZJSOJ6o1aXS6rfV3mMRve9bQ9yt8jcKXw
5HhOCEmMtaSKfnOF1Ziih34Sxsb7O2428DiX0mV/YHtBnPsAJidL0SdLWIapBzeg
KHArByIRkwE6IvJvwpGMdaex1PIGhx5i/3VZL9qiq/ElT05PhIb+UXgoWMabCp84
OgxDK20CgYAeaFo8BdQ7FmVX2+EEejF+8xSge6WVLtkaon8bqcn6P0O8lLypoOhd
mJAYH8WU+UAy9pecUnDZj14LAGNVmYcse8HFX71MoshnvCTFEPVo4rZxIAGwMpeJ
5jgQ3slYLpqrGlcbLgUXBUgzEO684Wk/UV9DFPlHALVqCfXQ9dpJPg==
-----END RSA PRIVATE KEY-----`)

func (a *App) Initialize(user, password, dbname, sslmode string) {
    connectionString :=
        fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbname, sslmode)

    var err error
    a.DB, err = sql.Open("postgres", connectionString)
    if err != nil {
        log.Fatal(err)
    }

    a.Router = mux.NewRouter()
    a.initializeRoutes()
}

func (a *App) Run(addr string) {
    log.Fatal(http.ListenAndServe(":8000", a.Router))
}

func (a *App) initializeRoutes() {
    a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
    a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
    a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
    a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
    a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
    count, _ := strconv.Atoi(r.FormValue("count"))
    start, _ := strconv.Atoi(r.FormValue("start"))

    if count > 10 || count < 1 {
        count = 10
    }
    if start < 0 {
        start = 0
    }

    products, err := getProducts(a.DB, start, count)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, products)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
    var p product
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&p); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    defer r.Body.Close()

    if err := p.createProduct(a.DB); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid product ID")
        return
    }

    p := product{ID: id}
    if err := p.getProduct(a.DB); err != nil {
        switch err {
        case sql.ErrNoRows:
            respondWithError(w, http.StatusNotFound, "Product not found")
        default:
            respondWithError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }

    respondWithJSON(w, http.StatusOK, p)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid product ID")
        return
    }

    var p product
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&p); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
        return
    }
    defer r.Body.Close()
    p.ID = id

    if err := p.updateProduct(a.DB); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
        return
    }

    p := product{ID: id}
    if err := p.deleteProduct(a.DB); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}
