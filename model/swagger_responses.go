package model

// Health check response indicating the status and details of the service.
// swagger:response healthCheckResponse
type HealthCheckResponse struct {
	Status      string `json:"status" example:"available"`
	Environment string `json:"environment" example:"development"`
	Version     string `json:"version" example:"1.0.0"`
}

// Response for successfully created task.
// swagger:response taskCreatedResponse
type TaskCreatedResponse struct {
	// in: body
	Body Task `json:"body"`
}

// Response for getting all tasks
// swagger:response allTasksResponse
type AllTasksResponse struct {
	// in: body
	Body []Task `json:"body"`
}

// Response for a successfully created task comment.
// swagger:response taskCommentCreatedResponse
type TaskCommentCreatedResponse struct {
	// in: body
	Body TaskComment `json:"body"`
}

// Response for successfully retrieved task by ID.
// swagger:response taskResponse
type TaskResponse struct {
	// in: body
	Body Task `json:"body"`
}

// Bad request due to client-side error, e.g., invalid task ID.
// swagger:response invalidTaskIdError
type InvalidTaskIdError struct {
	Error string `json:"error"`
}

// Indicates the task was not found.
// swagger:response notFoundError
type NotFoundError struct {
	Error string `json:"error"`
}

// Bad request due to client-side error, e.g., invalid request body.
// swagger:response badRequestError
type BadRequestError struct {
	Error string `json:"error"`
}

// Server encountered a problem.
// swagger:response internalServerError
type InternalServerError struct {
	Error string `json:"error"`
}

// Response indicating successful deletion.
// swagger:response successfullyDeletedResponse
type SuccessfullyDeletedResponse struct {
	Message string `json:"message"`
}

// Invalid ID error due to client-side error, specifically when trying to assign a user to a task.
// swagger:response invalidIdError
type InvalidIdErrorResponse struct {
	Error string `json:"error" example:"Error assigning user to task"`
}

// Response for getting all comments for a task
// swagger:response allCommentsResponse
type AllCommentsResponse struct {
	// in: body
	Body []TaskComment `json:"body"`
}
