package api

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/errs"
	"github.com/KirillKhitev/goph_keeper/internal/mycrypto"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"net/http"
)

type ResponseType struct {
	LogMsg string
	Body   []byte
	Code   int
}

const RegisterErrPrefix = "Error by register new User"

type UserAuthBody struct {
	ID  string `json:"id,omitempty"`
	Msg string `json:"msg"`
	Key string `json:"key,omitempty"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request, s store.Store) ResponseType {
	requestData := auth.AuthorizingData{}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&requestData); err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable decode json - %v", RegisterErrPrefix, err),
			Code:   http.StatusBadRequest,
			Body:   prepareUserAuthBody("Ошибка в запросе", "", ""),
		}
	}

	if requestData.UserName == "" || requestData.Password == "" {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: empty required data", RegisterErrPrefix),
			Code:   http.StatusBadRequest,
			Body:   prepareUserAuthBody("Не передали логин или пароль!", "", ""),
		}
	}

	user, errCreateUser := s.CreateUser(r.Context(), requestData)
	if errCreateUser != nil {
		if errors.Is(errCreateUser, errs.ErrAlreadyExist) {
			return ResponseType{
				LogMsg: fmt.Sprintf("%s: user '%s' already exists!", RegisterErrPrefix, requestData.UserName),
				Code:   http.StatusConflict,
				Body:   prepareUserAuthBody("Данный пользователь уже зарегистрирован!", "", ""),
			}
		}

		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable create new User - %s", RegisterErrPrefix, errCreateUser),
			Code:   http.StatusInternalServerError,
		}
	}

	key, err := mycrypto.GenerateRandom(aes.BlockSize)
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: dont create privateKey file user %s - %s ", RegisterErrPrefix, user.ID, err),
			Code:   http.StatusInternalServerError,
		}
	}

	privateKey := base64.StdEncoding.EncodeToString(key)

	token, err := auth.BuildJWTString(user)
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable create auth token - %s ", RegisterErrPrefix, err),
			Code:   http.StatusInternalServerError,
		}
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.Header().Set("Content-Type", "application/json")

	return ResponseType{
		LogMsg: fmt.Sprintf("Успешно зарегистрировали и авторизовали нового пользователя '%s'\n", user.UserName),
		Code:   http.StatusOK,
		Body:   prepareUserAuthBody("Вы успешно зарегистрированы и авторизованы!", privateKey, user.ID),
	}
}

const LoginErrPrefix = "Error by login User"

func Login(w http.ResponseWriter, r *http.Request, s store.Store) ResponseType {
	requestData := auth.AuthorizingData{}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&requestData); err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable decode json - %v", LoginErrPrefix, err),
			Code:   http.StatusBadRequest,
			Body:   prepareUserAuthBody("Ошибка в запросе", "", ""),
		}
	}

	if requestData.UserName == "" || requestData.Password == "" {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: empty required data", LoginErrPrefix),
			Code:   http.StatusBadRequest,
			Body:   prepareUserAuthBody("Не передали логин или пароль!", "", ""),
		}
	}

	user, errFindUser := s.GetUserByUserName(r.Context(), requestData.UserName)

	if errFindUser != nil {
		if !errors.Is(errFindUser, errs.ErrNotFound) {
			return ResponseType{
				LogMsg: fmt.Sprintf("%s: unable find User - %s", LoginErrPrefix, errFindUser),
				Code:   http.StatusInternalServerError,
			}
		}

		return ResponseType{
			LogMsg: fmt.Sprintf("%s: not find User - %s", LoginErrPrefix, requestData.UserName),
			Code:   http.StatusUnauthorized,
			Body:   prepareUserAuthBody("Неправильные логин/пароль", "", ""),
		}
	}

	hash := requestData.GenerateHashPassword()
	if hash != user.HashPassword {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: wrong password for User - %s", LoginErrPrefix, requestData.UserName),
			Code:   http.StatusUnauthorized,
			Body:   prepareUserAuthBody("Неправильные логин/пароль", "", ""),
		}
	}

	token, err := auth.BuildJWTString(user)
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable create auth token - %s ", LoginErrPrefix, err),
			Code:   http.StatusInternalServerError,
		}
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.Header().Set("Content-Type", "application/json")

	return ResponseType{
		LogMsg: fmt.Sprintf("Успешно авторизовали пользователя '%s'\n", user.UserName),
		Code:   http.StatusOK,
		Body:   prepareUserAuthBody("Вы успешно авторизованы!", "", user.ID),
	}
}

func prepareUserAuthBody(s string, k string, id string) []byte {
	b := &UserAuthBody{
		ID:  id,
		Msg: s,
		Key: k,
	}

	r, _ := json.Marshal(b)

	return r
}

func GetUserFromAuthHeader(w http.ResponseWriter, r *http.Request, s store.Store) (string, ResponseType) {
	userID, err := auth.GetUserIDFromAuthHeader(r.Header.Get("Authorization"))
	if err != nil {
		return userID, ResponseType{
			LogMsg: fmt.Sprintf("error by Authorization - %v", err),
			Code:   http.StatusUnauthorized,
			Body:   []byte("Ошибка авторизации!"),
		}
	}

	user, err := s.GetUserByID(r.Context(), userID)
	if err != nil {
		return user.ID, ResponseType{
			LogMsg: fmt.Sprintf("unable find User - %s", err),
			Code:   http.StatusInternalServerError,
		}
	}

	return user.ID, ResponseType{}
}
