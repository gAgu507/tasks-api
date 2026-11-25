package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

// Implementation of UserStore interface

func (s *Postgres) CreateUser(ctx context.Context, email, hash string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, password_hash`, email, hash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // нарушение UNIQUE
		return User{}, ErrEmailTaken
	}
	return u, err
}

func (s *Postgres) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

// Implementation of TaskStore interface

func (store *Postgres) List(ctx context.Context) ([]Task, error) {
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

func (store *Postgres) Create(ctx context.Context, task Task) (Task, error) {
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

func (store *Postgres) Change(ctx context.Context, task Task) error {
	tag, err := store.pool.Exec(ctx, "UPDATE tasks SET title=$1, description=$2, completed=$3 WHERE id=$4", task.Title, task.Description, task.Completed, task.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (store *Postgres) Delete(ctx context.Context, id int) error {
	tag, err := store.pool.Exec(ctx, "DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (store *Postgres) Get(ctx context.Context, id int) (Task, error) {
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
