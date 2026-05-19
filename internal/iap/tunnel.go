package iap

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/oauth2"
	"nhooyr.io/websocket"
)

const (
	tunnelHost = "tunnel.cloudproxy.app"
	subprotocol = "relay.tunnel.cloudproxy.app"

	maxDataFrameSize = 16384

	tagConnectSuccessSID  = 0x0001
	tagReconnectSuccessACK = 0x0002
	tagData               = 0x0004
	tagACK                = 0x0007

	tagLen    = 2
	headerLen = 6 // tag(2) + length(4)
)

// TunnelConfig holds parameters for an IAP tunnel connection.
type TunnelConfig struct {
	Project    string
	Zone       string
	Instance   string
	Interface  string
	Port       int
	TokenSource oauth2.TokenSource
}

// Tunnel represents an active IAP tunnel.
type Tunnel struct {
	cfg       TunnelConfig
	conn      *websocket.Conn
	sid       []byte
	bytesSent uint64
	bytesRecv uint64
	bytesAcked uint64
	mu        sync.Mutex
}

// Listen starts a local TCP listener that proxies connections through the IAP tunnel.
// Returns the listener address (host:port).
func Listen(ctx context.Context, cfg TunnelConfig, localPort int) (net.Listener, error) {
	addr := fmt.Sprintf("localhost:%d", localPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listening on %s: %w", addr, err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			go handleConn(ctx, cfg, conn)
		}
	}()

	return ln, nil
}

func handleConn(ctx context.Context, cfg TunnelConfig, tcpConn net.Conn) {
	defer tcpConn.Close()

	t, err := connect(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "IAP tunnel connect error: %v\n", err)
		return
	}
	defer t.conn.Close(websocket.StatusNormalClosure, "done")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Proxy: WebSocket → TCP
	go func() {
		defer cancel()
		t.readLoop(ctx, tcpConn)
	}()

	// Proxy: TCP → WebSocket
	t.writeLoop(ctx, tcpConn)
}

func connect(ctx context.Context, cfg TunnelConfig) (*Tunnel, error) {
	if cfg.Interface == "" {
		cfg.Interface = "nic0"
	}

	token, err := cfg.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("obtaining token: %w", err)
	}

	u := url.URL{
		Scheme: "wss",
		Host:   tunnelHost,
		Path:   "/v4/connect",
	}
	q := u.Query()
	q.Set("project", cfg.Project)
	q.Set("zone", cfg.Zone)
	q.Set("instance", cfg.Instance)
	q.Set("interface", cfg.Interface)
	q.Set("port", fmt.Sprintf("%d", cfg.Port))
	q.Set("newWebsocket", "true")
	u.RawQuery = q.Encode()

	headers := http.Header{
		"Origin":        {"bot:iap-tunneler"},
		"Authorization": {fmt.Sprintf("Bearer %s", token.AccessToken)},
	}

	conn, _, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{
		Subprotocols: []string{subprotocol},
		HTTPHeader:   headers,
	})
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}

	// Increase read limit for large frames.
	conn.SetReadLimit(maxDataFrameSize + headerLen + 64)

	t := &Tunnel{cfg: cfg, conn: conn}

	// Wait for CONNECT_SUCCESS_SID.
	if err := t.waitForConnect(ctx); err != nil {
		conn.Close(websocket.StatusAbnormalClosure, err.Error())
		return nil, err
	}

	return t, nil
}

func (t *Tunnel) waitForConnect(ctx context.Context) error {
	_, msg, err := t.conn.Read(ctx)
	if err != nil {
		return fmt.Errorf("reading connect response: %w", err)
	}
	if len(msg) < headerLen {
		return fmt.Errorf("connect response too short: %d bytes", len(msg))
	}

	tag := binary.BigEndian.Uint16(msg[0:2])
	if tag != tagConnectSuccessSID {
		return fmt.Errorf("expected CONNECT_SUCCESS_SID (0x0001), got 0x%04x", tag)
	}

	dataLen := binary.BigEndian.Uint32(msg[2:6])
	if len(msg) < headerLen+int(dataLen) {
		return fmt.Errorf("connect response data truncated")
	}
	t.sid = msg[headerLen : headerLen+int(dataLen)]
	return nil
}

// readLoop reads DATA frames from the WebSocket and writes to the TCP connection.
func (t *Tunnel) readLoop(ctx context.Context, w io.Writer) {
	for {
		_, msg, err := t.conn.Read(ctx)
		if err != nil {
			return
		}
		if len(msg) < tagLen {
			continue
		}

		tag := binary.BigEndian.Uint16(msg[0:2])
		switch tag {
		case tagData:
			if len(msg) < headerLen {
				continue
			}
			dataLen := binary.BigEndian.Uint32(msg[2:6])
			if len(msg) < headerLen+int(dataLen) {
				continue
			}
			payload := msg[headerLen : headerLen+int(dataLen)]

			t.mu.Lock()
			t.bytesRecv += uint64(len(payload))
			shouldAck := (t.bytesRecv - t.bytesAcked) > 2*maxDataFrameSize
			t.mu.Unlock()

			if _, err := w.Write(payload); err != nil {
				return
			}

			if shouldAck {
				t.sendACK(ctx)
			}

		case tagACK:
			// Server acknowledging our sent bytes — no action needed for basic flow.
		}
	}
}

// writeLoop reads from the TCP connection and sends DATA frames over WebSocket.
func (t *Tunnel) writeLoop(ctx context.Context, r io.Reader) {
	buf := make([]byte, maxDataFrameSize)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if writeErr := t.sendData(ctx, buf[:n]); writeErr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (t *Tunnel) sendData(ctx context.Context, payload []byte) error {
	frame := make([]byte, headerLen+len(payload))
	binary.BigEndian.PutUint16(frame[0:2], tagData)
	binary.BigEndian.PutUint32(frame[2:6], uint32(len(payload)))
	copy(frame[headerLen:], payload)

	t.mu.Lock()
	t.bytesSent += uint64(len(payload))
	t.mu.Unlock()

	return t.conn.Write(ctx, websocket.MessageBinary, frame)
}

func (t *Tunnel) sendACK(ctx context.Context) {
	t.mu.Lock()
	recv := t.bytesRecv
	t.bytesAcked = recv
	t.mu.Unlock()

	frame := make([]byte, tagLen+8)
	binary.BigEndian.PutUint16(frame[0:2], tagACK)
	binary.BigEndian.PutUint64(frame[2:10], recv)

	t.conn.Write(ctx, websocket.MessageBinary, frame)
}
