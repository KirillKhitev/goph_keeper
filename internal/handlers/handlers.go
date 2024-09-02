package handlers

import (
	"github.com/KirillKhitev/goph_keeper/internal/api"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"log"
	"net/http"
)

// Тип хендлер.
type Handler struct {
	Store store.Store
}

// Хендлер регистрации.
type Register Handler

// Обработка запросов регистрации.
func (ch *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := api.RegisterUser(w, r, ch.Store)

	sendResponse(response, w)
}

// Хендлер авторизации.
type Login Handler

// Обработка запросов авторизации.
func (ch *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := api.Login(w, r, ch.Store)

	sendResponse(response, w)
}

// Хендлер обновления данных в хранилище.
type Update Handler

// Обработка запросов обновления данных.
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

// Хендлер получения списка записей.
type List Handler

// Обработка запросов списка записей пользователя.
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

// Хендлер получения информации записи.
type Get Handler

// Обработка запросов информации записи.
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

// Отправка ответа сервера агенту.
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
