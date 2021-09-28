// Copyright (C) 2008-2018 by Nicolas Piganeau and the TS2 TEAM
// (See AUTHORS file)
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the
// Free Software Foundation, Inc.,
// 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/ts2/ts2-sim-server/simulation"
	log "gopkg.in/inconshreveable/log15.v2"
)

func TestMain(m *testing.M) {
	mainLogger := log.New()
	if os.Getenv("TS2_DEBUG") == "" {
		mainLogger.SetHandler(log.DiscardHandler())
	}
	InitializeLogger(mainLogger)
	simulation.InitializeLogger(mainLogger)
	data, _ := ioutil.ReadFile("../simulation/testdata/demo.json")
	var s simulation.Simulation
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Println("Unable to load demo.json:", err)
		os.Exit(1)
	}
	go Run(&s, "0.0.0.0", "22222")
	s.Initialize()
	os.Exit(m.Run())
}

func clientDial(t *testing.T) *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:22222", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Error(err)
	}
	return conn
}

// register dials to the server and logs the client in
func register(_ *testing.T, c *websocket.Conn, ct ClientType, mt ManagerType, token string) error {
	loginRequest := RequestRegister{1234, "server", "register", ParamsRegister{ct, mt, token}}
	if err := c.WriteJSON(loginRequest); err != nil {
		return err
	}
	var expectedResponse ResponseStatus
	c.ReadJSON(&expectedResponse)
	if expectedResponse.Data.Status == Ok {
		return nil
	} else {
		return fmt.Errorf(expectedResponse.Data.Message)
	}
}

// addListener for the given event
func addListener(t *testing.T, c *websocket.Conn, event simulation.EventName) {
	err := c.WriteJSON(RequestListener{
		Object: "server",
		Action: "addListener",
		Params: ParamsListener{
			Event: event,
		},
	})
	if err != nil {
		t.Error(err)
	}
	var resp ResponseStatus
	for resp.MsgType != TypeResponse {
		err = c.ReadJSON(&resp)
		if err != nil {
			t.Error(err)
		}
	}
	if resp.Data.Status != Ok {
		t.Errorf("error while setting up listener: %v", resp)
	}
}