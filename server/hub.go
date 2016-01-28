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
)

/*
Hub makes the interface between the Simulation and the websocket clients
*/
type Hub struct {
	// Registered client connections
	clientConnections map[*connection]bool

	// Registered train manager connections
	managerConnections map[*connection]bool

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
	// make channels
	h.registerChan = make(chan *connection)
	h.unregisterChan = make(chan *connection)
	h.readChan = make(chan *connection)

	for {
		select {
		case e := <-sim.EventChan:
			logger.Debug("Received event from simulation", "submodule", "hub", "object", e)
		case c := <-h.readChan:
			logger.Debug("Reading request from client", "submodule", "hub", "object", c.LastRequest)
			go h.dispatchObject(c.LastRequest, c.pushChan)
		case c := <-h.registerChan:
			logger.Debug("Registering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.register(c)
		case c := <-h.unregisterChan:
			logger.Debug("Unregistering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.unregister(c)
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
dispatchObject process a request.

- req is the request to process
- ch is the channel on which to send the response
*/
func (h *Hub) dispatchObject(req Request, ch chan interface{}) {
	switch req.Object {
	case "Server":
		h.dispatchServer(req, ch)
	case "Simulation":
		h.dispatchSimulation(req, ch)
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
		ch <- NewErrorResponse(fmt.Errorf("Unknwon object %s", req.Object))
		logger.Debug("Request for unknown object received", "submodule", "hub", "object", req.Object)
	}
}

/*
dispatchServer processes requests made on the Server object
*/
func (h *Hub) dispatchServer(req Request, ch chan interface{}) {
	switch req.Action {
	case "login":
		ch <- NewErrorResponse(fmt.Errorf("Can't call login when already logged in"))
		logger.Debug("Request for second login received", "submodule", "hub", "object", req.Object, "action", req.Action)
	default:
		ch <- NewErrorResponse(fmt.Errorf("Unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

/*
dispatchSimulation processes requests made on the Simulation object
*/
func (h *Hub) dispatchSimulation(req Request, ch chan interface{}) {
	switch req.Action {
	case "start":
		logger.Debug("Request for simulation start received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sim.Start()
		ch <- NewOkResponse()
	case "pause":
		logger.Debug("Request for simulation pause received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sim.Pause()
		ch <- NewOkResponse()
	default:
		ch <- NewErrorResponse(fmt.Errorf("Unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}
