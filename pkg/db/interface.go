package db

import "database/sql"

type Type string

const (
	Postgres Type = "postgres"
	SQLite   Type = "sqlite"
	Mock     Type = "mock"
)

//Introducing our own interface to be able to add logging capabilities
//and make testing simpler (allows injection of mocks)
type Connection interface {
	QueryRow(query string, args ...interface{}) DataRow
	Query(query string, args ...interface{}) (DataRows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
	Close() error
	Type() Type
}

type ConnectionFactory interface {
	Init() error
	NewConnection() (Connection, error)
}

type DatabaseEntity interface {
	Table() string
	Marshaller() *EntityMarshaller
	New() DatabaseEntity
	Equal(other DatabaseEntity) bool
}

//DataRow introduces a interface which is implemented by sql.Row and sql.Rows
//to make both usable for retrieving raw data
type DataRow interface {
	Scan(dest ...interface{}) error
}

type DataRows interface {
	Scan(dest ...interface{}) error
	Next() bool
}