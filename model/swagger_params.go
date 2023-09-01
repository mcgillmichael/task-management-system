package model

// GetTask input params.
// swagger:parameters getTaskEndpoint deleteTaskEndpoint
type GetTaskParams struct {
	// The ID of the task to retrieve.
	// in: path
	// required: true
	ID int `json:"id"`
}

// UpdateTaskParams defines the input parameters for updating a task.
// swagger:parameters updateTaskEndpoint
type UpdateTaskParams struct {
	// The ID of the task to update.
	// in: path
	// required: true
	ID int `json:"id"`
	// The details of the task to update.
	// in: body
	// required: true
	Body Task
}

// swagger:parameters assignTaskEndpoint
type AssignTaskParams struct {
	// The ID of the task to be assigned.
	// in: path
	// required: true
	TaskID int `json:"taskID"`

	// The ID of the user to which the task will be assigned.
	// in: path
	// required: true
	UserID int `json:"userID"`
}

// GetUserAssignedTasks input params.
// swagger:parameters getUserAssignedTasksEndpoint
type GetUserAssignedTasksParams struct {
	// The ID of the user whose assigned tasks are to be retrieved.
	// in: path
	// required: true
	UserID int `json:"userID"`
}

// GetTaskComments input params.
// swagger:parameters getTaskCommentsEndpoint
type GetTaskCommentsParams struct {
	// The ID of the task whose comments are to be retrieved.
	// in: path
	// required: true
	TaskID int `json:"taskID"`
}
