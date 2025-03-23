package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"todo_restapi/internal/config"
	"todo_restapi/internal/constants"
	"todo_restapi/internal/http-server/middlewares"
	"todo_restapi/internal/models"
	"todo_restapi/internal/services"
	"todo_restapi/internal/storage"
)

type TaskHandler struct {
	Storage     *storage.Storage
	Config      *config.Config
	AuthService *middlewares.AuthService
}

func NewTaskHandler(storage *storage.Storage, cfg *config.Config) *TaskHandler {
	return &TaskHandler{
		Storage:     storage,
		Config:      cfg,
		AuthService: middlewares.NewAuthService(cfg),
	}
}

func (h *TaskHandler) NextDate(write http.ResponseWriter, request *http.Request) {

	timeNow := request.FormValue("now")
	date := request.FormValue("date")
	repeat := request.FormValue("repeat")

	timeParse, err := time.Parse(constants.DateFormat, timeNow)
	if err != nil {
		http.Error(write, fmt.Sprintf("time parse error: %v", err), http.StatusInternalServerError)
		return
	}

	result, err := services.NextDate(timeParse, date, repeat)
	if err != nil {
		http.Error(write, fmt.Sprintf("NextDate: function error: %v", err), http.StatusBadRequest)
		return
	}

	write.WriteHeader(http.StatusOK)

	if _, err := write.Write([]byte(result)); err != nil {
		http.Error(write, "failed to write response", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) GetTask(write http.ResponseWriter, request *http.Request) {

	id := request.FormValue("id")

	task, err := h.Storage.GetTask(id)
	if err != nil {
		services.WriteJSONError(write, http.StatusBadRequest, fmt.Sprintf("GetTask: function error: %v", err))
		return
	}

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(write).Encode(task); err != nil {
		http.Error(write, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) AddTask(write http.ResponseWriter, request *http.Request) {

	now := time.Now().Format(constants.DateFormat)
	newTask := new(models.Task)

	if err := json.NewDecoder(request.Body).Decode(newTask); err != nil {
		http.Error(write, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := services.ValidateTaskRequest(newTask, now); err != nil {
		services.WriteJSONError(write, http.StatusBadRequest, fmt.Sprintf("ValidateTaskRequest: function error: %v", err))
		return
	}

	taskID, err := h.Storage.AddTask(*newTask)
	if err != nil {
		http.Error(write, fmt.Sprintf("AddTask: add task error: %v", err), http.StatusInternalServerError)
		return
	}

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusCreated)

	response := map[string]int64{"id": taskID}

	if err := json.NewEncoder(write).Encode(response); err != nil {
		http.Error(write, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) EditTask(write http.ResponseWriter, request *http.Request) {

	now := time.Now().Format(constants.DateFormat)
	newTask := new(models.Task)

	if err := json.NewDecoder(request.Body).Decode(newTask); err != nil {
		http.Error(write, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := services.ValidateTaskRequest(newTask, now); err != nil {
		services.WriteJSONError(write, http.StatusBadRequest, fmt.Sprintf("ValidateTaskRequest: function error: %v", err))
		return
	}

	if err := h.Storage.EditTask(*newTask); err != nil {
		services.WriteJSONError(write, http.StatusInternalServerError, fmt.Sprintf("EditTask: function error: %v", err))
		return
	}

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(write).Encode(map[string]interface{}{}); err != nil {
		http.Error(write, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) DeleteTask(write http.ResponseWriter, request *http.Request) {

	id := request.FormValue("id")

	if err := h.Storage.DeleteTask(id); err != nil {
		services.WriteJSONError(write, http.StatusInternalServerError, fmt.Sprintf("DeleteTask: function error: %v", err))
		return
	}

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(write).Encode(struct{}{}); err != nil {
		http.Error(write, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) GetTasks(write http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodGet {
		http.Error(write, "invalid method", http.StatusMethodNotAllowed)
	}

	searchQuery := request.FormValue("search")

	if searchQuery != "" {
		searchTasks, err := h.Storage.SearchTasks(searchQuery)
		if err != nil {
			services.WriteJSONError(write, http.StatusInternalServerError, fmt.Sprintf("SearchTasks: function error: %v", err))
			return
		}

		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(http.StatusOK)

		response := map[string][]models.Task{"tasks": searchTasks}

		if err := json.NewEncoder(write).Encode(response); err != nil {
			http.Error(write, "failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	}

	tasks, err := h.Storage.GetTasks()
	if err != nil {
		http.Error(write, fmt.Sprintf("GetTasks: function error: %v", err), http.StatusInternalServerError)
		return
	}

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)

	response := map[string][]models.Task{"tasks": tasks}

	if err := json.NewEncoder(write).Encode(response); err != nil {
		http.Error(write, "failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (h *TaskHandler) TaskIsDone(write http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		http.Error(write, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	id := request.FormValue("id")

	task, err := h.Storage.GetTask(id)
	if err != nil {
		services.WriteJSONError(write, http.StatusBadRequest, fmt.Sprintf("GetTask: function error: %v", err))
		return
	}

	if task.Repeat == "" {
		if err := h.Storage.DeleteTask(id); err != nil {
			services.WriteJSONError(write, http.StatusInternalServerError, fmt.Sprintf("DeleteTask: function error: %v", err))
			return
		}
	} else {
		nextDate, err := services.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			services.WriteJSONError(write, http.StatusInternalServerError, fmt.Sprintf("NextDate error: %v", err))
			return
		}

		task.Date = nextDate

		if err := h.Storage.EditTask(task); err != nil {
			services.WriteJSONError(write, http.StatusInternalServerError, fmt.Sprintf("EditTask error: %v", err))
			return
		}
	}

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(write).Encode(struct{}{}); err != nil {
		http.Error(write, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) Authentication(write http.ResponseWriter, request *http.Request) {

	type authResponse struct {
		Token string `json:"token"`
	}

	type password struct {
		Password string `json:"password"`
	}

	pwdFromJSON := password{
		Password: "",
	}

	if request.Method != http.MethodPost {
		http.Error(write, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(request.Body).Decode(&pwdFromJSON); err != nil {
		http.Error(write, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	pwd := pwdFromJSON.Password

	if pwd == "" {
		services.WriteJSONError(write, http.StatusBadRequest, "password cannot be empty")
		return
	}

	token, err := h.AuthService.GenerateJWT(pwd)
	if err != nil {
		services.WriteJSONError(write, http.StatusBadRequest, fmt.Sprintf("GenerateJWT: function error: %v", err))
		return
	}

	response := authResponse{Token: token}

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(write).Encode(response); err != nil {
		http.Error(write, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
