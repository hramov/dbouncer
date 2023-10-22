package worker

import (
	"context"
	"fmt"
	"github.com/hramov/dbouncer/internal"
	"github.com/hramov/dbouncer/internal/config"
	"github.com/hramov/dbouncer/pkg/storage"
	"log"
)

type Worker struct {
	id      int
	queryCh <-chan *internal.QueryRequest
	respCh  chan<- *internal.QueryResponse
	errCh   chan<- error
	dbs     map[int]*internal.Storage
}

func NewWorker(id int, queryCh <-chan *internal.QueryRequest, errCh chan<- error, respCh chan<- *internal.QueryResponse, storages []config.Storage) (*Worker, error) {
	w := &Worker{
		id:      id,
		queryCh: queryCh,
		errCh:   errCh,
		respCh:  respCh,
		dbs:     make(map[int]*internal.Storage),
	}

	for i, v := range storages {
		db, err := storage.New(v.Name, v.Dsn)
		if err != nil {
			return nil, err
		}
		w.dbs[i] = &db
	}

	return w, nil
}

func (w *Worker) Process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case query, ok := <-w.queryCh:
			if !ok {
				return
			}
			resp, err := w.process(ctx, query)
			if err != nil {
				w.errCh <- err
				continue
			}
			log.Printf("worker %d processed query: %d from app %s\n", w.id, query.Id, query.AppId)
			w.respCh <- resp
		}
	}
}

func (w *Worker) process(ctx context.Context, query *internal.QueryRequest) (*internal.QueryResponse, error) {
	st := storage.GetStorage(query.Database)

	if st == nil {
		return nil, fmt.Errorf("unknown database: %s", query.Database)
	}

	var data []string
	var err error

	switch query.Kind {
	case "query":
		data, err = st.QueryTx(ctx, query.Query, query.Params...)
	case "query_row":
		data, err = st.QueryRowTx(ctx, query.Query, query.Params...)
	case "exec":
		data, err = st.ExecTx(ctx, query.Query, query.Params...)
	}

	if err != nil {
		return &internal.QueryResponse{
			Id:     query.Id,
			AppId:  query.AppId,
			Kind:   query.Kind,
			Error:  true,
			Result: err.Error(),
		}, nil
	}

	return &internal.QueryResponse{
		Id:     query.Id,
		AppId:  query.AppId,
		Kind:   query.Kind,
		Error:  false,
		Result: data,
	}, nil
}
