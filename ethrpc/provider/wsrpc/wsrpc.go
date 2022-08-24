package wsrpc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/gorilla/websocket"
	"github.com/howjmay/multicall/errors"
	"github.com/howjmay/multicall/ethrpc/jsonrpc"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 60 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type WSProvider struct {
	url           *url.URL
	retry         bool
	client        *websocket.Conn
	mu            sync.Mutex
	send          chan []byte
	requests      map[string]chan *jsonrpc.JSONRPCResponse
	subscriptions map[string]chan *json.RawMessage
	cancel        chan struct{}
	dead          bool
	deadMu        sync.Mutex
}

// Start connects to parity and starts listening for notifications
func (p *WSProvider) Start() error {
	err := p.connect()
	if err != nil {
		return err
	}
	p.deadMu.Lock()
	p.dead = false
	p.deadMu.Unlock()
	go p.receivePump()
	go p.sendPump()
	return nil
}

// Stop closes the websocket connection
func (p *WSProvider) Stop() {
	p.fatality()
}

// CallRaw calls a RPC method and returns the raw result
func (p *WSProvider) CallRaw(method string, params ...interface{}) ([]byte, error) {
	receiver := make(chan *jsonrpc.JSONRPCResponse)
	err := p.sendRequest(receiver, method, params)
	if err != nil {
		return nil, fmt.Errorf("call: %s", err)
	}

	var resp *jsonrpc.JSONRPCResponse
	select {
	case resp = <-receiver:
		break
	case <-p.cancel:
		return nil, errors.Err_ConnectionClosed
	}

	return resp.Raw, nil
}

// Call calls a RPC method and returns corresponding object
func (p *WSProvider) Call(result interface{}, method string, params ...interface{}) error {
	receiver := make(chan *jsonrpc.JSONRPCResponse)
	err := p.sendRequest(receiver, method, params)
	if err != nil {
		return fmt.Errorf("call: %s", err)
	}

	var resp *jsonrpc.JSONRPCResponse
	select {
	case resp = <-receiver:
		break
	case <-p.cancel:
		return errors.Err_ConnectionClosed
	}

	if resp.IsResultNull() {
		return errors.Err_Null
	}

	if resp.Error != nil {
		switch resp.Error.Code {
		case -32015: // VM execution error
			err := errors.Err_VMExecutionError.(*errors.RpcError)
			err.Code = resp.Error.Code
			err.Details = resp.Error.Data
			return err
		default:
			return errors.New(resp.Error.Message, resp.Error.Code, resp.Error.Data)
		}
	}

	err = json.Unmarshal(resp.Result, &result)
	if err != nil {
		return err
	}
	return nil
}

// Subscribe creates a subscription to event using method
func (p *WSProvider) Subscribe(receiver chan *json.RawMessage, method string, event string, params ...interface{}) error {
	var subscriptionID string
	pa := append([]interface{}{}, event)
	pa = append(pa, params...)

	err := p.Call(&subscriptionID, method, pa...)
	if err != nil {
		return fmt.Errorf("subscription creation: %s", err)
	}

	p.mu.Lock()
	p.subscriptions[subscriptionID] = receiver
	p.mu.Unlock()

	return nil
}

// unsubscribe closes a subscription
// TODO this should be available to the outside world as well, but subscriptionID is not public
func (p *WSProvider) unsubscribe(subscriptionID string) {
	p.mu.Lock()
	close(p.subscriptions[subscriptionID])
	delete(p.subscriptions, subscriptionID)
	p.mu.Unlock()
}

func (p *WSProvider) sendRequest(receiver chan *jsonrpc.JSONRPCResponse, method string, params []interface{}) error {
	p.deadMu.Lock()
	dead := p.dead
	p.deadMu.Unlock()
	if dead {
		return errors.Err_ConnectionClosed
	}

	id := strconv.FormatInt(rand.Int63(), 16)
	request, err := jsonrpc.EncodeClientRequest(method, params, id)
	if err != nil {
		return err
	}

	// ensure only one write at a time
	p.mu.Lock()
	p.requests[id] = receiver
	p.mu.Unlock()

	// sending request to write pump
	p.send <- request
	return nil
}

func (p *WSProvider) connect() error {
	r := rate.Every(time.Minute)
	limiter := rate.NewLimiter(r, 1)
	log.Debugf("connecting to server on %s", p.url.String())
	for {
		c, _, err := websocket.DefaultDialer.Dial(p.url.String(), nil)
		if err != nil {
			if limiter.Allow() {
				log.Warnf("error connecting to server: %s ", err)
			}
			if p.retry {
				time.Sleep(time.Second)
				continue
			} else {
				return err
			}

		}
		logrus.Debug("connected to server over websocket")
		p.client = c

		// TODO disable for now check https://github.com/gorilla/websocket/issues/355
		//c.SetReadDeadline(time.Now().Add(pongWait))
		c.SetPongHandler(p.handlePong)
		//connected and subscribed, leave
		break
	}

	return nil
}

func (p *WSProvider) handlePong(string) error {
	//p.client.SetReadDeadline(time.Now().Add(pongWait))
	return nil
}

func (p *WSProvider) receivePump() {
	for {
		_, message, err := p.client.ReadMessage()
		if err != nil {
			log.Debugf("message read error: %s", err)
			p.fatality()
			return
		}
		msg, err := jsonrpc.DecodeResponse(message)
		if err != nil {
			log.Warnf("decode rpc message: %s", err)
		}
		// TODO this has the potential to block the queue lower down the line
		// move to go routine, add some auto expiry for channels that never got called
		// look into how a channel might get blocked and never get called
		go p.handleMessage(msg)

	}
}

func (p *WSProvider) sendPump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.client.Close()
	}()
	for {
		select {
		case message, ok := <-p.send:
			p.client.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				p.client.WriteMessage(websocket.CloseMessage, []byte{})
				log.Warn("hub close the channel. investigate!!")
				p.fatality()
				return
			}

			w, err := p.client.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Warnf("websocket writer: %s", err)
				p.fatality()
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				log.Warnf("websocket connection closed: %s", err)
				p.fatality()
				return
			}
		case <-ticker.C:
			p.client.SetWriteDeadline(time.Now().Add(writeWait))
			err := p.client.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Warnf("set write deadline: %s", err)
				p.fatality()
				return
			}
		case <-p.cancel:
			// ending the misery
			return
		}
	}
}

func (p *WSProvider) handleMessage(msg *jsonrpc.JSONRPCResponse) {
	switch {
	case msg.IsNotification():
		if !strings.HasSuffix(msg.Method, "_subscription") {
			log.Warn(fmt.Sprint("dropping non-subscription message: ", msg))
			return
		}
		var notification jsonrpc.JSONRPCNotification

		if err := json.Unmarshal(msg.Params, &notification); err != nil {
			log.Warnf(fmt.Sprint("dropping invalid subscription message: ", msg))
			return
		}
		id, err := notification.ValidID()
		if err != nil {
			log.Warn("notification json id", err)
		}

		p.mu.Lock()
		c := p.subscriptions[id]
		p.mu.Unlock()

		c <- &notification.Result

	case msg.IsResponse():
		id, err := msg.ValidID()
		if err != nil {
			log.Warn("response json id", err)
		}

		p.mu.Lock()
		c := p.requests[id]
		delete(p.requests, id)
		p.mu.Unlock()

		c <- msg

	default:
		log.Warnf("message not handled: %s", msg.String())
	}
}

func (p *WSProvider) fatality() {
	p.deadMu.Lock()
	if !p.dead {
		p.dead = true

		// kill any ongoing requests
		close(p.cancel)
		// kill any subscriptions
		for k := range p.subscriptions {
			p.unsubscribe(k)
		}
	}
	p.deadMu.Unlock()

	_ = p.client.Close()
}

// New creates a new WSProvider struct
func New(u string, retry bool) (*WSProvider, error) {
	// fail early. bail out if the url is invalid
	url, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return &WSProvider{
			url:           url,
			retry:         retry,
			send:          make(chan []byte),
			requests:      make(map[string]chan *jsonrpc.JSONRPCResponse),
			subscriptions: make(map[string]chan *json.RawMessage),
			cancel:        make(chan struct{}),
			dead:          true,
		},
		nil
}
