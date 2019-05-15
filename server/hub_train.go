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

// dispatch processes requests made on the Service object
func (t *trainObject) dispatch(h *Hub, req Request, conn *connection) {
	logger.Debug("Request for train received", "submodule", "hub", "object", req.Object, "action", req.Action)
	ch := conn.pushChan
	switch req.Action {
	case "list":
		sl, err := json.Marshal(sim.Trains)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req.ID, sl)
	case "show":
		var idsParams = struct {
			IDs []int `json:"ids"`
		}{}
		err := json.Unmarshal(req.Params, &idsParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ts := make([]*simulation.Train, len(idsParams.IDs))
		for i, id := range idsParams.IDs {
			if id < 0 || id >= len(sim.Trains) {
				ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown train: %d", id))
				return
			}
			ts[i] = sim.Trains[id]
		}
		tid, err := json.Marshal(ts)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req.ID, tid)
	case "reverse":
		var idParams = struct {
			ID int `json:"id"`
		}{}
		err := json.Unmarshal(req.Params, &idParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		if idParams.ID < 0 || idParams.ID >= len(sim.Trains) {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown train: %d", idParams.ID))
			return
		}
		train := sim.Trains[idParams.ID]
		if err = train.Reverse(); err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unable to reverse train %d: %s", idParams.ID, err))
			return
		}
		ch <- NewOkResponse(req.ID, "train reversed successfully")
	case "setService":
		var smParams = struct {
			ID      int    `json:"id"`
			Service string `json:"service"`
		}{}
		err := json.Unmarshal(req.Params, &smParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		if smParams.ID < 0 || smParams.ID >= len(sim.Trains) {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown train: %d", smParams.ID))
			return
		}
		if err = sim.Trains[smParams.ID].AssignService(smParams.Service); err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unable to assign service %s to train %d: %s", smParams.Service, smParams.ID, err))
			return
		}
		ch <- NewOkResponse(req.ID, "service assigned successfully")
	case "resetService":
		var idParams = struct {
			ID int `json:"id"`
		}{}
		err := json.Unmarshal(req.Params, &idParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		if idParams.ID < 0 || idParams.ID >= len(sim.Trains) {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown train: %d", idParams.ID))
			return
		}
		train := sim.Trains[idParams.ID]
		_ = train.ResetService()
		ch <- NewOkResponse(req.ID, "service reset successfully")
	case "proceed":
		var idParams = struct {
			ID int `json:"id"`
		}{}
		err := json.Unmarshal(req.Params, &idParams)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		if idParams.ID < 0 || idParams.ID >= len(sim.Trains) {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown train: %d", idParams.ID))
			return
		}
		train := sim.Trains[idParams.ID]
		if err = train.ProceedWithCaution(); err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("unable to proceed for train %d: %s", idParams.ID, err))
			return
		}
		ch <- NewOkResponse(req.ID, "proceed order passed successfully")
	default:
		ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

var _ hubObject = new(trainObject)

func init() {
	hub.objects["train"] = new(trainObject)
}
