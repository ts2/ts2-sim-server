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
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestLogin(t *testing.T) {
	// Wait for server to come up
	time.Sleep(100 * time.Millisecond)
	c := clientDial(t)
	defer func() {
		c.Close()
	}()

	// Try to send something that is not a login request
	badRequest := Request{"Dummy", "dummy", nil}
	if err := c.WriteJSON(badRequest); err != nil {
		t.Error(err)
	}
	var serverResponse ResponseStatus
	c.ReadJSON(&serverResponse)
	assertEqual(t, serverResponse, ResponseStatus{RESPONSE, DataStatus{FAIL, "Error: Register required"}}, "Register/Wrong request")
	_, _, err := c.ReadMessage()
	if _, ok := err.(*websocket.CloseError); err == nil || !ok {
		t.Errorf("Register/Wrong request/Connection should be closed")
	}
	c.Close()

	// Incorrect login
	c, err = register(t, CLIENT, "", "wrong-token")
	expectedErrorMsg := "Error: Invalid register parameters"
	if err == nil || err.Error() != expectedErrorMsg {
		t.Errorf("Register/Incorrect: Unexpected behaviour")
	}
	_, _, err = c.ReadMessage()
	if _, ok := err.(*websocket.CloseError); err == nil || !ok {
		t.Errorf("Register/Wrong request/Connection should be closed")
	}
	c.Close()

	// Correct login
	if _, err = register(t, CLIENT, "", "client-secret"); err != nil {
		t.Errorf(err.Error())
	}
}

func TestDoubleLogin(t *testing.T) {
	// Wait for server to come up
	time.Sleep(100 * time.Millisecond)
	c, err := register(t, CLIENT, "", "client-secret")
	defer func() {
		c.Close()
	}()

	if err != nil {
		t.Errorf(err.Error())
	} else {
		c.WriteJSON(RequestRegister{"server", "register", ParamsRegister{CLIENT, "", "client-secret"}})
		var expectedResponse ResponseStatus
		c.ReadJSON(&expectedResponse)
		if expectedResponse.Data.Status != FAIL {
			t.Errorf("Double login: should have failed")
		} else if expectedResponse.Data.Message != "Error: Can't call register when already registered" {
			t.Errorf("Double login: Wrong error message : %s", expectedResponse.Data.Message)
		}
	}
}
