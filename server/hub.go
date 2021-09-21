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
	registry map[registryEntry]map[*connection]bool

	// lastEvents holds the last event sent for each registryEntry
	lastEvents map[registryEntry]simulation.Event

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
		e simulation.Event
		c *connection
	)
	for {
		select {
		case c = <-h.readChan:
			logger.Debug("Reading request from client", "submodule", "hub", "data", c.getRequest())
			h.dispatchObject(c)
		default:
		}
		select {
		case c = <-h.registerChan:
			logger.Debug("Registering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.register(c)
		case c = <-h.unregisterChan:
			logger.Info("Unregistering connection", "submodule", "hub", "connection", c.RemoteAddr())
			h.unregister(c)
		default:
		}
		select {
		case e = <-sim.EventChan:
			logger.Debug("Received event from simulation", "submodule", "hub", "event", e.Name, "object", e.Object)
			if e.Name == simulation.ClockEvent {
				sim.ProcessTimeStep()
			}
			h.notifyClients(e)
		default:
		}
	}
}

// register the given connection to this hub
func (h *Hub) register(c *connection) {
	switch c.clientType {
	case Client:
		h.clientConnections[c] = true
	}
}

// addConnectionToRegistry adds this connection to the registry for eventName and id.
func (h *Hub) addConnectionToRegistry(conn *connection, eventName simulation.EventName, id string) {
	re := registryEntry{eventName: eventName, id: id}
	if _, ok := h.registry[re]; !ok {
		h.registry[re] = make(map[*connection]bool)
	}
	h.registry[re][conn] = true
}

// removeEntryFromRegistry removes this connection from the registry for eventName and id.
func (h *Hub) removeEntryFromRegistry(conn *connection, eventName simulation.EventName, id string) {
	re := registryEntry{eventName: eventName, id: id}
	delete(h.registry[re], conn)
}

// removeConnectionFromRegistry removes all entries of this connection in the registry.
func (h *Hub) removeConnectionFromRegistry(conn *connection) {
	for re, rv := range h.registry {
		if _, ok := rv[conn]; ok {
			delete(h.registry[re], conn)
		}
	}
}

// unregister unregisters the connection to this hub
func (h *Hub) unregister(c *connection) {
	switch c.clientType {
	case Client:
		if _, ok := h.clientConnections[c]; ok {
			delete(h.clientConnections, c)
		}
		h.removeConnectionFromRegistry(c)
	}
}

// notifyClients sends the given event to all registered clients.
func (h *Hub) notifyClients(e simulation.Event) {
	logger.Debug("Notifying clients", "submodule", "hub", "event", e)
	h.updateLastEvents(e)
	// Notify clients that subscribed to all objects
	for conn := range h.registry[registryEntry{eventName: e.Name, id: ""}] {
		conn.pushChan <- NewNotificationResponse(e)
	}
	if e.Object.ID() == "" {
		// Object has no ID. Don't send twice
		return
	}
	// Notify clients that subscribed to specific object IDs
	for conn := range h.registry[registryEntry{eventName: e.Name, id: e.Object.ID()}] {
		conn.pushChan <- NewNotificationResponse(e)
	}
}

// updateLastEvents updates the lastEvents map in a concurrently safe way
func (h *Hub) updateLastEvents(e simulation.Event) {
	h.lastEvents[registryEntry{eventName: e.Name, id: e.Object.ID()}] = e
}

// dispatchObject process a request.
func (h *Hub) dispatchObject(conn *connection) {
	req := conn.popRequest()
	obj, ok := h.objects[req.Object]
	if !ok {
		conn.pushChan <- NewErrorResponse(req.ID, fmt.Errorf("unknown object %s", req.Object))
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
	h.registry = make(map[registryEntry]map[*connection]bool)
	h.lastEvents = make(map[registryEntry]simulation.Event)
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
