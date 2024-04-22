package db

import (
	"database/sql"
	"fmt"
	"math-calc/internal/operation"
	_ "modernc.org/sqlite"
	"sync"
	"time"
)

const SCHEMA = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username UNIQUE NOT NULL,
    password_salt NOT NULL,
    password_hash NOT NULL
);

CREATE TABLE IF NOT EXISTS operations (
    id INTEGER PRIMARY KEY,
    owner_id INTEGER NOT NULL,
    operator TEXT NOT NULL ,
    state INTEGER NOT NULL ,
    created_time TEXT NOT NULL ,
    finished_time TEXT ,
    "left" REAL,
    "right" REAL,
    left_operation_id INTEGER,
    right_operation_id INTEGER,
    result REAL,
    error TEXT,
    expression TEXT
);
`

type SqliteDatabase struct {
	conn *sql.DB
	mx   sync.RWMutex

	// UpdatingMutex can be used by third party to lock the database for updates.
	UpdatingMutex sync.Mutex
}

func NewSqlite(filename string) (*SqliteDatabase, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(SCHEMA)
	if err != nil {
		return nil, err
	}

	return &SqliteDatabase{
		conn: db,
	}, nil
}

func (d *SqliteDatabase) Create(op operation.Operation) (operation.ID, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	op.CreatedTime = time.Now()
	op.FinishedTime = time.Unix(0, 0)
	op.State = operation.StateCreated

	var q = `
	INSERT INTO operations (owner_id, operator, state, created_time, finished_time, left, right, left_operation_id, right_operation_id, expression, result, error) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := d.conn.Exec(q, op.OwnerID, op.Op, op.State, op.CreatedTime.Format(time.RFC3339), op.FinishedTime.Format(time.RFC3339), op.Left, op.Right, op.LeftOperationID, op.RightOperationID, op.Expression, 0, "")
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	op.Id = operation.ID(id)
	return operation.ID(id), nil
}

func (d *SqliteDatabase) Get(id operation.ID) (operation.Operation, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	var q = `
        SELECT * FROM operations WHERE id = ?
	`
	row := d.conn.QueryRow(q, id)
	var op operation.Operation
	createdTime := ""
	finishedTime := ""
	err := row.Scan(&op.Id, &op.OwnerID, &op.Op, &op.State, &createdTime, &finishedTime, &op.Left, &op.Right, &op.LeftOperationID, &op.RightOperationID, &op.Result, &op.Error, &op.Expression)
	if err != nil {
		return operation.Operation{}, fmt.Errorf("operation with id %d not found", id)
	}
	op.CreatedTime, err = time.Parse(time.RFC3339, createdTime)
	if err != nil {
		return operation.Operation{}, err
	}
	op.FinishedTime, err = time.Parse(time.RFC3339, finishedTime)
	if err != nil {
		return operation.Operation{}, err
	}
	return op, nil
}

func (d *SqliteDatabase) Update(op operation.Operation) error {
	d.mx.Lock()
	defer d.mx.Unlock()

	var q = `
	UPDATE operations SET operator = ?, state = ?, created_time = ?, finished_time = ?, left = ?, right = ?, left_operation_id = ?, right_operation_id = ?, result = ?, error = ?, expression = ? WHERE id = ?
	`
	res, err := d.conn.Exec(q, op.Op, op.State, op.CreatedTime.Format(time.RFC3339), op.FinishedTime.Format(time.RFC3339), op.Left, op.Right, op.LeftOperationID, op.RightOperationID, op.Result, op.Error, op.Expression, op.Id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("operation with id %d not found", op.Id)
	}
	return nil
}

func (d *SqliteDatabase) All() (map[operation.ID]operation.Operation, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	var q = `
	SELECT * FROM operations
	`
	rows, err := d.conn.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ops := make(map[operation.ID]operation.Operation)
	for rows.Next() {
		var op operation.Operation
		createdTime := ""
		finishedTime := ""
		err := rows.Scan(&op.Id, &op.OwnerID, &op.Op, &op.State, &createdTime, &finishedTime, &op.Left, &op.Right, &op.LeftOperationID, &op.RightOperationID, &op.Result, &op.Error, &op.Expression)
		if err != nil {
			return nil, err
		}
		op.CreatedTime, err = time.Parse(time.RFC3339, createdTime)
		if err != nil {
			return nil, err
		}
		op.FinishedTime, err = time.Parse(time.RFC3339, finishedTime)
		if err != nil {
			return nil, err
		}
		ops[op.Id] = op
	}
	return ops, nil
}

type User struct {
	ID           int
	Username     string
	PasswordSalt string
	PasswordHash string
}

func (d *SqliteDatabase) GetUserByID(id int) (User, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	var q = `
	SELECT * FROM users WHERE id = ?
	`
	row := d.conn.QueryRow(q, id)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordSalt, &user.PasswordHash)
	if err != nil {
		return User{}, fmt.Errorf("user with id %d not found", id)
	}
	return user, nil
}

func (d *SqliteDatabase) GetUserByUsername(username string) (User, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	var q = `
	SELECT * FROM users WHERE username = ?
	`
	row := d.conn.QueryRow(q, username)
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordSalt, &user.PasswordHash)
	if err != nil {
		return User{}, fmt.Errorf("user with username %s not found", username)
	}
	return user, nil
}

func (d *SqliteDatabase) CreateUser(username, passwordSalt, passwordHash string) (int, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	var q = `
	INSERT INTO users (username, password_salt, password_hash) VALUES (?, ?, ?)
	`
	result, err := d.conn.Exec(q, username, passwordSalt, passwordHash)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (d *SqliteDatabase) Close() error {
	return d.conn.Close()
}
