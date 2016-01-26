package server

import (
	"container/list"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type stateT uint8

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// player states
const (
	waiting stateT = iota
	playing
	dead
	lost
)

type sideT string

const (
	left  sideT = "LEFT"
	right sideT = "RIGHT"
)

const (
	boardStatePeriod = time.Duration(500) * time.Millisecond
	pingPeriod       = time.Duration(2) * time.Second
	pongWait         = time.Duration(3) * time.Second
	writeWait        = time.Duration(2) * time.Second
)

type courtT struct {
	waiters     *waitListT // waiting to play
	leftPlayer  *player
	rightPlayer *player
}

func newCourt() *courtT {
	waitList := newWaitListT(64)
	court := &courtT{waiters: waitList}

	go func() {
		for {
			<-time.NewTimer(boardStatePeriod).C

			// move any losers to the waiting list
			if court.sendLoserToWaitList(court.leftPlayer) {
				court.leftPlayer = nil
			}
			if court.sendLoserToWaitList(court.rightPlayer) {
				court.rightPlayer = nil
			}

			if err := court.doNetExchange(); err != nil {
				panic(err)
			}

			// ensure we have 2 players
			court.ensurePlayers()

			// ensure there's a ball on the court
			court.ensureBall()
		}
	}()

	return court
}

func (c *courtT) doNetExchange() error {
	if c.leftPlayer != nil && c.rightPlayer != nil {
		var neMsg string
		if c.leftPlayer.ball && c.leftPlayer.netExchange != "" {
			neMsg = c.leftPlayer.netExchange
		} else if c.rightPlayer.ball && c.rightPlayer.netExchange != "" {
			neMsg = c.rightPlayer.netExchange
		}

		if neMsg != "" {
			parts := strings.Split(neMsg, ",")

			yPos, err := strconv.ParseInt(parts[1], 0, 32)
			if err != nil {
				return err
			}
			angle, err := strconv.ParseInt(parts[2], 0, 32)
			if err != nil {
				return err
			}
			speed, err := strconv.ParseInt(parts[3], 0, 32)
			if err != nil {
				return err
			}

			if c.leftPlayer.ball {
				c.rightPlayer.sendBallInMsg(int(yPos), int(angle), int(speed))
				c.leftPlayer.ball = false
				c.leftPlayer.netExchange = ""
			} else if c.rightPlayer.ball {
				c.leftPlayer.sendBallInMsg(int(yPos), int(angle), int(speed))
				c.rightPlayer.ball = false
				c.rightPlayer.netExchange = ""
			}
		}
	}
	return nil
}

func (c *courtT) ensurePlayers() {
	// TODO: dry this up, this is getting ugly
	if c.leftPlayer == nil || c.leftPlayer.state == dead {
		if c.leftPlayer != nil && c.leftPlayer.state == dead {
			c.leftPlayer.wsConn.Close()
		}
		c.leftPlayer = c.waiters.Take()
		if c.leftPlayer != nil {
			log.Printf("taking left player from wait list. addr: %s\n", c.leftPlayer.addr())
			c.leftPlayer.play(left)
		}
	}
	if c.rightPlayer == nil || c.rightPlayer.state == dead {
		if c.rightPlayer != nil && c.rightPlayer.state == dead {
			c.rightPlayer.wsConn.Close()
		}
		c.rightPlayer = c.waiters.Take()
		if c.rightPlayer != nil {
			log.Printf("taking right player from wait list. addr: %s\n", c.leftPlayer.addr())
			c.rightPlayer.play(right)
		}
	}
}

func (c *courtT) ensureBall() {
	if c.leftPlayer != nil && c.rightPlayer != nil && !c.leftPlayer.ball && !c.rightPlayer.ball {
		yPos := rnd.Intn(800) + 100
		angle := rnd.Intn(90) + 135
		speed := rnd.Intn(4) + 2
		// we'll always serve to the left player for now
		log.Printf("serving ball to left player. add: %s\n", c.leftPlayer.addr())
		c.leftPlayer.sendBallInMsg(yPos, angle, speed)
	}
}

func (c *courtT) sendLoserToWaitList(p *player) bool {
	if p != nil && p.state == lost {
		p.state = waiting
		if err := c.waiters.Add(p); err != nil {
			log.Println(err)
		}
		return true
	}

	return false
}

var court = newCourt()

type waitListT struct {
	lst     *list.List
	lock    sync.RWMutex
	maxSize int
}

func newWaitListT(maxSize int) *waitListT {
	pl := &waitListT{maxSize: maxSize, lst: list.New()}

	go func() {
		pruneTicker := time.NewTicker(time.Second * 1)

		for {
			<-pruneTicker.C
			pl.pruneDead()
		}
	}()

	return pl
}

func (pl *waitListT) Add(w *player) error {
	pl.lock.Lock()
	defer pl.lock.Unlock()

	if pl.lst.Len() >= pl.maxSize {
		return ErrTooManyWaiting
	}
	if w.state != waiting {
		return ErrMustBeWaitingState
	}

	pl.lst.PushBack(w)

	return nil
}

func (pl *waitListT) Take() *player {
	pl.lock.Lock()
	defer pl.lock.Unlock()

	element := pl.lst.Front()

	if element == nil {
		return nil
	}

	player := pl.lst.Remove(element).(*player)

	return player
}

func (pl *waitListT) pruneDead() {
	log.Printf("Pruning dead players. Waiting player count: %d\n", pl.lst.Len())

	pl.lock.Lock()
	defer pl.lock.Unlock()

	for e := pl.lst.Front(); e != nil; e = e.Next() {
		p := e.Value.(*player)
		//log.Println(p)
		if p.state == dead {
			if err := p.wsConn.Close(); err != nil {
				log.Printf("error closing websocket connection for %s: %s\n", p.addr(), err)
			}
			pl.lst.Remove(e)
		}
	}
}

type player struct {
	ball        bool
	netExchange string
	state       stateT
	start       time.Time
	wsConn      *websocket.Conn
	send        chan string
}

func addPlayer(wsConn *websocket.Conn) error {
	now := time.Now()

	p := &player{state: waiting, start: now, wsConn: wsConn, send: make(chan string, 8)}

	// start reading from the websocket connection
	go p.readPump()
	// start writing to the websocket connection
	go p.writePump()

	if err := court.waiters.Add(p); err != nil {
		return err
	}

	return nil
}

func (p *player) play(side sideT) {
	// tell the client that it's playing
	p.sendPlayMsg(side)
	p.state = playing
}

func (p *player) playing() bool {
	return p.state == playing
}

func (p *player) notPlaying() bool {
	return !p.playing()
}

func (p *player) addr() string {
	return p.wsConn.RemoteAddr().String()
}

func (p *player) sendPlayMsg(side sideT) {
	p.send <- fmt.Sprintf("P,%s", string(side))
}

func (p *player) sendBallInMsg(pos int, angle int, speed int) {
	p.ball = true
	p.send <- fmt.Sprintf("B,%d,%d,%d", pos, angle, speed)
}

func (p *player) readPump() {
	defer func() {
		p.state = dead
	}()

	p.wsConn.SetReadLimit(1024)
	p.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	p.wsConn.SetPongHandler(func(string) error {
		p.wsConn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msgType, msg, err := p.wsConn.ReadMessage()
		remoteAddr := p.wsConn.RemoteAddr().String()

		if err != nil {
			log.Printf("ws read from %v, err: %v\n", remoteAddr, err)
			break
		}

		log.Printf("ws read from %v, msg type: %v, msg: %v\n", remoteAddr, msgType, msg)

		msgS := string(msg)
		parts := strings.Split(msgS, ",")

		if parts[0] == "L" {
			p.handleLostMsg()
		} else if parts[0] == "N" {
			log.Printf("net exchange msg: %s\n", msgS)
			p.netExchange = msgS
		} else {
			log.Printf("unsupported message: %v\n", msg)
		}
	}
}

func (p *player) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		p.state = dead
	}()

	for {
		select {
		case message, ok := <-p.send:
			if ok {
				if err := p.write(websocket.TextMessage, []byte(message)); err != nil {
					log.Printf("websocket write error: %s\n", err)
					return
				}
			} else {
				p.write(websocket.CloseMessage, []byte{})
				return
			}
		case <-ticker.C:
			if err := p.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (p *player) write(messageType int, message []byte) error {
	p.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return p.wsConn.WriteMessage(messageType, message)
}

func (p *player) handleLostMsg() {
	p.state = lost
	p.ball = false
}

func (p *player) String() string {
	return fmt.Sprintf("%v, %v, %v, %v", p.addr(), p.ball, p.state, p.start)
}
