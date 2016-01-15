/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package server

import (
	"fmt"
	"log"
	"net/http"
	"errors"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/*
H_Websocket() handles  the WebSocket `/ws` endpoint of the server.

  - reads JSON from the client and sends a `Request` object to the hub.
  - receives `Response` objects from the hub and send JSON to the client.
*/
func H_Websocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn := &connection{
		Conn:     *ws,
		pushChan: make(chan interface{}, 256),
	}
	defer func() {
		conn.Close()
	}()
	// reply back with a simple message
	payload := NewErrorResponse(errors.New(fmt.Sprintf("%s - Login required", conn.RemoteAddr() )))
	conn.Conn.WriteJSON(payload)
	conn.loop()
}
