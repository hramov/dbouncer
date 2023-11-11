package worker

import (
	"context"
	"fmt"
	"github.com/hramov/dbouncer/internal"
	"github.com/hramov/dbouncer/pkg/storage"
	"log"
	"time"
)

type Worker struct {
	id            int
	queryCh       chan *internal.QueryRequest
	errCh         chan<- error
	respCh        chan<- *internal.QueryResponse
	statCh        chan map[string]int
	dbName        string
	db            internal.Storage
	queriesForApp map[string]int
	queriesAll    int
}

func NewWorker(id int, queryCh chan *internal.QueryRequest, errCh chan<- error, respCh chan<- *internal.QueryResponse, statCh chan map[string]int, dbName string, storage internal.Storage) (*Worker, error) {
	w := &Worker{
		id:            id,
		queryCh:       queryCh,
		errCh:         errCh,
		respCh:        respCh,
		statCh:        statCh,
		dbName:        dbName,
		db:            storage,
		queriesForApp: make(map[string]int),
	}
	return w, nil
}

func (w *Worker) Process(ctx context.Context) {
	log.Printf("worker %d (%s) started\n", w.id, w.dbName)

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.statCh <- w.queriesForApp

			for k, _ := range w.queriesForApp {
				w.queriesForApp[k] = 0
			}

		case query, ok := <-w.queryCh:
			if !ok {
				return
			}
			if query.Database != w.dbName {
				w.queryCh <- query
				continue
			}

			app, exists := w.queriesForApp[query.AppName]
			if !exists {
				w.queriesForApp[query.AppName] = 1
			} else {
				w.queriesForApp[query.AppName] = app + 1
			}

			w.queriesAll++

			resp, err := w.process(query)
			if err != nil {
				w.errCh <- err
				continue
			}

			log.Printf("worker %d processed query from app %s (all: %d)\n", w.id, query.AppName, w.queriesAll)
			w.respCh <- resp
		}
	}
}

func (w *Worker) process(query *internal.QueryRequest) (*internal.QueryResponse, error) {

	if query.Ctx.Err() != nil {
		return nil, query.Ctx.Err()
	}

	st := storage.GetStorage(query.Database)

	if st == nil {
		return nil, fmt.Errorf("unknown database: %s", query.Database)
	}

	var data []string
	var err error

	switch query.Kind {
	case "query":
		data, err = st.QueryTx(query.Ctx, query.Query, query.Params...)
	case "query_row":
		data, err = st.QueryRowTx(query.Ctx, query.Query, query.Params...)
	case "exec":
		data, err = st.ExecTx(query.Ctx, query.Query, query.Params...)
	}

	if err != nil {
		log.Println("data fetching error: ", err.Error())
		return &internal.QueryResponse{
			Ctx:    query.Ctx,
			Id:     query.Id,
			AppId:  query.AppId,
			Kind:   query.Kind,
			Error:  true,
			Result: err.Error(),
		}, nil
	}

	return &internal.QueryResponse{
		Ctx:    query.Ctx,
		Id:     query.Id,
		AppId:  query.AppId,
		Kind:   query.Kind,
		Error:  false,
		Result: data,
	}, nil
}
