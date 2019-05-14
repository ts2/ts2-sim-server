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

type simulationObject struct{}

// dispatch processes requests made on the Simulation object
func (s *simulationObject) dispatch(h *Hub, req Request, conn *connection) {
	ch := conn.pushChan
	logger.Debug("Request for simulation received", "submodule", "hub", "object", req.Object, "action", req.Action)
	switch req.Action {
	case "start":
		sim.Start()
		ch <- NewOkResponse(req, "Simulation started successfully")
	case "pause":
		sim.Pause()
		ch <- NewOkResponse(req, "Simulation paused successfully")
	case "isStarted":
		j, err := json.Marshal(sim.IsStarted())
		if err != nil {
			ch <- NewErrorResponse(req, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req, RawJSON(j))
	case "dump":
		data, err := json.Marshal(sim)
		if err != nil {
			ch <- NewErrorResponse(req, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req, data)
	default:
		ch <- NewErrorResponse(req, fmt.Errorf("unknown action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

var _ hubObject = new(simulationObject)

func init() {
	hub.objects["simulation"] = new(simulationObject)
}
