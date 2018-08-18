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
	"encoding/json"
	"testing"
	"time"

	"github.com/ts2/ts2-sim-server/simulation"
)

func TestStartPauseSimulation(t *testing.T) {
	// Wait for server to come up
	time.Sleep(100 * time.Millisecond)
	c, err := register(t, CLIENT, "", "client-secret")
	defer func() {
		c.Close()
	}()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	c.WriteJSON(Request{Object: "simulation", Action: "start"})
	var expectedResponse ResponseStatus
	c.ReadJSON(&expectedResponse)
	if expectedResponse.Data.Status != OK {
		t.Errorf("The response from server is NOOK (Simulation/start)")
	}

	c.WriteJSON(Request{Object: "simulation", Action: "pause"})
	c.ReadJSON(&expectedResponse)
	if expectedResponse.Data.Status != OK {
		t.Errorf("The response from server is NOOK (Simulation/pause)")
	}

}

func TestAddRemoveListeners(t *testing.T) {
	// Wait for server to come up
	time.Sleep(100 * time.Millisecond)
	c, err := register(t, CLIENT, "", "client-secret")
	defer func() {
		c.Close()
	}()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	// add listener for clock
	c.WriteJSON(RequestListener{
		Object: "server",
		Action: "addListener",
		Params: ParamsListener{
			Event: simulation.CLOCK,
		},
	})
	var expectedResponse ResponseStatus
	c.ReadJSON(&expectedResponse)
	if expectedResponse.Data.Status != OK {
		t.Errorf("The response from server is NOOK (Server/addListener)")
	}

	// start simulation
	c.WriteJSON(Request{Object: "simulation", Action: "start"})
	c.ReadJSON(&expectedResponse)
	if expectedResponse.Data.Status != OK {
		t.Errorf("The response from server is NOOK (Simulation/start)")
	}

	// check we receive events
	var clockEvent ResponseEvent
	c.ReadJSON(&clockEvent)
	if clockEvent.MsgType != EVENT || clockEvent.Data.Name != simulation.CLOCK {
		t.Errorf("No clock event received from server !")
	}

	time.Sleep(1 * time.Second)
	// remove listener
	c.WriteJSON(Request{Object: "Server", Action: "removeListener", Params: json.RawMessage("{\"event\": \"clock\"}")})
	c.ReadJSON(&expectedResponse)
	if expectedResponse.Data.Status != OK {
		t.Errorf("The response from server is NOOK (Server/removeListener)")
	}
}
