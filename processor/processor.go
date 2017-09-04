package processor

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
)

type Processor struct {
	cmd         *exec.Cmd
	concurrency int

	enc      *json.Encoder
	inCloser io.Closer
	dec      *json.Decoder
	err      io.ReadCloser

	requestChannels []chan Response
	slots           chan int
}

type Request struct {
	ID   int
	Data json.RawMessage
}

type Response struct {
	ID      int
	IsError bool
	Data    json.RawMessage
}

func New(command string, concurrency int) *Processor {
	cmd := exec.Command(command)

	return &Processor{
		cmd:         cmd,
		concurrency: concurrency,
	}
}

func (p *Processor) Start() error {
	out, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	in, err := p.cmd.StdinPipe()
	if err != nil {
		return err
	}
	p.cmd.Stderr = os.Stderr
	/*	outerr, err := p.cmd.StderrPipe()
		if err != nil {
			return err
		}*/
	err = p.cmd.Start()
	if err != nil {
		return err
	}

	p.requestChannels = make([]chan Response, p.concurrency)
	p.slots = make(chan int)
	for i := range p.requestChannels {
		p.requestChannels[i] = make(chan Response)
		go func(i int) { p.slots <- i }(i)
	}
	p.enc = json.NewEncoder(in)
	p.inCloser = in
	p.dec = json.NewDecoder(out)
	//	p.err = outerr
	go p.handle()
	return nil
}

func (p *Processor) Stop() {
	p.inCloser.Close()
	p.cmd.Wait()
}

func (p *Processor) handle() {
	for {
		var res Response
		err := p.dec.Decode(&res)
		if err == io.EOF {
			return
		}
		if err != nil {
			p.inCloser.Close()
			return
		}
		if res.ID >= 0 && res.ID < len(p.requestChannels) {
			p.requestChannels[res.ID] <- res
		}
	}
}

func (p Processor) Process(data json.RawMessage) json.RawMessage {
	slot := <-p.slots
	req := &Request{
		ID:   slot,
		Data: data,
	}
	p.enc.Encode(req)
	res := <-p.requestChannels[slot]

	go func(i int) { p.slots <- i }(slot)

	if res.IsError {
		//todo: log error to something
	}

	return res.Data
}
