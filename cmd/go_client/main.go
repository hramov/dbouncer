package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hramov/dbouncer/internal"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp4", "127.0.0.1:2379")
	if err != nil {
		panic(err)
	}

	b := make([]byte, 4<<20)
	var appId = uuid.Nil

	var n int

	go func() {
		for {
			n, err = conn.Read(b)
			if err != nil {
				panic(err)
			}
			rawData := b[:n]

			var response internal.QueryResponse
			err = json.Unmarshal(rawData, &response)
			if err != nil {
				panic(err)
			}
			appId = response.AppId
			fmt.Println(response.Result)
		}
	}()

	time.Sleep(1 * time.Second)

	var data []byte

	go func() {
		for i := 0; i < 100; i++ {
			msg := &internal.QueryRequest{
				Id:       1,
				AppId:    appId,
				AppName:  "dbouncer",
				Database: "postgres",
				Kind:     "query",
				Query:    `select pg_sleep(2)`,
				Params:   nil,
			}

			data, err = json.Marshal(msg)
			if err != nil {
				panic(err)
			}

			data = append(data, '\n')

			n, err = conn.Write(data)
			if err != nil {
				panic(err)
			}

			fmt.Println("sent message", i)

			time.Sleep(10 * time.Millisecond)
		}
	}()

	for {
	}
}
