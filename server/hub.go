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
	"fmt"

	"github.com/ts2/ts2-sim-server/simulation"
)

// The Hub makes the interface between the Simulation and the websocket clients
type Hub struct {
	// Registered client connections
	clientConnections map[*connection]bool

	// Registry of client listeners
	registry map[*registryEntry]bool

	// Register requests from the connection
	registerChan chan *connection

	// Unregister requests from connection
	unregisterChan chan *connection

	// Received requests channel
	readChan chan *connection

	objects map[string]hubObject
}

type hubObject interface {
	dispatch(h *Hub, req Request, c *connection)
}

// run is the loop for handling dispatching requests and responses
func (h *Hub) run(hubUp chan bool) {
	logger.Info("Hub starting...", "submodule", "hub")

	hubUp <- true
	var (
		e *simulation.Event
		c *connection
	)
	for {
		select {
		case e = <-sim.EventChan:
			logger.Debug("Received event from simulation", "submodule", "hub", "object", e)
			h.notifyClients(e)
		case c = <-h.readChan:
			logger.Debug("Reading request from client", "submodule", "hub", "object", c.LastRequest)
			go h.dispatchObject(c)
		case c = <-h.registerChan:
			logger.Debug("Registering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.register(c)
		case c = <-h.unregisterChan:
			logger.Debug("Unregistering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.unregister(c)
		}
	}
}

// register registers the given connection to this hub
func (h *Hub) register(c *connection) {
	switch c.clientType {
	case Client:
		h.clientConnections[c] = true
	}
}

// unregister unregisters the connection to this hub
func (h *Hub) unregister(c *connection) {
	switch c.clientType {
	case Client:
		if _, ok := h.clientConnections[c]; ok {
			delete(h.clientConnections, c)
		}
	}
}

// notifyClients sends the given event to all registered clients.
func (h *Hub) notifyClients(e *simulation.Event) {
	logger.Debug("Notifying clients", "submodule", "hub", "event", e)
	for re := range h.registry {
		if re.eventName == e.Name {
			re.conn.pushChan <- NewNotificationResponse(e)
		}
	}
}

// dispatchObject process a request.
//
// - req is the request to process
// - ch is the channel on which to send the response
func (h *Hub) dispatchObject(conn *connection) {
	req := conn.LastRequest
	obj, ok := h.objects[req.Object]
	if !ok {
		conn.pushChan <- NewErrorResponse(req.ID, fmt.Errorf("unknwon object %s", req.Object))
		logger.Debug("Request for unknown object received", "submodule", "hub", "object", req.Object)
		return
	}
	obj.dispatch(h, req, conn)
}

// newHub returns a pointer to a new Hub instance
func newHub() *Hub {
	h := new(Hub)
	// make connection maps
	h.clientConnections = make(map[*connection]bool)
	// make registry map
	h.registry = make(map[*registryEntry]bool)
	// make channels
	h.registerChan = make(chan *connection)
	h.unregisterChan = make(chan *connection)
	h.readChan = make(chan *connection)
	h.objects = make(map[string]hubObject)
	return h
}

func init() {
	hub = newHub()
}
