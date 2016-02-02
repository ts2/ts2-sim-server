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
	"fmt"

	"github.com/ts2/ts2-sim-server/simulation"
)

/*
Hub makes the interface between the Simulation and the websocket clients
*/
type Hub struct {
	// Registered client connections
	clientConnections map[*connection]bool

	// Registered train manager connections
	managerConnections map[*connection]bool

	// Registry of client listeners
	registry map[*registryEntry]bool

	// Register requests from the connection
	registerChan chan *connection

	// Unregister requests from connection
	unregisterChan chan *connection

	// Received requests channel
	readChan chan *connection
}

/*
Hub.run() is the loop for handling dispatching requests and responses
*/
func (h *Hub) run() {
	logger.Info("Hub starting...", "submodule", "hub")
	// make connection maps
	h.clientConnections = make(map[*connection]bool)
	h.managerConnections = make(map[*connection]bool)
	// make registry map
	h.registry = make(map[*registryEntry]bool)
	// make channels
	h.registerChan = make(chan *connection)
	h.unregisterChan = make(chan *connection)
	h.readChan = make(chan *connection)

	for {
		select {
		case e := <-sim.EventChan:
			logger.Debug("Received event from simulation", "submodule", "hub", "object", e)
			h.notifyClients(e)
		case c := <-h.readChan:
			logger.Debug("Reading request from client", "submodule", "hub", "object", c.LastRequest)
			go h.dispatchObject(c)
		case c := <-h.registerChan:
			logger.Debug("Registering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.register(c)
		case c := <-h.unregisterChan:
			logger.Debug("Unregistering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.unregister(c)
		default:
		}
	}
}

/*
Hub.register() registers the connection to this hub
*/
func (h *Hub) register(c *connection) {
	switch c.clientType {
	case CLIENT:
		h.clientConnections[c] = true
	case MANAGER:
		h.managerConnections[c] = true
	}
}

/*
Hub.unregister() unregisters the connection to this hub
*/
func (h *Hub) unregister(c *connection) {
	switch c.clientType {
	case CLIENT:
		if _, ok := h.clientConnections[c]; ok {
			delete(h.clientConnections, c)
		}
	case MANAGER:
		if _, ok := h.managerConnections[c]; ok {
			delete(h.managerConnections, c)
		}
	}
}

/*
Hub.notifyClients sends the event received on the hub to all registered clients.
*/
func (h *Hub) notifyClients(e *simulation.Event) {
	logger.Debug("Notifying clients", "submodule", "hub", "event", e)
	for re := range h.registry {
		if re.eventName == e.Name {
			re.conn.pushChan <- NewEventResponse(e)
		}
	}
}

/*
dispatchObject process a request.

- req is the request to process
- ch is the channel on which to send the response
*/
func (h *Hub) dispatchObject(conn *connection) {
	req := conn.LastRequest
	switch req.Object {
	case "server":
		h.dispatchServer(req, conn)
	case "simulation":
		h.dispatchSimulation(req, conn)
		//	case "TrackItem":
		//		h.dispatchTrackItem(req, ch)
		//	case "Route":
		//		h.dispatchRoute(req, ch)
		//	case "TrainType":
		//		h.dispatchTrainType(req, ch)
		//	case "Service":
		//		h.dispatchService(req, ch)
		//	case "Train":
		//		h.dispatchTrain(req, ch)
	default:
		conn.pushChan <- NewErrorResponse(fmt.Errorf("Unknwon object %s", req.Object))
		logger.Debug("Request for unknown object received", "submodule", "hub", "object", req.Object)
	}
}

/*
dispatchServer processes requests made on the Server object
*/
func (h *Hub) dispatchServer(req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {
	case "register":
		ch <- NewErrorResponse(fmt.Errorf("Can't call register when already registered"))
		logger.Debug("Request for second register received", "submodule", "hub", "object", req.Object, "action", req.Action)
	case "addListener":
		logger.Debug("Request for addListener received", "submodule", "hub", "object", req.Object, "action", req.Action)
		h.addRegistryEntry(req, conn)
		ch <- NewOkResponse("Listener added successfully")
	case "removeListener":
		logger.Debug("Request for removeListener received", "submodule", "hub", "object", req.Object, "action", req.Action)
		h.removeRegistryEntry(req, conn)
		ch <- NewOkResponse("Listener removed successfully")
	default:
		ch <- NewErrorResponse(fmt.Errorf("Unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

/*
dispatchSimulation processes requests made on the Simulation object
*/
func (h *Hub) dispatchSimulation(req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {
	case "start":
		logger.Debug("Request for simulation start received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sim.Start()
		ch <- NewOkResponse("Simulation started successfully")
	case "pause":
		logger.Debug("Request for simulation pause received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sim.Pause()
		ch <- NewOkResponse("Simulation paused successfully")
	default:
		ch <- NewErrorResponse(fmt.Errorf("Unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

/*
Hub.addRegistryEntry adds the given registry entry to the registry.
*/
func (h *Hub) addRegistryEntry(req Request, conn *connection) {
	var pl ParamsListener
	if err := json.Unmarshal(req.Params, &pl); err != nil {
		logger.Error("Unparsable request (addRegistryEntry)", "submodule", "hub", "request", req)
	}
	re := registryEntry{conn: conn, eventName: pl.Event, ids: pl.Ids}
	h.registry[&re] = true
	logger.Debug("Registry entry added", "submodule", "hub", "entry", re)
}

/*
Hub.removeRegistryEntry removes the given registry entry from the registry.
*/
func (h *Hub) removeRegistryEntry(req Request, conn *connection) {
	var pl ParamsListener
	if err := json.Unmarshal(req.Params, &pl); err != nil {
		logger.Error("Unparsable request (addRegistryEntry)", "submodule", "hub", "request", req)
	}
	re := registryEntry{conn: conn, eventName: pl.Event}
	for r, _ := range h.registry {
		if r.conn == re.conn && r.eventName == re.eventName {
			delete(h.registry, r)
			break
		}
	}
}
