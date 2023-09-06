package model

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Unit testing
func newMockDB() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	return db, mock
}

func TestGetAllTasks_SuccessfulRetrieval(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	taskDto := TaskDto{DB: db}

	// Mock the expected rows
	rows := sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_at", "updated_at", "assigned_user_id", "item"}).
		AddRow(1, "TestTitle1", "TestDescription1", false, time.Now(), time.Now(), 0, "Item1").
		AddRow(2, "TestTitle2", "TestDescription2", true, time.Now(), time.Now(), 1, "Item2")
	mock.ExpectQuery(`SELECT (.+) FROM task`).WillReturnRows(rows)

	tasks, err := taskDto.GetAllTasks()
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestInsert(t *testing.T) {
	db, mock := newMockDB()
	defer db.Close()

	taskDto := TaskDto{DB: db}

	mock.ExpectPrepare(`INSERT INTO task (.+) RETURNING id`).ExpectQuery().
		WithArgs("TestTitle", "TestDescription", false, time.Now(), time.Now()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	task := &Task{
		Title:       "TestTitle",
		Description: "TestDescription",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := taskDto.Insert(task)
	assert.NoError(t, err)
	assert.Equal(t, 1, task.ID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestUpdateTask_SuccessfulUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	task := &Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   false,
		Items:       []string{"item1", "item2"},
	}

	// Mock for the initial UPDATE.
	mock.ExpectPrepare("^UPDATE task SET title.*WHERE id = \\$5").
		ExpectExec().
		WithArgs(task.Title, task.Description, task.Completed, sqlmock.AnyArg(), task.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock for DELETE task items.
	mock.ExpectExec("^DELETE FROM task_item WHERE task_id = \\$1$").
		WithArgs(task.ID).
		WillReturnResult(sqlmock.NewResult(1, 2))

	// Mock for inserting new task items.
	mock.ExpectPrepare("^INSERT INTO task_item \\(task_id, item\\) VALUES \\(\\$1, \\$2\\)$")
	for _, item := range task.Items {
		mock.ExpectExec("^INSERT INTO task_item \\(task_id, item\\) VALUES \\(\\$1, \\$2\\)$").
			WithArgs(task.ID, item).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Call method
	err = taskDto.UpdateTask(task.ID, task)
	assert.NoError(t, err)

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAssignUserToTask_SuccessfulAssign(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	id, userID := 1, 42
	updatedAt := time.Now()

	// Mock the preparation and execution of the UPDATE query.
	mock.ExpectPrepare("^UPDATE task SET assigned_to.*WHERE id = \\$3").
		ExpectExec().
		WithArgs(userID, updatedAt, id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method.
	err = taskDto.AssignUserToTask(id, userID, updatedAt)
	assert.NoError(t, err)

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDeleteTask_SuccessfulDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	id := 1

	// Mock the preparation and execution of the DELETE query.
	mock.ExpectPrepare("^DELETE FROM task WHERE id = \\$1$").
		ExpectExec().
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method.
	err = taskDto.DeleteTask(id)
	assert.NoError(t, err)

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestInsertTaskItem_SuccessfulInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	taskID := 1
	item := "Test item"

	// Mock the preparation and execution of the INSERT query.
	mock.ExpectPrepare("^INSERT INTO task_item \\(task_id, item\\) VALUES \\(\\$1, \\$2\\)$").
		ExpectExec().
		WithArgs(taskID, item).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method.
	err = taskDto.InsertTaskItem(taskID, item)
	assert.NoError(t, err)

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestInsertTaskComment_SuccessfulInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	taskComment := &TaskComment{
		TaskID:    1,
		Comment:   "Sample comment",
		CreatedAt: time.Now(),
	}

	// Mock the preparation and execution of the INSERT query.
	mock.ExpectPrepare("^INSERT INTO task_comment \\(task_id, comment, created_at\\) VALUES \\(\\$1, \\$2, \\$3\\)$").
		ExpectExec().
		WithArgs(taskComment.TaskID, taskComment.Comment, taskComment.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method.
	err = taskDto.InsertTaskComment(taskComment)
	assert.NoError(t, err)

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetTask_SuccessfulGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	id := 1
	// Mocking the rows you'll be retrieving.
	columns := []string{"id", "title", "description", "completed", "created_at", "updated_at", "assigned_user_id", "item"}
	mockRows := sqlmock.NewRows(columns).
		AddRow(1, "Test Title", "Test Description", false, time.Now(), time.Now(), 42, "Item 1").
		AddRow(1, "Test Title", "Test Description", false, time.Now(), time.Now(), 42, "Item 2")

	mock.ExpectQuery("^SELECT t.id, t.title, t.description.*FROM task t.*WHERE t.id = \\$1$").
		WithArgs(id).
		WillReturnRows(mockRows)

	// Call the method.
	task, err := taskDto.GetTask(id)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, 2, len(task.Items)) // since we added two rows with two items.

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetAllTaskByAssignedUserID_SuccessfulGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	userID := 42
	// Mocking the rows you'll be retrieving.
	columns := []string{"id", "title", "description", "completed", "created_at", "updated_at", "assigned_user_id", "item"}
	mockRows := sqlmock.NewRows(columns).
		AddRow(1, "Test Title 1", "Description 1", false, time.Now(), time.Now(), userID, "Item 1").
		AddRow(1, "Test Title 1", "Description 1", false, time.Now(), time.Now(), userID, "Item 2").
		AddRow(2, "Test Title 2", "Description 2", false, time.Now(), time.Now(), userID, "Item A").
		AddRow(2, "Test Title 2", "Description 2", false, time.Now(), time.Now(), userID, "Item B")

	mock.ExpectQuery("^SELECT t.id, t.title, t.description.*FROM task t.*WHERE t.assigned_user_id = \\$1$").
		WithArgs(userID).
		WillReturnRows(mockRows)

	// Call the method.
	tasks, err := taskDto.GetAllTaskByAssignedUserID(userID)
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Equal(t, 2, len(tasks))
	assert.Equal(t, 2, len(tasks[0].Items)) // 2 items for task 1.
	assert.Equal(t, 2, len(tasks[1].Items)) // 2 items for task 2.

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetAllTaskCommentsByTaskID_SuccessfulGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	taskDto := TaskDto{DB: db}

	taskID := 1
	// Mocking the rows you'll be retrieving.
	columns := []string{"id", "task_id", "comment", "created_at"}
	mockRows := sqlmock.NewRows(columns).
		AddRow(1, taskID, "Comment 1", time.Now()).
		AddRow(2, taskID, "Comment 2", time.Now()).
		AddRow(3, taskID, "Comment 3", time.Now())

	mock.ExpectQuery("^SELECT tc.id, tc.task_id, tc.comment.*FROM task_comment tc.*WHERE tc.task_id = \\$1$").
		WithArgs(taskID).
		WillReturnRows(mockRows)

	// Call the method.
	taskComments, err := taskDto.GetAllTaskCommentsByTaskID(taskID)
	assert.NoError(t, err)
	assert.NotNil(t, taskComments)
	assert.Equal(t, 3, len(taskComments)) // 3 comments for the task.

	// Ensure all mock expectations were met.
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// Integration Tests
func TestHealthcheckHandler(t *testing.T) {
	resp, err := http.Get("http://localhost:4000/healthcheck")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Decode the response body
	var data map[string]string
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}

	// Check for the 'status' key-value
	if data["status"] != "available" {
		t.Errorf("want status to be 'available'; got %s", data["status"])
	}
}

func TestCreateTaskHandler2(t *testing.T) {
	// Define the task to be created
	taskData := `{
        "title": "Test Task",
        "description": "This is a test task",
        "items": [
			"This is a test item",
			"This is a test item 2"
        ]
    }`

	// Create a new HTTP request with the task data
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBufferString(taskData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if the status code is 201 Created
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code 201; got %d", resp.StatusCode)
	}

	// Decode the response body
	var taskResponse Task
	err = json.NewDecoder(resp.Body).Decode(&taskResponse)
	if err != nil {
		t.Fatal(err)
	}

	// Further assertions can be made on the taskResponse if needed
	if taskResponse.Title != "Test Task" {
		t.Errorf("Expected task title 'Test Task'; got %s", taskResponse.Title)
	}
}

func TestGetAllTasksHandler(t *testing.T) {
	resp, err := http.Get("http://localhost:4000/tasks")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	// Check for status code 200 OK.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", resp.StatusCode)
	}

	// Decode the response.
	var tasks []Task
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestCreateTaskCommentsHandler(t *testing.T) {
	// Define the task comment to be created
	taskCommentData := `{
        "task_id": 1, 
        "comment": "This is a test comment for task"
    }`

	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/comments", bytes.NewBufferString(taskCommentData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	// Check for status code 201 Created
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %v", resp.StatusCode)
	}

	// Decode the response to verify the comment was properly created
	var comment TaskComment
	err = json.NewDecoder(resp.Body).Decode(&comment)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check if the returned comment matches the one we've created
	if comment.TaskID != 1 {
		t.Errorf("Expected TaskID 1, got %d", comment.TaskID)
	}
	if comment.Comment != "This is a test comment for task" {
		t.Errorf("Expected comment 'This is a test comment for task', got '%s'", comment.Comment)
	}
}

func TestGetTaskHandler(t *testing.T) {
	// First, let's test with a valid task ID that exists in the database.
	taskID := 1 // You might want to change this to an actual valid ID in your database.
	resp, err := http.Get(fmt.Sprintf("http://localhost:4000/tasks/%d", taskID))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	// Check for status code 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", resp.StatusCode)
	}

	// Decode the response to verify the task
	var task Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check if the returned task matches the ID we've requested
	if task.ID != taskID {
		t.Errorf("Expected task ID %d, got %d", taskID, task.ID)
	}

	// Now, test with an invalid task ID.
	resp, err = http.Get("http://localhost:4000/tasks/invalidID")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %v", resp.StatusCode)
	}

	// Finally, test with a non-existing task ID.
	taskID = 99999 // Assuming this ID doesn't exist.
	resp, err = http.Get(fmt.Sprintf("http://localhost:4000/tasks/%d", taskID))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %v", resp.StatusCode)
	}
}

func TestUpdateTaskHandler(t *testing.T) {
	// URL for the task with ID 1
	url := "http://localhost:4000/tasks/1"

	// Define the updated task data.
	updateData := `{
        "title": "Updated Task",
        "description": "This is an updated task description",
        "completed": true,
        "items": [
			"Updated Item 1",
			"Updated Item 2"
        ]
    }`

	// Making the PUT request to update the task with ID 1
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(updateData))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make PUT request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d; got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode the response body to check if the task was updated correctly.
	var updatedTask Task
	err = json.NewDecoder(resp.Body).Decode(&updatedTask)
	if err != nil {
		t.Fatalf("Error decoding response body: %v", err)
	}

	// Validate the updated task data.
	if updatedTask.Title != "Updated Task" {
		t.Errorf("Expected title to be 'Updated Task'; got '%s'", updatedTask.Title)
	}
}

func TestDeleteTaskHandler(t *testing.T) {
	// Step 1: Create a temporary task

	// Define the task to be created
	taskData := `{
        "title": "Temp Task",
        "description": "This is a temporary task for testing deletion",
        "items": [
			"Temp Item 1",
			"Temp Item 2"
        ]
    }`

	// Create a new HTTP request with the task data
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBufferString(taskData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP request to create the task
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if the status code is 201 Created
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code 201 for task creation; got %d", resp.StatusCode)
	}

	// Decode the response body to get the task details
	var taskResponse Task
	err = json.NewDecoder(resp.Body).Decode(&taskResponse)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Delete the temporary task

	// Create a new HTTP request for deletion
	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:4000/tasks/%d", taskResponse.ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Perform the HTTP request to delete the task
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if the status code is 200 OK after deletion
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200 for task deletion; got %d", resp.StatusCode)
	}

	// Step 3: Verify that the task has been deleted

	// Try to fetch the deleted task
	resp, err = http.Get(fmt.Sprintf("http://localhost:4000/tasks/%d", taskResponse.ID))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Verify that the status code is 404 Not Found for the deleted task
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status code 404 for the deleted task; got %d", resp.StatusCode)
	}
}

func TestAssignTaskHandler(t *testing.T) {
	// Step 1: Create a temporary task
	taskData := `{
        "title": "Temp Task for Assignment",
        "description": "This is a temporary task for testing assignment",
        "items": ["Temp Item 1", "Temp Item 2"]
    }`

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBufferString(taskData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error during task creation: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 for task creation; got %d", resp.StatusCode)
	}

	var taskResponse Task
	if err := json.NewDecoder(resp.Body).Decode(&taskResponse); err != nil {
		t.Fatalf("Error decoding task creation response: %v", err)
	}

	// Step 2: Assign the user to the task
	const userID = 1
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:4000/tasks/%d/assign/%d", taskResponse.ID, userID), nil)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error during task assignment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Expected 200 for task assignment; got %d. Response: %s", resp.StatusCode, responseBody)
	}

	var assignedTaskResponse Task
	if err := json.NewDecoder(resp.Body).Decode(&assignedTaskResponse); err != nil {
		t.Fatalf("Error decoding task assignment response: %v", err)
	}

	if assignedTaskResponse.AssignedUserID != userID {
		t.Fatalf("Expected assigned user ID %d; got %d", userID, assignedTaskResponse.AssignedUserID)
	}

	// Cleanup: Delete the test task after the test
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:4000/tasks/%d", taskResponse.ID), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error during task deletion post-test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 for task deletion post-test; got %d", resp.StatusCode)
	}
}

func TestGetTasksAssignedToUserHandler(t *testing.T) {
	// Use a random user ID for testing since there's no user creation API.
	const userID = 1

	// Step 1: Create a temporary task
	taskData := `{
		"title": "Temp Task for Assignment",
		"description": "This is a temporary task for testing assignment",
		"items": ["Temp Item 1", "Temp Item 2"]
	}`

	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBufferString(taskData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error during task creation request: %v", err)
	}
	defer resp.Body.Close()

	var taskResponse Task
	err = json.NewDecoder(resp.Body).Decode(&taskResponse)
	if err != nil {
		t.Fatalf("Error decoding task creation response: %v", err)
	}

	// Step 2: Assign this task to the user
	req, err = http.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:4000/tasks/%d/assign/%d", taskResponse.ID, userID), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Step 3: Use the getTasksAssignedToUserHandler to fetch tasks for the user
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:4000/users/%d/tasks/assigned", userID), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var tasksResponse []Task
	err = json.NewDecoder(resp.Body).Decode(&tasksResponse)
	if err != nil {
		t.Fatalf("Error decoding tasks response: %v", err)
	}

	// Step 4: Validate the tasks retrieved
	found := false
	for _, task := range tasksResponse {
		if task.ID == taskResponse.ID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Expected to retrieve task with ID %d, but it was not found in the response", taskResponse.ID)
	}

	// Step 5: Cleanup - Delete the temporary task
	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:4000/tasks/%d", taskResponse.ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}

func TestGetAllTaskCommentsHandler(t *testing.T) {
	// Step 1: Create a temporary task
	taskData := `{
		"title": "Temp Task for Comment Testing",
		"description": "This is a temporary task for testing comments retrieval",
		"items": ["Temp Item 1", "Temp Item 2"]
	}`

	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBufferString(taskData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error during task creation request: %v", err)
	}
	defer resp.Body.Close()

	var taskResponse Task
	err = json.NewDecoder(resp.Body).Decode(&taskResponse)
	if err != nil {
		t.Fatalf("Error decoding task creation response: %v", err)
	}

	// Step 2: Add comments to the task
	comments := []string{"Comment 1", "Comment 2"}
	for _, commentText := range comments {
		commentData := fmt.Sprintf(`{"task_id": %d, "comment": "%s"}`, taskResponse.ID, commentText)
		req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/comments", bytes.NewBufferString(commentData))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Error adding comment: %v", err)
		}
		defer resp.Body.Close()
	}

	// Step 3: Use the getAllTaskCommentsHandler to fetch comments for the task
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:4000/comments/%d", taskResponse.ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var commentsResponse []TaskComment
	err = json.NewDecoder(resp.Body).Decode(&commentsResponse)
	if err != nil {
		t.Fatalf("Error decoding comments response: %v", err)
	}

	// Step 4: Validate the comments retrieved
	if len(commentsResponse) != len(comments) {
		t.Fatalf("Expected %d comments, got %d", len(comments), len(commentsResponse))
	}

	for _, commentText := range comments {
		found := false
		for _, comment := range commentsResponse {
			if comment.Comment == commentText {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Expected to retrieve comment '%s', but it was not found in the response", commentText)
		}
	}

	// Step 5: Cleanup - Delete the temporary task
	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:4000/tasks/%d", taskResponse.ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}
