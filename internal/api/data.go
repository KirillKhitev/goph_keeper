package api

import (
	"encoding/json"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"io"
	"net/http"
	"os"
	"time"
)

const ListDataErrPrefix = "Error by get list data"

func ListData(w http.ResponseWriter, r *http.Request, userID string, s store.Store) ResponseType {
	result, err := s.List(r.Context(), userID)
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable get list data for user [%s] - %v", ListDataErrPrefix, userID, err),
			Code:   http.StatusInternalServerError,
		}
	}

	body, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable marshal list data for user [%s] - %v", ListDataErrPrefix, userID, err),
			Code:   http.StatusInternalServerError,
		}
	}

	return ResponseType{
		LogMsg: fmt.Sprintf("get list data for user [%s]", userID),
		Code:   http.StatusOK,
		Body:   body,
	}
}

const GetDataErrPrefix = "Error by get record data"

func GetData(w http.ResponseWriter, r *http.Request, userID string, s store.Store) ResponseType {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: error by read Request Body - %v", GetDataErrPrefix, err),
			Code:   http.StatusBadRequest,
			Body:   []byte("Ошибка в запросе"),
		}
	}

	var data models.Data

	err = json.Unmarshal(body, &data)
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: error by read Request Body - %v", GetDataErrPrefix, err),
			Code:   http.StatusBadRequest,
			Body:   []byte("Ошибка в запросе"),
		}
	}

	result, ok, err := s.Get(r.Context(), data)

	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable get list data for user [%s] - %v", ListDataErrPrefix, userID, err),
			Code:   http.StatusInternalServerError,
		}
	}

	if !ok {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: this recod is not exist!", ListDataErrPrefix, userID, err),
			Code:   http.StatusInternalServerError,
		}
	}

	if result.Type == "file" {
		result.Body, _ = os.ReadFile(FilesDir + string(os.PathSeparator) + data.ID)
	}

	body, err = json.MarshalIndent(result, "", "    ")
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable marshal list data for user [%s] - %v", ListDataErrPrefix, userID, err),
			Code:   http.StatusInternalServerError,
		}
	}

	return ResponseType{
		LogMsg: fmt.Sprintf("get record data for user [%s]", userID),
		Code:   http.StatusOK,
		Body:   body,
	}
}

const UpdateDataErrPrefix = "Error by save data"
const FilesDir = "files"

func UpdateData(w http.ResponseWriter, r *http.Request, userID string, s store.Store) ResponseType {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: error by read Request Body - %v", UpdateDataErrPrefix, err),
			Code:   http.StatusBadRequest,
			Body:   []byte("Ошибка в запросе"),
		}
	}

	data := models.Data{
		UserID: userID,
		Date:   time.Now(),
	}

	if err = json.Unmarshal(body, &data); err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable decode json - %v", UpdateDataErrPrefix, err),
			Code:   http.StatusBadRequest,
			Body:   []byte("Ошибка в запросе"),
		}
	}

	if data.Type == "file" {
		return saveDataFile(w, r, data, s)
	}

	if data, err = s.Save(r.Context(), data); err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable save data - %v", UpdateDataErrPrefix, err),
			Code:   http.StatusInternalServerError,
		}
	}

	return ResponseType{
		LogMsg: fmt.Sprintf("data [%s] saved for user [%s]", data.ID, userID),
		Code:   http.StatusOK,
	}
}

func saveDataFile(w http.ResponseWriter, r *http.Request, data models.Data, s store.Store) ResponseType {
	var err error

	bodyData := data.Body

	data.Body = []byte{}

	if data, err = s.Save(r.Context(), data); err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable save data - %v", UpdateDataErrPrefix, err),
			Code:   http.StatusInternalServerError,
		}
	}

	f, err := os.Create(FilesDir + string(os.PathSeparator) + data.ID)
	defer f.Close()

	if err != nil {
		return ResponseType{
			LogMsg: fmt.Sprintf("%s: unable save file [%s], data - %v", UpdateDataErrPrefix, data.ID, err),
			Code:   http.StatusInternalServerError,
		}
	}

	f.Write(bodyData)

	return ResponseType{
		LogMsg: fmt.Sprintf("file saved for user [%s]", data.UserID),
		Code:   http.StatusOK,
	}
}
