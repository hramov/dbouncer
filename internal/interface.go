package internal

import (
	"context"
	"github.com/google/uuid"
	"net"
)

type Storage interface {
	QueryTx(ctx context.Context, query string, args ...interface{}) ([]string, error)
	QueryRowTx(ctx context.Context, query string, args ...interface{}) ([]string, error)
	ExecTx(ctx context.Context, query string, args ...interface{}) ([]string, error)
}

type QueryKind string

const (
	Query    QueryKind = "query"
	QueryRow QueryKind = "query_row"
	Exec     QueryKind = "exec"
)

type QueryRequest struct {
	Ctx      context.Context `json:"-"`
	Id       int             `json:"id"`
	AppId    uuid.UUID       `json:"app_id"`
	AppName  string          `json:"app_name"`
	Database string          `json:"database"`
	Kind     QueryKind       `json:"kind"`
	Query    string          `json:"query"`
	Params   []any           `json:"params"`
}

type QueryResponse struct {
	Ctx    context.Context `json:"-"`
	Id     int             `json:"id"`
	AppId  uuid.UUID       `json:"app_id"`
	Kind   QueryKind       `json:"kind"`
	Error  bool            `json:"error"`
	Result any             `json:"result"`
}

type App struct {
	Id      uuid.UUID
	Conn    net.Conn
	QueryId int
}

type Apps map[uuid.UUID]App
