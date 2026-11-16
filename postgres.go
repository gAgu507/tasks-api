package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func (store *PostgresStore) List(ctx context.Context) ([]Task, error) {
	rows, err := store.pool.Query(ctx, "SELECT id, title, description, completed FROM tasks ORDER BY id")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (store *PostgresStore) Create(ctx context.Context, task Task) (Task, error) {
	err := store.pool.QueryRow(ctx,
		`INSERT INTO tasks (title, description, completed)
		 VALUES ($1, $2, $3)
		 RETURNING id, title, description, completed`,
		task.Title, task.Description, task.Completed,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Completed)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (store *PostgresStore) Change(ctx context.Context, task Task) error {
	tag, err := store.pool.Exec(ctx, "UPDATE tasks SET title=$1, description=$2, completed=$3 WHERE id=$4", task.Title, task.Description, task.Completed, task.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (store *PostgresStore) Delete(ctx context.Context, id int) error {
	tag, err := store.pool.Exec(ctx, "DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (store *PostgresStore) Get(ctx context.Context, id int) (Task, error) {
	var task Task
	row := store.pool.QueryRow(ctx, "SELECT id, title, description, completed FROM tasks WHERE id=$1", id)
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.Completed)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}
	return task, nil
}
