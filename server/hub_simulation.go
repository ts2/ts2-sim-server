// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package server

import "fmt"

type simulationObject struct{}

// dispatch processes requests made on the Simulation object
func (s *simulationObject) dispatch(h *Hub, req Request, conn *connection) {
	ch := conn.pushChan
	switch req.Action {
	case "start":
		logger.Debug("Request for simulation start received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sim.Start()
		ch <- NewOkResponse("Simulation started successfully")
	case "pause":
		logger.Debug("Request for simulation pause received", "submodule", "hub", "object", req.Object, "action", req.Action)
		sim.Pause()
		ch <- NewOkResponse("Simulation paused successfully")
	default:
		ch <- NewErrorResponse(fmt.Errorf("unknwon action %s/%s", req.Object, req.Action))
		logger.Debug("Request for unknown action received", "submodule", "hub", "object", req.Object, "action", req.Action)
	}
}

var _ hubObject = new(simulationObject)

func init() {
	hub.objects["simulation"] = new(simulationObject)
}
