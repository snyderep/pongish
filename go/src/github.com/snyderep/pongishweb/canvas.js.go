// +build js

package main

import (
	"fmt"
	"honnef.co/go/js/console"
	"honnef.co/go/js/dom"
	"math"
	"math/rand"
	"time"
)

const (
	animationFramesPerSecond int     = 60
	ballEndAngle             float64 = math.Pi * 2.0
	degreeToRadian           float64 = math.Pi / 180.0
	radianToDegree           float64 = 180 / math.Pi
)

type ball struct {
	xMovement float64
	yMovement float64
	xPos      int
	yPos      int
	radius    int
}

func (b *ball) draw(canvasEl *dom.HTMLCanvasElement) {
	b.xPos += int(math.Floor(b.xMovement + 0.5))
	b.yPos += int(math.Floor(b.yMovement + 0.5))

	ctx := canvasEl.GetContext2d()
	ctx.FillStyle = "red"
	ctx.BeginPath()
	ctx.Arc(b.xPos, b.yPos, b.radius, 0, 6, false)
	ctx.Fill()
	ctx.ClosePath()
}

type paddle struct {
	yMovement int
	xPos      int
	yPos      int
	height    int
	width     int
	hit       bool
}

func (p *paddle) draw(canvasEl *dom.HTMLCanvasElement) {
	newYPos := p.yPos + p.yMovement
	if newYPos > 5 && newYPos < (canvasEl.Height-p.height-5) {
		p.yPos = newYPos
	}

	ctx := canvasEl.GetContext2d()
	ctx.FillStyle = "#0000ff"
	ctx.FillRect(p.xPos, p.yPos, p.width, p.height)
}

type canvas struct {
	canvasEl *dom.HTMLCanvasElement
	bll      *ball
	pddl     *paddle
	side     string
	event    chan string
}

func newCanvas(canvasEl *dom.HTMLCanvasElement) *canvas {
	c := &canvas{canvasEl: canvasEl, event: make(chan string)}

	canvasEl.AddEventListener("keydown", false, func(event dom.Event) {
		c.handleKeyDown(event.(*dom.KeyboardEvent))
	})

	canvasEl.AddEventListener("keyup", false, func(event dom.Event) {
		c.handleKeyUp(event.(*dom.KeyboardEvent))
	})

	go func() {
		ticker := time.NewTicker(time.Duration(1000/animationFramesPerSecond) * time.Millisecond)
		for {
			<-ticker.C

			c.draw()

			if c.checkLost() {
				c.bll = nil
				c.event <- "L"
			} else {
				c.checkTopBottomCollision()
				c.checkPaddleCollision()
				if c.checkOverNet() {
					rad := math.Atan2(c.bll.yMovement, c.bll.xMovement)
					deg := rad * radianToDegree
					if deg < 0 {
						deg = 360 + deg
					}
					speed := math.Floor(c.bll.xMovement/math.Cos(rad) + 0.5)
					if speed < 2 {
						speed = 2
					}
					c.event <- fmt.Sprintf("N,%d,%d,%d", c.bll.yPos, int(deg), int(speed))
					c.bll = nil
				}
			}
		}
	}()

	return c
}

func (c *canvas) handleKeyDown(e *dom.KeyboardEvent) {
	if e.KeyIdentifier == "Up" {
		c.pddl.yMovement = -4
	} else if e.KeyIdentifier == "Down" {
		c.pddl.yMovement = 4
	}
}

func (c *canvas) handleKeyUp(e *dom.KeyboardEvent) {
	if e.KeyIdentifier == "Up" || e.KeyIdentifier == "Down" {
		c.pddl.yMovement = 0
	}
}

func (c *canvas) ballStart(v *vector) {
	var xPos int

	if c.side == "LEFT" {
		xPos = c.canvasEl.Width - 5
	} else {
		xPos = 5
	}

	radians := v.angle * degreeToRadian

	xMovement := math.Cos(radians) * v.speed
	yMovement := math.Sin(radians) * v.speed

	c.bll = &ball{xPos: xPos, yPos: v.yPos, radius: 20, xMovement: xMovement, yMovement: yMovement}
	c.pddl.hit = false
}

func (c *canvas) draw() {
	c.clear()

	if c.bll != nil {
		c.bll.draw(c.canvasEl)
	}
	if c.pddl != nil {
		c.pddl.draw(c.canvasEl)
	}
}

func (c *canvas) clear() {
	ctx := c.canvasEl.GetContext2d()
	ctx.ClearRect(0, 0, c.canvasEl.Width, c.canvasEl.Height)
}

func (c *canvas) checkLost() bool {
	// if the ball hits the end wall then the player has lost,
	lost := false

	if c.bll != nil {
		if c.side == "LEFT" {
			if c.bll.xPos <= 0 {
				lost = true
			}
		} else if c.bll.xPos >= c.canvasEl.Width {
			lost = true
		}
	}

	return lost
}

func (c *canvas) checkTopBottomCollision() {
	if c.bll == nil {
		return
	}

	if c.bll.yPos <= c.bll.radius || c.bll.yPos >= c.canvasEl.Height-c.bll.radius {
		c.bll.yMovement *= -1
	}
}

func (c *canvas) checkPaddleCollision() {
	if c.bll == nil {
		return
	}

	// If the ball is in the vicinity of where the paddle couldbe then do some more fancy collision detection.
	// First check to see if the ball already hit the paddle.
	if (!c.pddl.hit) && ((c.side == "LEFT" && c.bll.xPos < (c.bll.radius+c.pddl.xPos+c.pddl.width+10)) ||
		(c.side == "RIGHT" && c.bll.xPos > (c.pddl.xPos-c.bll.radius-10))) {

		detectionArea := int(float64(c.bll.radius) * float64(1.5))
		var m int
		if c.side == "LEFT" {
			m = -1
		} else {
			m = 1
		}
		whatColor := getImageData(c.canvasEl.GetContext2d(), c.bll.xPos+(c.bll.radius*m), c.bll.yPos+(c.bll.radius*m), detectionArea, detectionArea)
		if whatColor.anyBlue() {
			c.bll.xMovement *= -1
			c.bll.yMovement += float64(rand.Intn(3) - 1)
			c.pddl.hit = true
		}
	}
}

func (c *canvas) checkOverNet() bool {
	if c.bll == nil {
		return false
	}
	return (c.side == "LEFT" && (c.bll.xPos > c.canvasEl.Width)) || (c.side == "RIGHT" && (c.bll.xPos < 0))
}

func (c *canvas) reset(side string) {
	c.side = side

	var xPos int
	var offsetFromEnd = 10
	var paddleWidth = 20

	if side == "LEFT" {
		xPos = offsetFromEnd
	} else {
		xPos = c.canvasEl.Width - offsetFromEnd - paddleWidth
	}

	c.pddl = &paddle{xPos: xPos, yPos: 350, height: 150, width: paddleWidth}
	c.bll = nil
}
