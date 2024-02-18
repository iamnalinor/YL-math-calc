package db

import (
	"fmt"
	"maps"
	"math-calc/internal/operation"
	"sync"
)

type Database struct {
	storage map[operation.ID]operation.Operation
	mx      sync.RWMutex

	// UpdatingMutex can be used by third party to lock the database for updates.
	UpdatingMutex sync.Mutex
}

func New() (*Database, error) {
	return &Database{
		storage: make(map[operation.ID]operation.Operation),
	}, nil
}

func (d *Database) Create(op operation.Operation) (operation.ID, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	newId := operation.ID(len(d.storage) + 1)

	// in case of collision
	for _, ok := d.storage[newId]; ok; {
		newId++
	}
	op.Id = newId

	d.storage[newId] = op

	return newId, nil
}

func (d *Database) Get(id operation.ID) (operation.Operation, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	if op, ok := d.storage[id]; ok {
		return op, nil
	}
	return operation.Operation{}, fmt.Errorf("operation with id %d not found", id)
}

func (d *Database) Update(op operation.Operation) error {
	d.mx.Lock()
	defer d.mx.Unlock()

	if _, ok := d.storage[op.Id]; ok {
		d.storage[op.Id] = op
		return nil
	}
	return fmt.Errorf("operation with id %d not found", op.Id)
}

func (d *Database) All() (map[operation.ID]operation.Operation, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	duplicate := make(map[operation.ID]operation.Operation, len(d.storage))
	maps.Copy(d.storage, duplicate)
	return duplicate, nil
}
