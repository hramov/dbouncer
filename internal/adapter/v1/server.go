package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hramov/dbouncer/internal"
	"log"
	"net"
	"time"
)

const (
	TCP = "tcp4"
)

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
	}, nil
}

func (s *Server) Serve(ctx context.Context) {
	var ln net.Listener
	var conn net.Conn
	var err error
	var n int

	ln, err = net.Listen(TCP, fmt.Sprintf(":%d", s.port))
	if err != nil {
		s.errCh <- fmt.Errorf("cannot listen tcp, returning: %v\n", err)
		return
	}

	log.Printf("server started on %d\n", s.port)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err = ln.Accept()
			if err != nil {
				s.errCh <- err
				return
			}

			var buf []byte

			n, err = conn.Read(buf)
			if err != nil {
				s.errCh <- fmt.Errorf("cannot read from listener: %v\n", err)
				continue
			}

			var query *internal.QueryRequest

			query, err = s.parse(buf[:n])
			if err != nil {
				s.errCh <- fmt.Errorf("cannot parse query: %v\n", err)

				errResp := &internal.QueryResponse{
					Id:     0,
					Kind:   "",
					Error:  true,
					Result: err,
				}
				err = s.send(ctx, conn, errResp)
				if err != nil {
					s.errCh <- fmt.Errorf("cannot send error: %v\n", err)
				}
				continue
			}

			log.Printf("received query: %v\n", query.Id)

			if query.AppId == uuid.Nil {
				var id uuid.UUID
				id, err = uuid.NewUUID()
				if err != nil {
					s.errCh <- fmt.Errorf("cannot generate uuid: %v\n", err)
				}
				app := &internal.App{
					Id:   id,
					Conn: conn,
				}
				s.apps[id] = *app
				query.AppId = id
				log.Println("creating new app with id", id)
			}
			s.queryCh <- query
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
			app := s.apps[resp.AppId]
			err := s.send(ctx, app.Conn, resp)
			if err != nil {
				s.errCh <- fmt.Errorf("cannot send response: %v\n", err)
			}
			log.Printf("sent response: %v\n", resp.Id)
		}
	}
}

func (s *Server) send(ctx context.Context, conn net.Conn, data any) error {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Printf("cannot marshal response: %v\n", err)
		return fmt.Errorf("cannot marshal response: %v\n", err)
	}

	_, err = conn.Write(msg)
	if err != nil {
		log.Printf("cannot write to conn: %v\n", err)
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
