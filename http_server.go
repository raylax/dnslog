package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randomName() string {
	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func startHttpServer(db *db, domain string, stop chan bool) {
	if err := http.ListenAndServe(":80", newMux(db, domain)); err != nil {
		println("ERROR [HTTP] -" + err.Error())
		stop <- true
	}
}

func newMux(db *db, domain string) *mux.Router {
	handler := &httpHandler{
		db: db,
		domain: domain,
	}
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.Path("/new").HandlerFunc(handler.handleNew)
	api.Path("/records").HandlerFunc(handler.handleList)
	return r
}

type httpHandler struct {
	db *db
	domain string
}

func (h *httpHandler) handleNew(resp http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(resp).Encode(randomName() + "." + h.domain)
}

func (h *httpHandler) handleList(resp http.ResponseWriter, req *http.Request) {
	name := strings.TrimSuffix(req.URL.Query().Get("name"), "." + domain)
	_ = json.NewEncoder(resp).Encode(h.db.GetRecords(name))
}
