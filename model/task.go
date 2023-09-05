package model

import (
	"database/sql"
	"time"
)

type Task struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Completed      bool      `json:"completed"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AssignedUserID int       `json:"assigned_user_id,omitempty"`
	Items          []string  `json:"items"`
	Comments       []string  `json:"comments,omitempty"`
}

type TaskDto struct {
	DB *sql.DB
}

type TaskComment struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

func (taskDto TaskDto) GetAllTasks() ([]Task, error) {

	query := `
		SELECT t.id, t.title, t.description, t.completed, t.created_at, t.updated_at, t.assigned_user_id, ti.item
		FROM task t
		LEFT JOIN task_item ti ON t.id = ti.task_id
		ORDER BY t.id, ti.id
	`

	rows, err := taskDto.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskMap := make(map[int]*Task)

	taskItemsMap := make(map[int][]string)

	for rows.Next() {
		var taskItem sql.NullString
		task := &Task{}
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.AssignedUserID,
			&taskItem,
		)
		if err != nil {
			return nil, err
		}

		if taskItem.Valid {
			taskItemsMap[task.ID] = append(taskItemsMap[task.ID], taskItem.String)
		}

		if _, ok := taskMap[task.ID]; !ok {
			taskMap[task.ID] = task
		}
	}

	for taskID, taskItems := range taskItemsMap {
		if task, ok := taskMap[taskID]; ok {
			task.Items = taskItems
		}
	}

	tasks := make([]Task, 0, len(taskMap))
	for _, task := range taskMap {
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (taskDto TaskDto) Insert(task *Task) error {

	stmt, err := taskDto.DB.Prepare(`
			INSERT INTO task (title, description, completed, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var taskID int
	err = stmt.QueryRow(task.Title, task.Description, task.Completed, task.CreatedAt, task.UpdatedAt).Scan(&taskID)
	if err != nil {
		return err
	}

	task.ID = taskID

	return nil

}

func (taskDto TaskDto) Get(id int64) (*Task, error) {
	return nil, nil
}

func (taskDto TaskDto) UpdateTask(id int, task *Task) error {

	stmt, err := taskDto.DB.Prepare(`
		UPDATE task
		SET title = $1, description = $2, completed = $3, updated_at = $4
		WHERE id = $5
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(task.Title, task.Description, task.Completed, time.Now(), id)
	if err != nil {
		return err
	}

	_, err = taskDto.DB.Exec("DELETE FROM task_item WHERE task_id = $1", task.ID)
	if err != nil {
		return err
	}

	stmt, err = taskDto.DB.Prepare("INSERT INTO task_item (task_id, item) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range task.Items {
		_, err := stmt.Exec(task.ID, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (taskDto TaskDto) AssignUserToTask(id, userID int, updatedAt time.Time) error {
	stmt, err := taskDto.DB.Prepare(`
		UPDATE task
		SET assigned_to = $1, updated_at = $2
		WHERE id = $3
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, updatedAt, id)
	if err != nil {
		return err
	}

	return nil
}

func (taskDto TaskDto) DeleteTask(id int) error {

	stmt, err := taskDto.DB.Prepare(`
		DELETE FROM task WHERE id = $1
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

func (taskDto TaskDto) InsertTaskItem(taskID int, item string) error {

	stmt, err := taskDto.DB.Prepare(`
		INSERT INTO task_item (task_id, item)
		VALUES ($1, $2)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskID, item)
	if err != nil {
		return err
	}

	return nil
}

func (taskDto TaskDto) InsertTaskComment(taskComment *TaskComment) error {

	stmt, err := taskDto.DB.Prepare(`
		INSERT INTO task_comment (task_id, comment, created_at)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(taskComment.TaskID, taskComment.Comment, taskComment.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (taskDto TaskDto) GetTask(id int) (*Task, error) {
	query := `
		SELECT t.id, t.title, t.description, t.completed, t.created_at, t.updated_at, t.assigned_user_id, ti.item
		FROM task t
		LEFT JOIN task_item ti ON t.id = ti.task_id
		WHERE t.id = $1
	`

	rows, err := taskDto.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	task := &Task{}
	task.Items = make([]string, 0)

	hasRows := false
	for rows.Next() {
		hasRows = true
		var taskItem sql.NullString
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.AssignedUserID,
			&taskItem,
		)
		if err != nil {
			return nil, err
		}

		if taskItem.Valid {
			task.Items = append(task.Items, taskItem.String)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if !hasRows {
		return nil, sql.ErrNoRows
	}

	return task, nil
}

func (taskDto TaskDto) GetAllTaskByAssignedUserID(userID int) ([]Task, error) {
	query := `
		SELECT t.id, t.title, t.description, t.completed, t.created_at, t.updated_at, t.assigned_user_id, ti.item
		FROM task t
		LEFT JOIN task_item ti ON t.id = ti.task_id
		WHERE t.assigned_user_id = $1
	`

	rows, err := taskDto.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskMap := make(map[int]*Task)

	taskItemsMap := make(map[int][]string)

	for rows.Next() {
		var taskItem sql.NullString
		task := &Task{}
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.AssignedUserID,
			&taskItem,
		)
		if err != nil {
			return nil, err
		}

		if taskItem.Valid {
			taskItemsMap[task.ID] = append(taskItemsMap[task.ID], taskItem.String)
		}

		if _, ok := taskMap[task.ID]; !ok {
			taskMap[task.ID] = task
		}
	}

	for taskID, taskItems := range taskItemsMap {
		if task, ok := taskMap[taskID]; ok {
			task.Items = taskItems
		}
	}

	tasks := make([]Task, 0, len(taskMap))
	for _, task := range taskMap {
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (taskDto TaskDto) GetAllTaskCommentsByTaskID(taskID int) ([]TaskComment, error) {
	query := `
		SELECT tc.id, tc.task_id, tc.comment, tc.created_at
		FROM task_comment tc
		WHERE tc.task_id = $1
	`

	rows, err := taskDto.DB.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskCommentsMap := make(map[int]*TaskComment)

	for rows.Next() {
		taskComment := &TaskComment{}
		err := rows.Scan(
			&taskComment.ID,
			&taskComment.TaskID,
			&taskComment.Comment,
			&taskComment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if _, ok := taskCommentsMap[taskComment.ID]; !ok {
			taskCommentsMap[taskComment.ID] = taskComment
		}
	}

	taskComments := make([]TaskComment, 0, len(taskCommentsMap))
	for _, taskComment := range taskCommentsMap {
		taskComments = append(taskComments, *taskComment)
	}

	return taskComments, nil
}
