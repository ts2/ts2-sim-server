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
		logger.Warn("Request for second register received", "submodule", "hub", "object", req.Object, "action", req.Action)
	case "addListener":
		logger.Debug("Request for addListener received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", req.Params)
		if err := h.addRegistryEntry(req, conn); err != nil {
			ch <- NewErrorResponse(req.ID, err)
			return
		}
		ch <- NewOkResponse(req.ID, "Listener added successfully")
	case "removeListener":
		logger.Debug("Request for removeListener received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", req.Params)
		if err := h.removeRegistryEntry(req, conn); err != nil {
			ch <- NewErrorResponse(req.ID, err)
			return
		}
		ch <- NewOkResponse(req.ID, "Listener removed successfully")
	case "renotify":
		logger.Debug("Request for renotify received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", req.Params)
		if err := h.renotifyClient(req, conn); err != nil {
			ch <- NewErrorResponse(req.ID, err)
			return
		}
		ch <- NewOkResponse(req.ID, "Renotify request taken into account")
	default:
		ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", req.Params)
	}
}

// addRegistryEntry adds the given event registry entry to the registry.
func (h *Hub) addRegistryEntry(req Request, conn *connection) error {
	var pl ParamsListener
	if err := json.Unmarshal(req.Params, &pl); err != nil {
		logger.Error("Unparsable request (addRegistryEntry)", "submodule", "hub", "error", err, "request", req)
		return fmt.Errorf("unparsable request: %s (%s)", err, req.Params)
	}
	if len(pl.IDs) == 0 {
		h.addConnectionToRegistry(conn, pl.Event, "")
		logger.Debug("Registry entry added", "submodule", "hub", "eventName", pl.Event)
		return nil
	}
	for _, id := range pl.IDs {
		h.addConnectionToRegistry(conn, pl.Event, id)
	}
	logger.Debug("Registry entries added", "submodule", "hub", "eventName", pl.Event, "ids", pl.IDs)
	return nil
}

// removeRegistryEntry removes the given event registry entry from the registry.
func (h *Hub) removeRegistryEntry(req Request, conn *connection) error {
	var pl ParamsListener
	if err := json.Unmarshal(req.Params, &pl); err != nil {
		logger.Error("Unparsable request (addRegistryEntry)", "submodule", "hub", "error", err, "request", req)
		return fmt.Errorf("unparsable request: %s (%s)", err, req.Params)
	}
	if len(pl.IDs) == 0 {
		h.removeEntryFromRegistry(conn, pl.Event, "")
		logger.Debug("Registry entry deleted", "submodule", "hub", "eventName", pl.Event)
		return nil
	}
	for _, id := range pl.IDs {
		h.removeEntryFromRegistry(conn, pl.Event, id)
	}
	logger.Debug("Registry entries added", "submodule", "hub", "eventName", pl.Event, "ids", pl.IDs)
	return nil
}

// renotifyClient will resend the last notification for each event and object ID
func (h *Hub) renotifyClient(req Request, conn *connection) error {
	h.lastEventsMutex.RLock()
	defer h.lastEventsMutex.RUnlock()
	for re, event := range h.lastEvents {
		if _, ok := h.registry[registryEntry{eventName: re.eventName, id: ""}]; ok {
			if h.registry[registryEntry{eventName: event.Name, id: ""}][conn] {
				conn.pushChan <- NewNotificationResponse(event)
			}
		}
		if event.Object.ID() == "" {
			// Object has no ID. Don't send twice
			continue
		}
		if _, ok := h.registry[re]; ok {
			if h.registry[registryEntry{eventName: event.Name, id: event.Object.ID()}][conn] {
				conn.pushChan <- NewNotificationResponse(event)
			}
		}
	}
	return nil
}

var _ hubObject = new(serverObject)

func init() {
	hub.objects["server"] = new(serverObject)
}
