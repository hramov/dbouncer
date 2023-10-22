package internal

import (
	"context"
	"github.com/google/uuid"
	"net"
)

type Storage interface {
	QueryTx(ctx context.Context, query string, args ...interface{}) ([]byte, error)
	QueryRowTx(ctx context.Context, query string, args ...interface{}) ([]byte, error)
	ExecTx(ctx context.Context, query string, args ...interface{}) ([]byte, error)
}

type QueryKind string

const (
	Query    QueryKind = "query"
	QueryRow QueryKind = "query_row"
	Exec     QueryKind = "exec"
)

type QueryRequest struct {
	Id       int       `json:"id"`
	AppId    uuid.UUID `json:"app_id"`
	Database string    `json:"database"`
	Kind     QueryKind `json:"kind"`
	Query    string    `json:"query"`
	Params   []any     `json:"params"`
}

type QueryResponse struct {
	Id     int       `json:"id"`
	AppId  uuid.UUID `json:"app_id"`
	Kind   QueryKind `json:"kind"`
	Error  bool      `json:"error"`
	Result []byte    `json:"result"`
}

type App struct {
	Id      uuid.UUID
	Conn    net.Conn
	QueryId int
}

type Apps map[uuid.UUID]App
