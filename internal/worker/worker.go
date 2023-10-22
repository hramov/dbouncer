package worker

import (
	"context"
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
			log.Printf("worker %d received query: %v\n", w.id, query.Id)
			if !ok {
				return
			}
			resp, err := w.process(ctx, query)
			if err != nil {
				w.errCh <- err
				continue
			}
			log.Printf("worker %d processed query: %v\n", w.id, query.Id)
			w.respCh <- resp
		}
	}
}

func (w *Worker) process(ctx context.Context, query *internal.QueryRequest) (*internal.QueryResponse, error) {

	return nil, nil
}
