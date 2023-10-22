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
			fmt.Println(string(response.Result))
			appId = response.AppId
		}
	}()

	time.Sleep(1 * time.Second)

	var data []byte

	for i := 0; i < 5; i++ {
		msg := &internal.QueryRequest{
			Id:       1,
			AppId:    appId,
			Database: "postgres",
			Kind:     "query",
			Query:    "select * from pg_stat_activity",
			Params:   nil,
		}

		data, err = json.Marshal(msg)
		if err != nil {
			panic(err)
		}

		n, err = conn.Write(data)
		if err != nil {
			panic(err)
		}

		time.Sleep(2 * time.Second)
	}
}
