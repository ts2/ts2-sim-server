// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package simulation_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	_ "github.com/ts2/ts2-sim-server/plugins/lines"
	_ "github.com/ts2/ts2-sim-server/plugins/points"
	_ "github.com/ts2/ts2-sim-server/plugins/routes"
	_ "github.com/ts2/ts2-sim-server/plugins/signals"
	_ "github.com/ts2/ts2-sim-server/plugins/trains"
	"github.com/ts2/ts2-sim-server/simulation"
)

func TestMarshalling(t *testing.T) {
	Convey("JSON Marshalling test", t, func() {
		var sim simulation.Simulation
		data, _ := ioutil.ReadFile("testdata/demo.json")
		err := json.Unmarshal(data, &sim)
		So(err, ShouldBeNil)
		Convey("Marshalling / Unmarshalling should work both ways", func() {
			sData, err := json.Marshal(sim)
			So(err, ShouldBeNil)

			var sim2 simulation.Simulation
			err = json.Unmarshal(sData, &sim2)
			So(err, ShouldBeNil)
			So(sim2.TrackItems, ShouldHaveLength, 29)
			So(sim2.Routes, ShouldHaveLength, 5)
			So(sim2.Trains, ShouldHaveLength, 2)
			So(sim2.Services, ShouldHaveLength, 3)
			So(sim2.Options.TimeFactor, ShouldEqual, 5)
		})
	})
}

func TestSimulationRun(t *testing.T) {
	Convey("Testing simulation runs", t, func() {
		var sim simulation.Simulation
		data, _ := ioutil.ReadFile("testdata/demo.json")
		err := json.Unmarshal(data, &sim)
		So(err, ShouldBeNil)
		err = sim.Initialize()
		sim.Trains[0].AppearTime = simulation.ParseTime("05:00:00")
		So(err, ShouldBeNil)
		Convey("Starting and stopping the simulation should work", func() {
			So(sim.Options.CurrentTime, ShouldResemble, simulation.ParseTime("06:00:00"))
			So(sim.Trains[0].TrainHead.TrackItemID, ShouldEqual, "2")
			So(sim.Trains[0].TrainHead.PreviousItemID, ShouldEqual, "1")
			So(sim.Trains[0].TrainHead.PositionOnTI, ShouldEqual, 3)
			So(sim.TrackItems["3"].(*simulation.SignalItem).ActiveAspect().Name, ShouldEqual, "UK_CLEAR")
			sim.Start()
			time.Sleep(600 * time.Millisecond)
			sim.Pause()
			sim.Options.TrackCircuitBased = true
			So(sim.Options.CurrentTime, ShouldResemble, simulation.ParseTime("06:00:02.5"))
			So(sim.Trains[0].TrainHead.TrackItemID, ShouldEqual, "2")
			So(sim.Trains[0].TrainHead.PreviousItemID, ShouldEqual, "1")
			So(sim.Trains[0].TrainHead.PositionOnTI, ShouldEqual, 18.625)
			So(sim.Trains[0].Speed, ShouldEqual, 6.25)
			time.Sleep(600 * time.Millisecond)
			So(sim.Options.CurrentTime, ShouldResemble, simulation.ParseTime("06:00:02.5"))
			So(sim.Options.CurrentTime, ShouldResemble, simulation.ParseTime("06:00:02.5"))
			So(sim.Trains[0].TrainHead.TrackItemID, ShouldEqual, "2")
			So(sim.Trains[0].TrainHead.PreviousItemID, ShouldEqual, "1")
			So(sim.Trains[0].TrainHead.PositionOnTI, ShouldEqual, 18.625)
			So(sim.Trains[0].Speed, ShouldEqual, 6.25)
			So(sim.TrackItems["3"].(*simulation.SignalItem).ActiveAspect().Name, ShouldEqual, "UK_DANGER")
			So(sim.TrackItems["5"].(*simulation.SignalItem).ActiveAspect().Name, ShouldEqual, "UK_CLEAR")
			err := sim.Routes["1"].Deactivate()
			So(err, ShouldBeNil)
			err = sim.Routes["2"].Activate(false)
			So(err, ShouldBeNil)
			So(sim.TrackItems["5"].(*simulation.SignalItem).ActiveAspect().Name, ShouldEqual, "UK_DANGER")
			sim.Start()
			time.Sleep(7 * time.Second)
			sim.Pause()
			So(sim.TrackItems["5"].(*simulation.SignalItem).ActiveAspect().Name, ShouldEqual, "UK_CAUTION")
			So(sim.TrackItems["3"].(*simulation.SignalItem).ActiveAspect().Name, ShouldEqual, "UK_CLEAR")
		})
	})
}
