// +build js

package main

func main() {
	// start the gateway, start listening on the websocket and handling events
	g := newGateway()
	go g.start()
}
