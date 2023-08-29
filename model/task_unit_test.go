package model

import (
	"database/sql"
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
