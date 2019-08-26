package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// TODO(netux): implement something similar to this
// https://github.com/gorilla/websocket/tree/master/examples/chat

type wsConn struct {
	*websocket.Conn
	ctx       context.Context
	user      *User
	sendQueue chan interface{}
}

func (conn *wsConn) queue(msg interface{}) {
	conn.sendQueue <- msg
}

type wsMessageType string
type wsMessage struct {
	Type wsMessageType `json:"type"`
}

func withType(t wsMessageType) wsMessage {
	return wsMessage{t}
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  MaxWebsocketReadBufferSize,
	WriteBufferSize: MaxWebsocketSendBufferSize,
}

// Connections is an array of all active websocket connections.
var Connections = make(map[string]*wsConn)

func getReqIP(r *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	return host, err
}

func getReqPxlsToken(r *http.Request) (string, error) {
	pxlsTokenCookie, err := r.Cookie("pxls-token")
	if err != nil {
		if err != http.ErrNoCookie {
			return "", fmt.Errorf("cannot get token from request: %v", err)
		}
		return "", &NotFoundError{"pxls token cookie not found"}
	}

	return pxlsTokenCookie.Value, nil
}

func newUserByIP(ip, ua string) (*User, error) {
	dbUser, err := App.DB.CreateUser("-snip-", UserLogin{"ip", ip}, ip, ua)
	if err != nil {
		return nil, fmt.Errorf("cannot create user in database with IP %s: %v", ip, err)
	}

	u, err := App.Users.MakeAndAdd(dbUser, ip)
	return u, err
}

func getReqUser(r *http.Request) (u *User, err error) {
	ua := r.UserAgent()

	ip, err := getReqIP(r)
	if err != nil {
		return nil, fmt.Errorf("cannot get host part of IP address \"%s\": %v", r.RemoteAddr, err)
	}

	token, err := getReqPxlsToken(r)
	if err != nil && !IsNotFoundError(err) {
		return nil, err
	}

	/// Get user by IP:
	if App.Conf.GetBoolean("oauth.useIp") && token == "" {
		u, ok := App.Users.GetByTokenOrIP(ip)
		if ok {
			// found user in cache.
			return u, nil
		}

		dbUser, err := App.DB.GetUserByLogin(UserLogin{"ip", ip})
		if err != nil {
			if !IsNotFoundError(err) {
				return nil, err
			}

			u, err := newUserByIP(ip, ua)
			return u, err
		}

		u, err = App.Users.MakeAndAdd(dbUser, ip)
		if err != nil {
			return nil, err
		}

		return u, nil
	}

	/// Get user by token:
	if token == "" {
		// no token => no user
		return nil, nil
	}

	u, ok := App.Users.GetByTokenOrIP(token)
	if ok {
		// found user in cache.
		return u, nil
	}

	dbUser, err := App.DB.GetUserByToken(token)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch user with IP %s auth data from database: %v", ip, err)
	}

	u, err = App.Users.MakeAndAdd(dbUser, token)
	if err != nil {
		return nil, fmt.Errorf("cannot make and add user to user list: %v", err)
	}

	return u, nil
}

func upgradeSocket(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	conn.SetCloseHandler(func(_ int, _ string) error {
		cancel()
		return nil
	})

	user, err := getReqUser(r)
	if err != nil {
		return nil, err
	}

	return &wsConn{
		conn,
		ctx,
		user,
		make(chan interface{}),
	}, nil
}

// HandleWebsocketPath is an http.HandleFunc which upgrades the request,
// setups the connection and starts handling websocket messages
func HandleWebsocketPath(w http.ResponseWriter, r *http.Request) {
	conn, err := upgradeSocket(w, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "websocket upgrade err: %v\n", err)
		return
	}

	ip, err := getReqIP(r)
	if err != nil {
		return
	}

	Connections[ip] = conn
	go func() {
		for {
			select {
			case m := <-conn.sendQueue:
				conn.WriteJSON(m)
			case <-conn.ctx.Done():
				return
			}
		}
	}()

	if conn.user != nil {
		// Note(netux): Needed so the max stacked on the client updates
		sendPixelsAvailable(conn, "auth")
		sendUserInfo(conn)
		if conn.user.PixelStacker.Stack > 0 {
			sendPixelsAvailable(conn, "connected")
		}
		conn.user.PixelStacker.StartTimer()
		if conn.user.PixelStacker.Stack == 0 {
			sendCooldown(conn, conn.user.PixelStacker.GetCooldown())
		}
	}

	go handleIncomingMessages(conn)
	if conn.user != nil {
		go handleUserEvents(conn)
	}
}

func handleUserEvents(conn *wsConn) {
	for {
		select {
		case isGain := <-conn.user.PixelStacker.C:
			cause := "consume"
			if isGain {
				cause = "stackGain"
			}
			sendPixelsAvailable(conn, cause)
		case <-conn.ctx.Done():
			return
		}
	}
}

func handleIncomingMessages(conn *wsConn) {
	defer func() {
		delete(Connections, conn.RemoteAddr().String())
		conn.Close()
	}()

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				return
			}
			fmt.Fprintf(os.Stderr, "conn reading err: %v\n", err)
			continue
		}
		var msgType wsMessageType
		{
			var wsMsg wsMessage
			if err := json.Unmarshal(rawMsg, &wsMsg); err != nil {
				fmt.Fprintf(os.Stderr, "websocket JSON parsing error: %v\n", err)
				return
			}
			msgType = wsMsg.Type
		}

		switch msgType {
		case wsPixelType:
			if conn.user == nil {
				break
			}

			var pixelMsg wsPixelReq
			if err := json.Unmarshal(rawMsg, &pixelMsg); err != nil {
				fmt.Fprintf(os.Stderr, "websocket JSON Pixel parsing error: %v\n", err)
				break
			}
			handlePixel(conn, pixelMsg)
		default:
			fmt.Fprintf(os.Stderr, "unhandled websocket msgType %s: %v\n", msgType, string(rawMsg))
		}
	}
}

const wsUserInfoType wsMessageType = "userinfo"

type wsUserInfo struct {
	wsMessage
	AuthMethod string   `json:"method"`
	Role       UserRole `json:"role"`
	Username   string   `json:"username"`

	CooldownOverride bool `json:"cdOverride"`

	IsBanned           bool   `json:"banned"`
	BanExpiry          int    `json:"banExpiry"`
	BanReason          string `json:"ban_reason"`
	IsChatBanned       bool   `json:"chatBanned"`
	ChatBanExpiry      int    `json:"chatBanExpiry"`
	ChatBanIsPermanent bool   `json:"chatBanIsPerma"`
}

// sendUserInfo sends an "userinfo" message through the websocket connection conn
func sendUserInfo(conn *wsConn) {
	conn.queue(wsUserInfo{
		wsMessage:  withType(wsUserInfoType),
		AuthMethod: conn.user.Login.Method,
		Role:       conn.user.Role,
		Username:   conn.user.Name,
	})
}

const wsAckType = "ACK"

type wsAck struct {
	wsMessage
	ackFor string
}

func ackFor(a string) wsAck {
	return wsAck{withType(wsAckType), a}
}

const wsPixelsAvailableType = "pixels"

type wsPixelsAvailable struct {
	wsMessage
	Count uint   `json:"count"`
	Cause string `json:"cause"`
}

func sendPixelsAvailable(conn *wsConn, cause string) {
	conn.queue(wsPixelsAvailable{
		withType(wsPixelsAvailableType),
		conn.user.PixelStacker.Stack,
		cause,
	})
}

const wsCooldownType = "cooldown"

type wsCooldown struct {
	wsMessage
	Wait float32 `json:"wait"`
}

func sendCooldown(conn *wsConn, wait time.Duration) {
	conn.queue(wsCooldown{
		withType(wsCooldownType),
		float32(wait) / float32(time.Second),
	})
}

const wsPixelType = "pixel"

type wsPixel struct {
	PosX     uint `json:"x"`
	PosY     uint `json:"y"`
	ColorIdx byte `json:"color"`
}

type wsPixelReq struct {
	wsMessage
	wsPixel
}

type wsPixelRes struct {
	wsMessage
	Pixels []wsPixel `json:"pixels"`
}

type wsAckForPixel struct {
	wsAck
	PosX uint `json:"x"`
	PosY uint `json:"y"`
}

func handlePixel(conn *wsConn, pixelMsg wsPixelReq) {
	if conn.user == nil {
		return
	}

	var ps = conn.user.PixelStacker
	if ps.Stack == 0 {
		return
	}

	if App.Canvas.GetPixelColorIndex(pixelMsg.PosX, pixelMsg.PosY) == pixelMsg.ColorIdx {
		return
	}

	ps.StopTimer()
	conn.queue(wsAckForPixel{
		ackFor("PLACE"),
		pixelMsg.PosX,
		pixelMsg.PosY,
	})

	ps.Consume()

	App.Canvas.SetPixelColor(pixelMsg.PosX, pixelMsg.PosY, pixelMsg.ColorIdx)
	if err := App.DB.PlacePixel(pixelMsg.PosX, pixelMsg.PosY, pixelMsg.ColorIdx, conn.user); err != nil {
		fmt.Fprintf(os.Stderr, "cannot place pixel at (%d, %d) by user with ID %d in database: %v", pixelMsg.PosX, pixelMsg.PosY, conn.user.ID, err)
	}
	ps.StartTimer()

	if conn.user.PixelStacker.Stack == 0 {
		cd := ps.GetCooldown()
		if err := App.DB.SetUserCooldownExpiry(conn.user.ID, time.Now().Add(cd)); err != nil {
			fmt.Fprintf(os.Stderr, "cannot set cooldown expiry for user with ID %d in database: %v", conn.user.ID, err)
		}
		sendCooldown(conn, cd)
	}

	pixelsMsg := wsPixelRes{
		wsMessage{
			Type: wsPixelType,
		},
		[]wsPixel{pixelMsg.wsPixel},
	}
	for _, conn := range Connections {
		conn.queue(pixelsMsg)
	}
}
