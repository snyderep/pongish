// +build js

package main

import (
	"fmt"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/console"
	"honnef.co/go/js/dom"
	"strings"
	"time"
)

type gateway struct {
	conn     *websocket.Conn
	send     chan string
	statusEl dom.HTMLElement
	canvas   *canvas
}

func newGateway() *gateway {
	win := dom.GetWindow()
	doc := win.Document()

	wsEndpoint := doc.GetElementByID("ws-endpoint").(dom.HTMLElement).TextContent()
	statusEl := doc.GetElementByID("status").(dom.HTMLElement)

	canvas := newCanvas(doc.GetElementByID("board").(*dom.HTMLCanvasElement))

	statusEl.SetTextContent("Connecting")
	conn := connect(wsEndpoint)
	statusEl.SetTextContent("Waiting To Play")

	wsSend := make(chan string)

	gw := &gateway{conn: conn, send: wsSend, statusEl: statusEl, canvas: canvas}

	// start send loop (send over websocket to server)
	go func(s chan string) {
		for {
			msg := <-gw.send

			_, err := conn.Write([]byte(msg))
			if err != nil {
				console.Error(err.Error())
			}
		}
	}(wsSend)

	// listen for canvas events
	go func() {
		for {
			e := <-canvas.event

			parts := strings.Split(e, ",")
			if parts[0] == "L" {
				gw.processLostEvent()
			} else if parts[0] == "N" {
				gw.processNetExchangeEvent(e)
			} else {
				console.Log(fmt.Sprintf("unsupported event: %s\n", e))
			}
		}
	}()

	return gw
}

func (g *gateway) start() {
	for {
		buf := make([]byte, 1024)
		n, err := g.conn.Read(buf) // Blocks until a WebSocket frame is received
		if err == nil {
			g.handleMessage(buf[:n])
		} else {
			console.Error(err.Error())
		}
	}
}

func (g *gateway) handleMessage(msg []byte) {
	m := string(msg)

	parts := strings.Split(m, ",")

	if parts[0] == "P" {
		g.handlePlayMessage(parts[1])
	} else if parts[0] == "B" {
		// 1 = y position
		// 2 = angle
		// 3 = speed
		v, err := newVectorFromStrings(parts[1], parts[2], parts[3])
		if err != nil {
			console.Log(err.Error())
		}

		g.handleBallInPlayMessage(v)
	} else {
		console.Log(fmt.Sprintf("unsupported message: %s\n", m))
	}
}

func (g *gateway) handlePlayMessage(side string) {
	dSide := strings.ToUpper(side)

	console.Log(fmt.Sprintf("handling play message - side: %s\n", dSide))

	g.statusEl.SetTextContent("Playing (" + dSide + ")")

	g.canvas.reset(dSide)
}

func (g *gateway) handleBallInPlayMessage(v *vector) {
	console.Log(fmt.Sprintf("handling ball in play message - y pos: %d, angle: %v, speed: %v\n", v.yPos, v.angle, v.speed))

	// Note angle is always specified as if for the LEFT side player, the RIGHT side player
	// will do the inverse.

	g.canvas.ballStart(v)
}

func (g *gateway) processLostEvent() {
	g.statusEl.SetTextContent("Lost - Waiting To Play")
	g.send <- "L"
}

func (g *gateway) processNetExchangeEvent(eventMsg string) {
	console.Log(fmt.Printf("processing net exchange event - %s\n", eventMsg))
	g.send <- eventMsg
}

func connect(wsEndpoint string) *websocket.Conn {
	count := 0

	ticker := time.NewTicker(time.Duration(1) * time.Second)
	defer ticker.Stop()

	// connect
	for {
		count++

		console.Log(fmt.Printf("trying to connect to server at %s, attempt %d", wsEndpoint, count))

		conn, err := websocket.Dial(wsEndpoint) // Blocks until connection is established
		if err == nil {
			return conn
		}
		console.Error(err.Error())

		<-ticker.C
	}
}
