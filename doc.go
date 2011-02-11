/*
	The socketio package is a simple abstraction layer for different web browser-
	supported transport mechanisms. It is meant to be fully compatible with the
	Socket.IO client side JavaScript socket API library by LearnBoost Labs
	(http://socket.io/), but through custom formatters it might fit other client
	implementations too.

	It (together with the LearnBoost's client-side libraries) provides an easy way for
	developers to access the most popular browser transport mechanism today:
	multipart- and long-polling XMLHttpRequests, HTML5 WebSockets and
	forever-frames [TODO]. The socketio package works hand-in-hand with the standard
	http package by plugging itself into a configurable ServeMux. It has an callback-style
	API for handling connection events. The callbacks are:

		- SocketIO.OnConnect
		- SocketIO.OnDisconnect
		- SocketIO.OnMessage

	Other utility-methods include:

		- SocketIO.Mux
		- SocketIO.Broadcast
		- SocketIO.BroadcastExcept
		- SocketIO.GetConn
		- Conn.Send

	Each new connection will be automatically assigned an unique session id and
	using those the clients can reconnect without losing messages: the server
	persists clients' pending messages (until some configurable point) if they can't
	be immediately delivered. All writes through `Conn.Send` by design asynchronous.

	Finally, the actual format on the wire is described by a separate `Codec`.
	The default codec is compatible with the LearnBoost's Socket.IO client.

	For example, here is a simple chat server:

		package main

		import (
			"http"
			"log"
			"socketio"
		)

		func main() {
			sio := socketio.NewSocketIO(nil, nil)
			sio.Mux("/socket.io/", nil)

			http.Handle("/", http.FileServer("www/", "/"))

			sio.OnConnect(func(c *socketio.Conn) {
				sio.Broadcast(struct{ announcement string }{"connected: " + c.String()})
			})

			sio.OnDisconnect(func(c *socketio.Conn) {
				sio.BroadcastExcept(c, struct{ announcement string }{"disconnected: " + c.String()})
			})

			sio.OnMessage(func(c *socketio.Conn, msg string) {
				sio.BroadcastExcept(c,
					struct{ message []string }{[]string{c.String(), msg}})
			})

			log.Println("Server started.")
			if err := http.ListenAndServe(":8080", nil); err != nil {
				log.Exitln("ListenAndServer:", err)
			}
		}
*/
package socketio
