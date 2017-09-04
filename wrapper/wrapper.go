package main

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"sync"
)

type Request struct {
	ID   int
	Data json.RawMessage
}

type Response struct {
	ID      int
	IsError bool
	Data    interface{}
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	res := responseWorker(enc)
	wg := &sync.WaitGroup{}

	for {
		req := Request{}
		err := dec.Decode(&req)
		if err != nil {
			log.Printf("Error decoding request: %s", err)
			break
		}
		wg.Add(1)
		go handleReq(req, res, wg)
	}
	wg.Wait()
}

func handleReq(req Request, res chan Response, wg *sync.WaitGroup) {
	resp := Response{
		ID: req.ID,
	}
	log.Println(req)
	defer func() {
		err := recover()
		if err != nil {
			resp.IsError = true
			resp.Data = err
			res <- resp
		}
		wg.Done()
	}()
	f := reflect.ValueOf(Handler)
	argType := f.Type().In(0)
	arg := reflect.New(argType)
	json.Unmarshal(req.Data, arg.Interface())
	ret := f.Call([]reflect.Value{arg.Elem()})
	if len(ret) == 1 {
		resp.Data = ret[0].Interface()
	}
	if len(ret) == 2 {
		if !ret[1].IsNil() {
			resp.Data = ret[1].Interface()
			resp.IsError = true
		} else {
			resp.Data = ret[1].Interface()
		}
	}
	res <- resp
}

func responseWorker(enc *json.Encoder) chan Response {
	c := make(chan Response)

	go func() {
		for res := range c {
			err := enc.Encode(res)
			if err != nil {
				log.Fatalf("Error encoding response: %s", err)
			}
		}
	}()
	return c
}
