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

type optionObject struct{}

// dispatch processes requests made on the Option object
func (s *optionObject) dispatch(_ *Hub, req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {
	case "list":
		logger.Debug("Request for option list received", "submodule", "hub", "object", req.Object, "action", req.Action)
		opts, err := json.Marshal(sim.Options)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("internal error: %s", err))
			return
		}
		ch <- NewResponse(req.ID, opts)
	case "set":
		var setParams = struct {
			Name  string      `json:"name"`
			Value interface{} `json:"value"`
		}{}
		err := json.Unmarshal(req.Params, &setParams)
		logger.Debug("Request for option set received", "submodule", "hub", "object", req.Object, "action", req.Action, "params", req.Params)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("error on parameters: %s", err))
			return
		}
		err = sim.Options.Set(setParams.Name, setParams.Value)
		if err != nil {
			ch <- NewErrorResponse(req.ID, fmt.Errorf("error while setting option: %s", err))
			return
		}
		ch <- NewOkResponse(req.ID, fmt.Sprintf("option %s set successfully to %v", setParams.Name, setParams.Value))
	default:
		ch <- NewErrorResponse(req.ID, fmt.Errorf("unknown action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

var _ hubObject = new(optionObject)

func init() {
	hub.objects["option"] = new(optionObject)
}
