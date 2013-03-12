package socketio

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	buf      bytes.Buffer
	dec      *Decoder
	enc      *Encoder
	sid      string
	endpoint string
	ws       *websocket.Conn
	mutex    sync.Mutex
	id       int
	timeout  time.Duration
}

func (c *Client) String() string {
	return c.sid
}

func (c *Client) Emit(ack bool, name string, args ...interface{}) error {
	c.mutex.Lock()
	id := c.id
	c.id++
	c.mutex.Unlock()
	return c.Send(&event{Name: name, Args: args, ack: ack, id: id})
}

func (c *Client) Receive(msg *Message) (err error) {
	c.ws.SetDeadline(time.Now().Add(c.timeout))
	var incoming string
	for {
		if err = c.dec.Decode(msg); err == io.EOF {
			if err = websocket.Message.Receive(c.ws, &incoming); err != nil {
				return
			}
			c.dec.Write([]byte(incoming))
			continue
		} else if err != nil {
			break
		}

		switch msg.typ {
		case MessageHeartbeat:
			Log.debug(c, " client: received heartbeat: ", msg.Inspect())
			c.Send(heartbeat(0))

		case MessageDisconnect:
			Log.info(c, " client: received disconnect: ", msg.Inspect())
			c.ws.Close()
			return io.EOF

		case MessageConnect:
			return

		case MessageError, MessageACK, MessageNOOP:
			Log.warn(c, " client: (TODO) ", msg.Inspect())

		case MessageEvent, MessageText, MessageJSON:
			if msg.id > 0 && !msg.ack {
				Log.debug(c, " client: automatically acking: ", msg.Inspect())
				c.Send(&ack{id: msg.id})
			}
			return

		default:
			Log.warn(c, " client: unknown message type: ", msg.Inspect())
		}
	}

	c.dec.Reset()
	return
}

func (c *Client) Reply(m *Message, a ...interface{}) error {
	ack := &ack{
		id:   m.id,
		data: a,
	}
	if len(a) > 0 {
		ack.event = true
	}
	return c.Send(ack)
}

func (c *Client) Close() error {
	c.Send(disconnect(""))
	return c.ws.Close()
}

func (c *Client) Send(data interface{}) (err error) {
	c.ws.SetDeadline(time.Now().Add(c.timeout))
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.buf.Reset()
	if err = c.enc.Encode(&c.buf, []interface{}{data}); err != nil {
		return
	}
	return websocket.Message.Send(c.ws, c.buf.String())
}

func (c *Client) SetDeadline(t time.Time) {
	c.ws.SetDeadline(t)
}

func Dial(url_, origin string, timeout time.Duration) (c *Client, err error) {
	var body []byte
	var r *http.Response

	if r, err = http.Get(fmt.Sprintf("%s%d", url_, ProtocolVersion)); err != nil {
		return
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return nil, errors.New("invalid status: " + r.Status)
	}
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		return
	}
	parts := strings.SplitN(string(body), ":", 4)
	if len(parts) != 4 {
		return nil, errors.New("invalid handshake: " + string(body))
	}
	if !strings.Contains(parts[3], "websocket") {
		return nil, errors.New("server does not support websockets")
	}

	c = &Client{dec: &Decoder{}, enc: &Encoder{}, id: 1, timeout: timeout}
	c.sid = parts[0]
	wsurl := "ws" + url_[4:]
	if c.ws, err = websocket.Dial(fmt.Sprintf("%s%d/websocket/%s", wsurl, ProtocolVersion, c.sid), "", origin); err != nil {
		return
	}

	var msg Message
	if err = c.Receive(&msg); err != nil {
		c.ws.Close()
		return
	}
	if msg.Type() != MessageConnect {
		c.ws.Close()
		err = errors.New("unexpected connect message: " + msg.Inspect())
	}
	return
}
