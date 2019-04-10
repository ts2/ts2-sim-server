// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package simulation_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

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
		Convey("Marshalling / Unmarshalling should work both ways", func() {
			var sim, sim2 simulation.Simulation
			data, _ := ioutil.ReadFile("testdata/demo.json")
			err := json.Unmarshal(data, &sim)
			So(err, ShouldBeNil)

			sData, err := json.Marshal(sim)
			So(err, ShouldBeNil)
			err = json.Unmarshal(sData, &sim2)
			So(err, ShouldBeNil)
			So(sim2.TrackItems, ShouldHaveLength, 22)
		})
	})
}
