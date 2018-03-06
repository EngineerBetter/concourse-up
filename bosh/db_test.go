package bosh

import (
	"database/sql"
	"errors"
)

type fakeOpener map[string]*sql.DB

func (f fakeOpener) Open(name string) (*sql.DB, error) {
	db, ok := f[name]
	if !ok {
		return nil, errors.New("database not found")
	}
	return db, nil
}

func (f fakeOpener) Close() error { return nil }
