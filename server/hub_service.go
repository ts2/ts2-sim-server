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

type serviceObject struct{}

// dispatch processes requests made on the Service object
func (s *serviceObject) dispatch(h *Hub, req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {
	case "list":
		logger.Debug("Request for service list received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sl, err := json.Marshal(sim.Services)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req.ID, sl)
	case "show":
		var idsParams = struct {
			IDs []string `json:"ids"`
		}{}
		err := json.Unmarshal(req.Params, &idsParams)
		logger.Debug("Request for service show received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", idsParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		sl := make(map[string]*simulation.Service)
		for _, id := range idsParams.IDs {
			sl[id] = sim.Services[id]
		}
		tid, err := json.Marshal(sl)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req.ID, tid)
	default:
		ch <- NewErrorResponse(req.ID, fmt.Errorf("unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

var _ hubObject = new(serviceObject)

func init() {
	hub.objects["service"] = new(serviceObject)
}
