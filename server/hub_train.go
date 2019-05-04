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

type trainObject struct{}

func (s *trainObject) objectName() string {
	return "train"
}

// dispatch processes requests made on the Service object
func (t *trainObject) dispatch(h *Hub, req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {

	case "list":
		logger.Debug("Request for train list received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sl, err := json.Marshal(sim.Trains)
		if err != nil {
			ch <- NewErrorResponse(req, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req, sl)

	case "show":
		var idsParams = struct {
			IDs []int `json:"ids"`
		}{}
		err := json.Unmarshal(req.Params, &idsParams)
		logger.Debug("Request for train show received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", idsParams)
		if err != nil {
			ch <- NewErrorResponse(req, fmt.Errorf("internal error: %s", err))
			return
		}
		ts := make([]*simulation.Train, len(idsParams.IDs))
		for i, id := range idsParams.IDs {
			if id < 0 || id >= len(sim.Trains) {
				ch <- NewErrorResponse(req, fmt.Errorf("unknown train: %d", id))
				return
			}
			ts[i] = sim.Trains[id]
		}
		tid, err := json.Marshal(ts)
		if err != nil {
			ch <- NewErrorResponse(req, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req, tid)

	default:
		ch <- NewErrorResponse(req, fmt.Errorf("unknown action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

var _ hubObject = new(trainObject)

func init() {
	hub.objects["train"] = new(trainObject)
}
