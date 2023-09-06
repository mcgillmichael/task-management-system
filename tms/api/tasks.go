package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"time"

	"github.com/julienschmidt/httprouter"

	"tms.zinkworks.com/model"
)

func (app *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fetching All Tasks")

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return
	}

	filePath := filepath.Join(wd, "tasks.json")

	jsonData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	var tasks []model.Task
	err = json.Unmarshal(jsonData, &tasks)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusOK, tasks, nil)
	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

func (app *application) showTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Println("Fetching Task with ID:", id)

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return
	}

	filePath := filepath.Join(wd, "tasks.json")

	jsonData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	var tasks []model.Task
	err = json.Unmarshal(jsonData, &tasks)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	desiredID := id

	// Find the task with the desired ID
	var foundTask *model.Task
	for _, task := range tasks {
		if int64(task.ID) == desiredID {
			foundTask = &task
			break
		}
	}

	// Check if the task was found
	if foundTask == nil {
		fmt.Printf("Task with ID %d not found\n", desiredID)
		return
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusOK, foundTask, nil)
	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

// swagger:route POST /tasks tasks createTaskEndpoint
// Create a new task.
// Inserts a new task and its items into the database.
// Consumes:
// - application/json
// Produces:
// - application/json
// Schemes: http, https
// responses:
//
//	201: taskCreatedResponse
//	400: badRequestError
//	500: internalServerError
func (app *application) createTaskHandler2(w http.ResponseWriter, r *http.Request) {

	var createTask model.Task

	err := json.NewDecoder(r.Body).Decode(&createTask)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createTask.CreatedAt = time.Now()
	createTask.UpdatedAt = time.Now()
	createTask.AssignedUserID = 0

	taskDto := model.TaskDto{DB: app.db}

	// Call the Insert method to insert the task into the database.
	err = taskDto.Insert(&createTask)
	if err != nil {
		http.Error(w, "Error inserting task", http.StatusInternalServerError)
		return
	}

	// Assuming you have the list of items for the task available in the createTask.Items slice.
	// You can now insert these items into the task_item table.
	for _, item := range createTask.Items {
		err = taskDto.InsertTaskItem(createTask.ID, item)
		if err != nil {
			http.Error(w, "Error inserting task items", http.StatusInternalServerError)
		}
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusCreated, createTask, nil)
	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}

// swagger:route GET /tasks tasks getAllTasksEndpoint
// Get all tasks.
// Fetches all tasks from the database.
// Produces:
// - application/json
// Schemes: http, https
// responses:
//
//	200: allTasksResponse
//	500: internalServerError
func (app *application) getAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	taskDto := model.TaskDto{DB: app.db}

	// Fetch all tasks from the database.
	tasks, err := taskDto.GetAllTasks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching tasks: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the tasks slice to JSON and send the response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// swagger:route DELETE /tasks/{id} tasks deleteTaskEndpoint
// Delete a task by ID.
// Removes a task from the database based on its ID.
// Produces:
// - application/json
// Schemes: http, https
//
// Responses:
//
//	200: successfullyDeletedResponse
//	400: invalidTaskIdError
//	500: internalServerError
func (app *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse the task ID from the URL parameters.
	taskID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskDto := model.TaskDto{DB: app.db}

	// Delete the task from the database.
	err = taskDto.DeleteTask(taskID)
	if err != nil {
		app.logger.Printf("Failed to delete task with ID %d: %v", taskID, err) // Log the error for more insight
		http.Error(w, "Error deleting task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:route PUT /tasks/{id} tasks updateTaskEndpoint
// Update an existing task.
// Updates a task by its ID with the details provided in the request body.
// Produces:
// - application/json
// Schemes: http, https
// Responses:
//
//	200: taskResponse
//	400: badRequestError
//	404: notFoundError
//	500: internalServerError
func (app *application) updateTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	taskDto := model.TaskDto{DB: app.db}

	// Parse the task ID from the URL parameters.
	taskID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	existingTask, err := taskDto.GetTask(taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Decode the request body into a new task object.
	var updateTask model.Task
	err = json.NewDecoder(r.Body).Decode(&updateTask)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingTask.Title = updateTask.Title
	existingTask.Description = updateTask.Description
	existingTask.Completed = updateTask.Completed
	existingTask.UpdatedAt = time.Now()

	existingTask.Items = updateTask.Items

	// Update the task in the database.
	err = taskDto.UpdateTask(taskID, existingTask)
	if err != nil {
		http.Error(w, "Error updating task", http.StatusInternalServerError)
		return
	}

	err = app.writeJSON(w, http.StatusOK, existingTask, nil)
	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

// swagger:route GET /tasks/{id} tasks getTaskEndpoint
// Get a task by ID.
// Fetches a task by its ID from the database.
// Produces:
// - application/json
// Schemes: http, https
//
// responses:
//
//	200: taskResponse
//	400: invalidTaskIdError
//	404: notFoundError
//	500: internalServerError
func (app *application) getTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse the task ID from the URL parameters.
	taskID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskDto := model.TaskDto{DB: app.db}

	// Fetch the task from the database by its ID.
	task, err := taskDto.GetTask(taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Error fetching task", http.StatusInternalServerError)
		}
		return
	}

	// Encode the task to JSON and send the response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// swagger:route PATCH /tasks/{taskID}/assign/{userID} tasks assignTaskEndpoint
// Assign a user to a task.
// Assigns a user to a specific task based on task and user IDs.
// Produces:
// - application/json
// Schemes: http, https
//
// Responses:
//
//	200: taskResponse
//	400: invalidIdError
//	404: notFoundError
//	500: internalServerError
func (app *application) assignTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	taskDto := model.TaskDto{DB: app.db}

	// Parse the task ID from the URL parameters.
	taskID, err := strconv.Atoi(ps.ByName("taskID"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(ps.ByName("userID"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	existingTask, err := taskDto.GetTask(taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	existingTask.AssignedUserID = userID
	existingTask.UpdatedAt = time.Now()

	// Use the AssignUserToTask method with updatedAt to perform the assignment of the user to the task.
	err = taskDto.AssignUserToTask(taskID, userID, existingTask.UpdatedAt)
	if err != nil {
		http.Error(w, "Error assigning user to task", http.StatusInternalServerError)
		return
	}

	err = app.writeJSON(w, http.StatusOK, existingTask, nil)
	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

// swagger:route GET /users/{userID}/tasks/assigned tasks getUserAssignedTasksEndpoint
// Get tasks assigned to a specific user.
// Returns a list of tasks assigned to a user based on their ID.
// Produces:
// - application/json
// Schemes: http, https
// Responses:
//
//	200: allTasksResponse
//	400: invalidIdError
//	500: internalServerError
func (app *application) getTasksAssignedToUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse the user ID from the URL parameters.
	userID, err := strconv.Atoi(ps.ByName("userID"))
	if err != nil || userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	taskDto := model.TaskDto{DB: app.db}

	// Fetch all tasks from the database.
	tasks, err := taskDto.GetAllTaskByAssignedUserID(userID)
	if err != nil {
		http.Error(w, "Error fetching tasks", http.StatusInternalServerError)
		return
	}

	// Encode the tasks slice to JSON and send the response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)

}

// swagger:route POST /comments tasks createTaskCommentsEndpoint
// Create a task comment.
// Inserts a new task comment into the database.
// Consumes:
// - application/json
// Produces:
// - application/json
// Schemes: http, https
// responses:
//
//	201: taskCommentCreatedResponse
//	400: badRequestError
//	500: internalServerError
func (app *application) createTaskCommentsHandler(w http.ResponseWriter, r *http.Request) {

	var createTaskComment model.TaskComment

	err := json.NewDecoder(r.Body).Decode(&createTaskComment)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createTaskComment.CreatedAt = time.Now()
	taskDto := model.TaskDto{DB: app.db}

	// Call the Insert method to insert the task comment into the database.
	err = taskDto.InsertTaskComment(&createTaskComment)
	if err != nil {
		http.Error(w, "Error inserting task", http.StatusInternalServerError)
		return
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusCreated, createTaskComment, nil)
	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}

// swagger:route GET /comments/{taskID} tasks getTaskCommentsEndpoint
// Get comments for a specific task.
// Returns a list of comments for a task based on its ID.
// Produces:
// - application/json
// Schemes: http, https
// Responses:
//
//	200: allCommentsResponse
//	400: invalidIdError
//	500: internalServerError
func (app *application) getAllTaskCommentsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse the task ID from the URL parameters.
	taskID, err := strconv.Atoi(ps.ByName("taskID"))
	if err != nil || taskID == 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskDto := model.TaskDto{DB: app.db}

	// Fetch all tasks from the database.
	tasks, err := taskDto.GetAllTaskCommentsByTaskID(taskID)
	if err != nil {
		http.Error(w, "Error fetching tasks", http.StatusInternalServerError)
		return
	}

	// Encode the tasks slice to JSON and send the response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)

}
