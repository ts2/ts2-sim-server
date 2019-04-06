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

	"github.com/ts2/ts2-sim-server/simulation"
)

type routeObject struct{}

// dispatch processes requests made on the route object
func (r *routeObject) dispatch(h *Hub, req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {
	case "list":
		logger.Debug("Request for route list received", "submodule", "hub", "object", req.Object, "action", req.Action)
		rtes, err := json.Marshal(sim.Routes)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req.ID, rtes)
	case "show":
		var idsParams = struct {
			IDs []string `json:"ids"`
		}{}
		err := json.Unmarshal(req.Params, &idsParams)
		logger.Debug("Request for route show received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", idsParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		rtes := make(map[string]*simulation.Route)
		for _, id := range idsParams.IDs {
			rtes[id] = sim.Routes[id]
		}
		rte, err := json.Marshal(rtes)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req.ID, rte)
	case "activate":
		var actParams = struct {
			ID         string `json:"id"`
			Persistent bool   `json:"persistent"`
		}{}
		err := json.Unmarshal(req.Params, &actParams)
		logger.Debug("Request for route activate received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", actParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		rte, ok := sim.Routes[actParams.ID]
		if !ok {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown route: %s", actParams.ID))
			return
		}
		err = rte.Activate(actParams.Persistent)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("cannot activate route %s: %s", actParams.ID, err))
			return
		}
		ch <- NewOkResponse(req.ID, fmt.Sprintf("Route %s activated successfully", actParams.ID))
	case "deactivate":
		var idParams = struct {
			ID string `json:"id"`
		}{}
		err := json.Unmarshal(req.Params, &idParams)
		logger.Debug("Request for route deactivate received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", idParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		rte, ok := sim.Routes[idParams.ID]
		if !ok {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown route: %s", idParams.ID))
			return
		}
		err = rte.Deactivate()
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("cannot deactivate route %s: %s", idParams.ID, err))
			return
		}
		ch <- NewOkResponse(req.ID, fmt.Sprintf("Route %s deactivated successfully", idParams.ID))
	default:
		ch <- NewErrorResponse(req.ID, fmt.Errorf("unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

var _ hubObject = new(routeObject)

func init() {
	hub.objects["route"] = new(routeObject)
}
