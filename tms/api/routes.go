package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/tasks", app.createTaskHandler2)
	router.HandlerFunc(http.MethodGet, "/tasks", app.getAllTasksHandler)
	router.HandlerFunc(http.MethodPost, "/comments", app.createTaskCommentsHandler)
	router.Handle(http.MethodGet, "/tasks/:id", httprouter.Handle(app.getTaskHandler))

	// @Summary Update a specific task by ID
	// @Description Updates a task based on the input provided
	// @Accept  json
	// @Produce  json
	// @Param   id     path    int     true        "Task ID"
	// @Param   task     body    model.Task     true        "Task body"
	// @Success 200 {object} model.Task
	// @Router /tasks/{id} [put]
	router.Handle(http.MethodPut, "/tasks/:id", httprouter.Handle(app.updateTaskHandler))

	// @Summary Delete a specific task by ID
	// @Description Deletes a task based on its ID
	// @Produce  json
	// @Param   id     path    int     true        "Task ID"
	// @Success 204 {string} string	"No Content"
	// @Router /tasks/{id} [delete]
	router.Handle(http.MethodDelete, "/tasks/:id", httprouter.Handle(app.deleteTaskHandler))

	// @Summary Assign a user to a task
	// @Description Assigns a user to a specific task by their IDs
	// @Accept  json
	// @Produce  json
	// @Param   taskID     path    int     true        "Task ID"
	// @Param   userID     path    int     true        "User ID"
	// @Success 200 {object} model.Assignment
	// @Router /tasks/{taskID}/assign/{userID} [patch]
	router.Handle(http.MethodPatch, "/tasks/:taskID/assign/:userID", httprouter.Handle(app.assignTaskHandler))

	// @Summary Get tasks assigned to a user
	// @Description Returns a list of tasks assigned to a specific user
	// @Produce json
	// @Param   userID     path    int     true        "User ID"
	// @Success 200 {array} model.Task
	// @Router /users/{userID}/tasks/assigned [get]
	router.Handle(http.MethodGet, "/users/:userID/tasks/assigned", httprouter.Handle(app.getTasksAssignedToUserHandler))

	// @Summary Get comments for a specific task
	// @Description Returns a list of comments for a task
	// @Produce json
	// @Param taskID path int true "Task Comment ID"
	// @Success 200 {array} model.TaskComment
	// @Router /comments/{taskID} [get]
	router.Handle(http.MethodGet, "/comments/:taskID", httprouter.Handle(app.getAllTaskCommentsHandler))

	return router
}
