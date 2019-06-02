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
	"testing"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConnection(t *testing.T) {
	// Wait for server to come up
	time.Sleep(2 * time.Second)
	Convey("Testing server connection", t, func() {
		c := clientDial(t)
		Convey("Login test", func() {
			Convey("First request that is not a register request should fail", func() {
				badRequest := Request{1234, "Dummy", "dummy", nil}
				err := c.WriteJSON(badRequest)
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp, ShouldResemble, ResponseStatus{1234, TypeResponse, DataStatus{Fail, "Error: register required"}})
				_, _, err = c.ReadMessage()
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, new(websocket.CloseError))
			})
			Convey("Incorrect login should fail", func() {
				err := register(t, c, Client, "", "wrong-token")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Error: invalid register parameters")
				_, _, err = c.ReadMessage()
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, new(websocket.CloseError))
			})
			Convey("Correct login should be allowed", func() {
				err := register(t, c, Client, "", "client-secret")
				So(err, ShouldBeNil)
			})
		})
		Convey("Login double test", func() {
			err := register(t, c, Client, "", "client-secret")
			So(err, ShouldBeNil)
			err = c.WriteJSON(RequestRegister{1234, "server", "register", ParamsRegister{Client, "", "client-secret"}})
			So(err, ShouldBeNil)
			var resp ResponseStatus
			err = c.ReadJSON(&resp)
			So(err, ShouldBeNil)
			So(resp.Data.Status, ShouldEqual, Fail)
			So(resp.Data.Message, ShouldEqual, "Error: can't call register when already registered")
		})
		Reset(func() {
			err := c.Close()
			So(err, ShouldBeNil)
		})
	})
}
