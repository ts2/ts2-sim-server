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
)

type serverObject struct{}

// dispatch processes requests made on the Server object
func (s *serverObject) dispatch(h *Hub, req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {
	case "register":
		ch <- NewErrorResponse(req.ID, fmt.Errorf("can't call register when already registered"))
		logger.Debug("Request for second register received", "submodule", "hub", "object", req.Object, "action", req.Action)
	case "addListener":
		logger.Debug("Request for addListener received", "submodule", "hub", "object", req.Object, "action", req.Action)
		h.addRegistryEntry(req, conn)
		ch <- NewOkResponse(req.ID, "Listener added successfully")
	case "removeListener":
		logger.Debug("Request for removeListener received", "submodule", "hub", "object", req.Object, "action", req.Action)
		h.removeRegistryEntry(req, conn)
		ch <- NewOkResponse(req.ID, "Listener removed successfully")
	default:
		ch <- NewErrorResponse(req.ID, fmt.Errorf("unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

// addRegistryEntry adds the given event registry entry to the registry.
func (h *Hub) addRegistryEntry(req Request, conn *connection) {
	var pl ParamsListener
	if err := json.Unmarshal(req.Params, &pl); err != nil {
		logger.Error("Unparsable request (addRegistryEntry)", "submodule", "hub", "request", req)
	}
	re := registryEntry{conn: conn, eventName: pl.Event, ids: pl.Ids}
	h.registry[&re] = true
	logger.Debug("Registry entry added", "submodule", "hub", "entry", re)
}

// removeRegistryEntry removes the given event registry entry from the registry.
func (h *Hub) removeRegistryEntry(req Request, conn *connection) {
	var pl ParamsListener
	if err := json.Unmarshal(req.Params, &pl); err != nil {
		logger.Error("Unparsable request (addRegistryEntry)", "submodule", "hub", "request", req)
	}
	re := registryEntry{conn: conn, eventName: pl.Event}
	for r := range h.registry {
		if r.conn == re.conn && r.eventName == re.eventName {
			delete(h.registry, r)
			break
		}
	}
}

var _ hubObject = new(serverObject)

func init() {
	hub.objects["server"] = new(serverObject)
}
