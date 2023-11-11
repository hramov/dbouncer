package v1

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hramov/dbouncer/internal"
	"log"
	"net"
	"sync"
	"time"
)

const (
	TCP = "tcp4"
)

var mu sync.RWMutex

type Server struct {
	port    int
	timeout time.Duration
	queryCh chan<- *internal.QueryRequest
	respCh  <-chan *internal.QueryResponse
	errCh   chan<- error
	apps    internal.Apps
}

func NewServer(port int, timeout time.Duration, queryCh chan<- *internal.QueryRequest, respCh <-chan *internal.QueryResponse, errCh chan<- error) (*Server, error) {
	return &Server{
		port:    port,
		timeout: timeout,
		queryCh: queryCh,
		respCh:  respCh,
		errCh:   errCh,
		apps:    make(internal.Apps),
	}, nil
}

func (s *Server) Serve(ctx context.Context) {
	var ln net.Listener
	var conn net.Conn
	var err error

	ln, err = net.Listen(TCP, fmt.Sprintf(":%d", s.port))
	if err != nil {
		s.errCh <- fmt.Errorf("cannot listen tcp, returning: %v\n", err)
		return
	}

	log.Printf("server started on %d\n", s.port)

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			if err != nil {
				s.errCh <- err
			}
			return
		default:
			conn, err = ln.Accept()
			if err != nil {
				s.errCh <- fmt.Errorf("cannot read from listener: %v\n", err)
				return
			}
			go s.handle(ctx, conn)
		}
	}
}

func (s *Server) Response(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-s.respCh:
			if !ok {
				return
			}

			if resp.Ctx.Err() != nil {
				return
			}

			mu.RLock()
			app := s.apps[resp.AppId]
			mu.RUnlock()
			err := s.send(ctx, app.Conn, resp)
			if err != nil {
				s.errCh <- fmt.Errorf("cannot send response: %v\n", err)
			}
		}
	}
}

func (s *Server) send(ctx context.Context, conn net.Conn, data *internal.QueryResponse) error {
	if conn == nil {
		return fmt.Errorf("cannot send response: conn is nil")
	}

	msg, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("cannot marshal response: %v\n", err)
	}
	_, err = conn.Write(msg)
	if err != nil {
		return fmt.Errorf("cannot write to conn: %v\n", err)
	}
	return nil
}

func (s *Server) parse(body []byte) (*internal.QueryRequest, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("empty body")
	}

	query := internal.QueryRequest{}
	err := json.Unmarshal(body, &query)
	if err != nil {
		return nil, err
	}
	return &query, nil
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	connCtx, cancel := context.WithCancel(ctx)

	appId := uuid.Nil

	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println(err.Error())
			cancel()
			return
		}

		go func(netData string, err error) {
			if err != nil {
				if appId != uuid.Nil {
					mu.Lock()
					app := s.apps[appId]
					if app.Conn != nil {
						app.Conn.Close()
					}
					delete(s.apps, appId)
					mu.Unlock()
				}
				return
			}
			var query *internal.QueryRequest
			query, err = s.parse([]byte(netData))
			if err != nil {
				s.errCh <- fmt.Errorf("cannot parse query: %v\n", err)
				errResp := &internal.QueryResponse{
					Id:     0,
					Kind:   "",
					Error:  true,
					Result: err.Error(),
				}
				err = s.send(ctx, conn, errResp)
				if err != nil {
					s.errCh <- fmt.Errorf("cannot send error: %v\n", err)
				}
				return
			}

			if query.AppId == uuid.Nil {
				var id uuid.UUID
				id, err = uuid.NewUUID()
				if err != nil {
					s.errCh <- fmt.Errorf("cannot generate uuid: %v\n", err)
				}
				app := &internal.App{
					Id:      id,
					Conn:    conn,
					QueryId: query.Id,
				}

				if app.QueryId == 0 {
					app.QueryId = 1
				}

				mu.Lock()
				s.apps[id] = *app
				mu.Unlock()
				query.AppId = id
				log.Println("creating new app with id", query.AppId, "and name", query.AppName)
			}

			mu.Lock()
			app := s.apps[query.AppId]
			app.QueryId++
			s.apps[query.AppId] = app
			mu.Unlock()

			query.Id = app.QueryId
			appId = query.AppId

			query.Ctx = connCtx
			s.queryCh <- query
		}(netData, err)
	}
}
