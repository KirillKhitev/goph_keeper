package handlers

import (
	"github.com/KirillKhitev/goph_keeper/internal/api"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"log"
	"net/http"
)

type Handler struct {
	Store store.Store
}

type Register Handler

func (ch *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := api.RegisterUser(w, r, ch.Store)

	sendResponse(response, w)
}

type Login Handler

func (ch *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := api.Login(w, r, ch.Store)

	sendResponse(response, w)
}

type Update Handler

func (ch *Update) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, response := api.GetUserFromAuthHeader(w, r, ch.Store)
	if response.Code != 0 {
		sendResponse(response, w)
		return
	}

	response = api.UpdateData(w, r, userID, ch.Store)

	sendResponse(response, w)
}

type List Handler

func (ch *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, response := api.GetUserFromAuthHeader(w, r, ch.Store)
	if response.Code != 0 {
		sendResponse(response, w)
		return
	}

	response = api.ListData(w, r, userID, ch.Store)

	sendResponse(response, w)
}

type Get Handler

func (ch *Get) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, response := api.GetUserFromAuthHeader(w, r, ch.Store)
	if response.Code != 0 {
		sendResponse(response, w)
		return
	}

	response = api.GetData(w, r, userID, ch.Store)

	sendResponse(response, w)
}

func sendResponse(res api.ResponseType, writer http.ResponseWriter) {
	if len(res.LogMsg) > 0 {
		log.Println(res.LogMsg)
	}

	if res.Code > 0 {
		writer.WriteHeader(res.Code)
	}

	if len(res.Body) > 0 {
		writer.Write(res.Body)
	}
}
