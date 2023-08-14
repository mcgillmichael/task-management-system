-- Create the 'task' table
CREATE TABLE task (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Create the 'task_item' table (to store the list items associated with each task)
CREATE TABLE task_item (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL,
    item TEXT NOT NULL,
    FOREIGN KEY (task_id) REFERENCES task (id) ON DELETE CASCADE
);
