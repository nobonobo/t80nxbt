package procon

import (
	"context"
	"fmt"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"sync"
	"time"
)

const (
	StateCrashed      = "crashed"
	StateConnected    = "connected"
	StateDisconnected = "disconnected"
)

type State struct {
	State  string  `json:"state"`
	Errors *string `json:"errors,omitempty"`
}

type codec struct {
	rpc.ClientCodec
	in  *io.PipeReader
	out *io.PipeWriter
}

func (c *codec) WriteRequest(res *rpc.Request, v interface{}) error {
	//log.Println("req:", res, v)
	if err := c.ClientCodec.WriteRequest(res, v); err != nil {
		return err
	}
	return nil
	//_, err := c.out.Write([]byte("\n"))
	//return err
}

func (c *codec) Close() error {
	c.in.Close()
	c.out.Close()
	return c.ClientCodec.Close()
}

type Client struct {
	*rpc.Client
	mu  sync.Mutex
	cmd *exec.Cmd
}

func New() *Client {
	return &Client{}
}

func (c *Client) Start(ctx context.Context, script string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd != nil {
		if c.cmd.Process != nil {
			c.cmd.Process.Kill()
		}
	}
	c.cmd = exec.Command("python3", script)
	r, _ := c.cmd.StdoutPipe()
	w, _ := c.cmd.StdinPipe()
	cc := jsonrpc.NewClientCodec(
		struct {
			io.Reader
			io.Writer
			io.Closer
		}{
			Reader: r, //io.TeeReader(rout, os.Stderr),
			Writer: w, //io.MultiWriter(win, os.Stderr),
			Closer: w,
		},
	)
	c.Client = rpc.NewClientWithCodec(cc)
	return c.cmd.Start()
}

func (c *Client) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd != nil {
		if c.cmd.Process != nil {
			c.cmd.Process.Kill()
		}
	}
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Call("disconnect", nil, nil)
	for r := 0; r < 3; r++ {
		if err := c.Call("connect", nil, nil); err != nil {
			return err
		}
		res := State{}
		score := 0
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			if err := c.Call("state", nil, &res); err != nil {
				return err
			}
			if res.State != "connected" {
				break
			}
			score++
			if score == 5 {
				return nil
			}
		}
		if res.State == "crashed" {
			continue
		}
	}
	return fmt.Errorf("connect failed: retry exceeded")
}

func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Call("disconnect", nil, nil)
}

func (c *Client) Input(in Input) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Call("input", in, nil)
}

func (c *Client) State() (*State, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := &State{}
	if err := c.Call("state", nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}
