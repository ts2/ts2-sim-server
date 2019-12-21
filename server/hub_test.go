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
	"testing"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/smartystreets/goconvey/convey"
	_ "github.com/ts2/ts2-sim-server/plugins/lines"
	_ "github.com/ts2/ts2-sim-server/plugins/points"
	_ "github.com/ts2/ts2-sim-server/plugins/routes"
	_ "github.com/ts2/ts2-sim-server/plugins/signals"
	_ "github.com/ts2/ts2-sim-server/plugins/trains"
	"github.com/ts2/ts2-sim-server/simulation"
)

type trackStruct struct {
	ID           string  `json:"id"`
	TiType       string  `json:"__type__"`
	TsName       string  `json:"name"`
	NextTiID     string  `json:"nextTiId"`
	PreviousTiID string  `json:"previousTiId"`
	TsMaxSpeed   float64 `json:"maxSpeed"`
	TsRealLength float64 `json:"realLength"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	ConflictTiId string  `json:"conflictTiId"`
	PlaceCode    string  `json:"placeCode"`
}

func sendRequestStatus(c *websocket.Conn, object, action, params string) ResponseStatus {
	if params == "" {
		params = "null"
	}
	err := c.WriteJSON(Request{Object: object, Action: action, Params: RawJSON(params)})
	So(err, ShouldBeNil)
	var resp ResponseStatus
	err = c.ReadJSON(&resp)
	So(err, ShouldBeNil)
	So(resp.MsgType, ShouldEqual, TypeResponse)
	return resp
}

func TestHub(t *testing.T) {
	// Wait for server to come up
	time.Sleep(100 * time.Millisecond)
	Convey("Testing hub functions", t, func() {
		c := clientDial(t)
		err := register(t, c, Client, "", "client-secret")
		So(err, ShouldBeNil)
		Convey("Calling unknown object should fail", func() {
			err = c.WriteJSON(Request{Object: "undefined", Action: "undefined"})
			So(err, ShouldBeNil)
			var resp ResponseStatus
			err = c.ReadJSON(&resp)
			So(err, ShouldBeNil)
			So(resp.MsgType, ShouldEqual, TypeResponse)
			So(resp.Data.Status, ShouldEqual, Fail)
			So(resp.Data.Message, ShouldEqual, "Error: unknown object undefined")
		})
		Convey("Option functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "option", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action option/undefined")
			})
			Convey("Listing options", func() {
				err = c.WriteJSON(Request{Object: "option", Action: "list"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				var opts simulation.Options
				err := json.Unmarshal(resp.Data, &opts)
				So(err, ShouldBeNil)
				So(opts.TimeFactor, ShouldEqual, 5)
				So(opts.Version, ShouldEqual, "0.7")
				So(opts.Title, ShouldEqual, "TS2 - Demo & Test Sim")
				So(opts.WarningSpeed, ShouldEqual, 8.34)
			})
			Convey("Setting an option by its name", func() {
				resp := sendRequestStatus(c, "option", "set", `{"name": "Title", "value": "New Title"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
				So(sim.Options.Title, ShouldEqual, "New Title")
			})
			Convey("Setting an option by its json name", func() {
				resp := sendRequestStatus(c, "option", "set", `{"name": "title", "value": "New Title again"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
				So(sim.Options.Title, ShouldEqual, "New Title again")
			})
			Convey("Setting an option with invalid params should fail", func() {
				resp := sendRequestStatus(c, "option", "set", `{"name": [], "value": "Another Title"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: error on parameters: json: cannot unmarshal array into Go struct field .name of type string")
			})
			Convey("Setting an option without value should fail", func() {
				resp := sendRequestStatus(c, "option", "set", `{"name": "title"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: error while setting option: option title cannot have nil value")
			})
			Convey("Setting an option with wrong type should fail", func() {
				resp := sendRequestStatus(c, "option", "set", `{"name": "title", "value": 85}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: error while setting option: cannot assign 85 (float64) to title (string)")
			})
			Convey("Setting an unknown option should fail", func() {
				resp := sendRequestStatus(c, "option", "set", `{"name": "undefined", "value": 85}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: error while setting option: unknown option undefined")
			})
		})
		Convey("Route functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "route", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action route/undefined")
			})
			Convey("Listing routes", func() {
				err = c.WriteJSON(Request{Object: "route", Action: "list"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var rtes map[string]simulation.Route
				err = json.Unmarshal(resp.Data, &rtes)
				So(err, ShouldBeNil)
				So(rtes, ShouldHaveLength, 5)
			})
			Convey("Showing a route", func() {
				err = c.WriteJSON(Request{Object: "route", Action: "show", Params: RawJSON(`{"ids": ["1"]}`)})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var rtes map[string]simulation.Route
				err = json.Unmarshal(resp.Data, &rtes)
				So(err, ShouldBeNil)
				So(rtes, ShouldHaveLength, 1)
				So(rtes, ShouldContainKey, "1")
				So(rtes["1"].BeginSignalId, ShouldEqual, "5")
				So(rtes["1"].EndSignalId, ShouldEqual, "101")
			})
			Convey("Show with a wrong route ID should fail", func() {
				resp := sendRequestStatus(c, "route", "show", `{"ids": ["1", "999"]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown route: 999")
			})
			Convey("Deactivating a route", func() {
				resp := sendRequestStatus(c, "route", "deactivate", `{"id": "2"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)

			})
			Convey("Trying to deactivate an unknown route", func() {
				resp := sendRequestStatus(c, "route", "deactivate", `{"id": "999"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown route: 999")
			})
			Convey("Activating a route", func() {
				resp := sendRequestStatus(c, "route", "activate", `{"id": "1"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
			})
			Convey("Trying to activate an unknown route", func() {
				resp := sendRequestStatus(c, "route", "activate", `{"id": "999"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown route: 999")
			})
			Convey("Trying to activate a conflicting route", func() {
				resp := sendRequestStatus(c, "route", "activate", `{"id": "2"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: cannot activate route 2: Standard Manager vetoed route activation: conflicting route 1 is active")
			})
		})
		Convey("Trains functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "train", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action train/undefined")
			})
			Convey("Listing trains", func() {
				err = c.WriteJSON(Request{Object: "train", Action: "list"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var trains []simulation.Train
				err = json.Unmarshal(resp.Data, &trains)
				So(err, ShouldBeNil)
				So(trains, ShouldHaveLength, 2)
			})
			Convey("Showing a train", func() {
				err = c.WriteJSON(Request{Object: "train", Action: "show", Params: RawJSON(`{"ids": [0]}`)})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var trains []simulation.Train
				err = json.Unmarshal(resp.Data, &trains)
				So(err, ShouldBeNil)
				So(trains, ShouldHaveLength, 1)
				So(trains[0].Status, ShouldEqual, simulation.Inactive)
				So(trains[0].ServiceCode, ShouldEqual, "S001")
				So(trains[0].TrainTypeCode, ShouldEqual, "UT")
			})
			Convey("Show with a wrong train ID should fail", func() {
				resp := sendRequestStatus(c, "train", "show", `{"ids": [0, 999]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown train: 999")

				resp = sendRequestStatus(c, "train", "show", `{"ids": [-1]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown train: -1")

				resp = sendRequestStatus(c, "train", "show", `{"ids": [3]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown train: 3")
			})
			Convey("Reversing a train", func() {
				resp := sendRequestStatus(c, "train", "reverse", `{"id": 0}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unable to reverse train 0: train is not stopped")
				pos := simulation.NewPosition(sim, "2", "1", 20)
				sim.Trains[0].TrainHead = pos
				sim.Trains[0].Speed = 0
				resp = sendRequestStatus(c, "train", "reverse", `{"id": 0}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
				So(resp.Data.Message, ShouldEqual, "train reversed successfully")
				resp = sendRequestStatus(c, "train", "reverse", `{"id": 0}`)
				So(resp.Data.Status, ShouldEqual, Ok)
			})
			Convey("Reverse with a wrong train ID should fail", func() {
				resp := sendRequestStatus(c, "train", "reverse", `{"id": 999}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown train: 999")
			})
			Convey("Setting a service to a train", func() {
				So(sim.Trains[0].ServiceCode, ShouldEqual, "S001")
				resp := sendRequestStatus(c, "train", "setService", `{"id": 0, "service": "S002"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
				So(resp.Data.Message, ShouldEqual, "service assigned successfully")
				So(sim.Trains[0].ServiceCode, ShouldEqual, "S002")
			})
			Convey("SetService with a wrong train ID or ServiceID should fail", func() {
				resp := sendRequestStatus(c, "train", "setService", `{"id": 999, "service": "S002"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown train: 999")
				resp = sendRequestStatus(c, "train", "setService", `{"id": 0, "service": "S042"}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unable to assign service S042 to train 0: unknown service: S042")
			})
			Convey("Resetting a service", func() {
				sim.Trains[0].NextPlaceIndex = 1
				resp := sendRequestStatus(c, "train", "resetService", `{"id": 0}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
				So(resp.Data.Message, ShouldEqual, "service reset successfully")
				So(sim.Trains[0].NextPlaceIndex, ShouldEqual, 0)
			})
			Convey("Resetting a service with a wrong train ID should fail", func() {
				resp := sendRequestStatus(c, "train", "resetService", `{"id": 999}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown train: 999")
			})
			Convey("Ask a train to proceed", func() {
				So(sim.Trains[0].ApplicableAction().Speed, ShouldEqual, simulation.VeryHighSpeed)
				So(sim.Trains[0].ApplicableAction().Target, ShouldEqual, simulation.ASAP)
				resp := sendRequestStatus(c, "train", "proceed", `{"id": 0}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
				So(resp.Data.Message, ShouldEqual, "proceed order passed successfully")
				So(sim.Trains[0].ApplicableAction().Speed, ShouldEqual, 8.34)
				So(sim.Trains[0].ApplicableAction().Target, ShouldEqual, simulation.ASAP)
			})
			Convey("Asking a wrong train ID to proceed should fail", func() {
				resp := sendRequestStatus(c, "train", "proceed", `{"id": 999}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown train: 999")
			})
			Convey("Asking a running train ID to proceed should fail", func() {
				sim.Trains[0].Speed = 5
				resp := sendRequestStatus(c, "train", "proceed", `{"id": 0}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unable to proceed for train 0: train is not stopped")
			})
		})
		Convey("TrackItems functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "trackItem", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action trackItem/undefined")
			})
			Convey("Listing trackItems", func() {
				err = c.WriteJSON(Request{Object: "trackItem", Action: "list"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var trackItems map[string]trackStruct
				err = json.Unmarshal(resp.Data, &trackItems)
				So(err, ShouldBeNil)
				So(trackItems, ShouldHaveLength, 29)
			})
			Convey("Showing a trackItem", func() {
				err = c.WriteJSON(Request{Object: "trackItem", Action: "show", Params: RawJSON(`{"ids": ["2"]}`)})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var trackItems map[string]trackStruct
				err = json.Unmarshal(resp.Data, &trackItems)
				So(err, ShouldBeNil)
				So(trackItems, ShouldHaveLength, 1)
				So(trackItems, ShouldContainKey, "2")
				So(trackItems["2"].TiType, ShouldEqual, simulation.TypeLine)
				So(trackItems["2"].ID, ShouldEqual, "2")
				So(trackItems["2"].PlaceCode, ShouldEqual, "LFT")
			})
			Convey("Show with a wrong trackItem ID should fail", func() {
				resp := sendRequestStatus(c, "trackItem", "show", `{"ids": ["5", "999"]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown trackItem: 999")
			})
		})
		Convey("Places functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "place", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action place/undefined")
			})
			Convey("Listing places", func() {
				err = c.WriteJSON(Request{Object: "place", Action: "list"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var places map[string]trackStruct
				err = json.Unmarshal(resp.Data, &places)
				So(err, ShouldBeNil)
				So(places, ShouldHaveLength, 3)
			})
			Convey("Showing a place", func() {
				err = c.WriteJSON(Request{Object: "place", Action: "show", Params: RawJSON(`{"ids": ["STN"]}`)})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var places map[string]trackStruct
				err = json.Unmarshal(resp.Data, &places)
				So(err, ShouldBeNil)
				So(places, ShouldHaveLength, 1)
				So(places, ShouldContainKey, "STN")
				So(places["STN"].TiType, ShouldEqual, simulation.TypePlace)
				So(places["STN"].ID, ShouldEqual, "20")
				So(places["STN"].PlaceCode, ShouldEqual, "STN")
				So(places["STN"].TsName, ShouldEqual, "STATION")
			})
			Convey("Show with a wrong place ID should fail", func() {
				resp := sendRequestStatus(c, "place", "show", `{"ids": ["LFT", "999"]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown place: 999")
			})
		})
		Convey("TrainTypes functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "trainType", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action trainType/undefined")
			})
			Convey("Listing trainTypes", func() {
				err = c.WriteJSON(Request{Object: "trainType", Action: "list"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var trainTypes map[string]*simulation.TrainType
				err = json.Unmarshal(resp.Data, &trainTypes)
				So(err, ShouldBeNil)
				So(trainTypes, ShouldHaveLength, 2)
			})
			Convey("Showing a trainType", func() {
				err = c.WriteJSON(Request{Object: "trainType", Action: "show", Params: RawJSON(`{"ids": ["UT2"]}`)})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var trainTypes map[string]*simulation.TrainType
				err = json.Unmarshal(resp.Data, &trainTypes)
				So(err, ShouldBeNil)
				So(trainTypes, ShouldHaveLength, 1)
				So(trainTypes, ShouldContainKey, "UT2")
				So(trainTypes["UT2"].Length, ShouldEqual, 140)
				So(trainTypes["UT2"].EmergBraking, ShouldEqual, 1.5)
				So(trainTypes["UT2"].Description, ShouldEqual, "Underground double unit")
			})
			Convey("Show with a wrong trainType ID should fail", func() {
				resp := sendRequestStatus(c, "trainType", "show", `{"ids": ["UT", "999"]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown trainType: 999")
			})
		})
		Convey("Services functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "service", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action service/undefined")
			})
			Convey("Listing services", func() {
				err = c.WriteJSON(Request{Object: "service", Action: "list"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var services map[string]*simulation.Service
				err = json.Unmarshal(resp.Data, &services)
				So(err, ShouldBeNil)
				So(services, ShouldHaveLength, 3)
			})
			Convey("Showing a service", func() {
				err = c.WriteJSON(Request{Object: "service", Action: "show", Params: RawJSON(`{"ids": ["S002"]}`)})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var services map[string]*simulation.Service
				err = json.Unmarshal(resp.Data, &services)
				So(err, ShouldBeNil)
				So(services, ShouldHaveLength, 1)
				So(services, ShouldContainKey, "S002")
				So(services["S002"].Description, ShouldEqual, "STATION->LEFT")
				So(services["S002"].Lines, ShouldHaveLength, 2)
			})
			Convey("Show with a wrong service ID should fail", func() {
				resp := sendRequestStatus(c, "service", "show", `{"ids": ["S001", "999"]}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown service: 999")
			})
		})
		Convey("Simulation functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "simulation", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action simulation/undefined")
			})
			Convey("Dumping simulation", func() {
				err = c.WriteJSON(Request{Object: "simulation", Action: "dump"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var simu simulation.Simulation
				err := json.Unmarshal(resp.Data, &simu)
				So(err, ShouldBeNil)
				So(simu.TrackItems, ShouldHaveLength, 29)
				So(simu.Places, ShouldHaveLength, 3)
				So(simu.Places, ShouldContainKey, "STN")
			})
			Convey("Starting simulation", func() {
				resp := sendRequestStatus(c, "simulation", "start", "")
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
			})
			Convey("checking simulation state", func() {
				err = c.WriteJSON(Request{Object: "simulation", Action: "isStarted"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var isStarted bool
				err := json.Unmarshal(resp.Data, &isStarted)
				So(err, ShouldBeNil)
				So(isStarted, ShouldBeTrue)
			})
			Convey("Stopping simulation", func() {
				resp := sendRequestStatus(c, "simulation", "pause", "")
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)
			})
			Convey("checking simulation state again", func() {
				err = c.WriteJSON(Request{Object: "simulation", Action: "isStarted"})
				So(err, ShouldBeNil)
				var resp Response
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				var isStarted bool
				err := json.Unmarshal(resp.Data, &isStarted)
				So(err, ShouldBeNil)
				So(isStarted, ShouldBeFalse)
			})
		})
		Convey("Server functions", func() {
			Convey("Calling unknown action should fail", func() {
				err = c.WriteJSON(Request{Object: "server", Action: "undefined"})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
				So(resp.Data.Message, ShouldEqual, "Error: unknown action server/undefined")
			})
			Convey("Adding listener for clock, start simulation, check we receive notifications and remove listener", func() {
				err = c.WriteJSON(RequestListener{
					Object: "server",
					Action: "addListener",
					Params: ParamsListener{
						Event: simulation.ClockEvent,
					},
				})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.Data.Status, ShouldEqual, Ok)

				resp = sendRequestStatus(c, "simulation", "start", "")
				So(resp.Data.Status, ShouldEqual, Ok)

				var event ResponseNotification
				err = c.ReadJSON(&event)
				So(err, ShouldBeNil)
				So(event.MsgType, ShouldEqual, TypeNotification)
				So(event.Data.Name, ShouldEqual, simulation.ClockEvent)

				resp = sendRequestStatus(c, "simulation", "pause", "")
				So(resp.Data.Status, ShouldEqual, Ok)

				resp = sendRequestStatus(c, "server", "removeListener", "{\"event\": \"clock\"}")
				So(resp.Data.Status, ShouldEqual, Ok)
			})
			Convey("Adding listener for selected IDs only and check we only receive events for these", func() {
				err = c.WriteJSON(RequestListener{
					Object: "server",
					Action: "addListener",
					Params: ParamsListener{
						Event: simulation.RouteActivatedEvent,
						IDs:   []string{"1"},
					},
				})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)

				err := c.WriteJSON(Request{Object: "route", Action: "activate", Params: RawJSON(`{"id": "1"}`)})
				So(err, ShouldBeNil)
				var haveResponse, haveNotification bool
				for i := 0; i < 2; i++ {
					var r Response
					err = c.ReadJSON(&r)
					So(err, ShouldBeNil)
					switch r.MsgType {
					case TypeResponse:
						So(haveResponse, ShouldBeFalse)
						haveResponse = true
						var rd DataStatus
						err := json.Unmarshal(r.Data, &rd)
						So(err, ShouldBeNil)
						So(rd.Status, ShouldEqual, Ok)
					case TypeNotification:
						So(haveNotification, ShouldBeFalse)
						haveNotification = true
						var de DataEvent
						err := json.Unmarshal(r.Data, &de)
						So(err, ShouldBeNil)
						So(de.Name, ShouldEqual, simulation.RouteActivatedEvent)
					}
				}

				resp = sendRequestStatus(c, "route", "deactivate", `{"id": "1"}`)
				So(resp.Data.Status, ShouldEqual, Ok)

				resp = sendRequestStatus(c, "route", "activate", `{"id": "2"}`)
				// Fails because route 1 did not have time to deactivate
				So(resp.Data.Status, ShouldEqual, Fail)

				err = c.WriteJSON(RequestListener{
					Object: "server",
					Action: "removeListener",
					Params: ParamsListener{
						Event: simulation.RouteActivatedEvent,
						IDs:   []string{"1"},
					},
				})
				So(err, ShouldBeNil)
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)

			})
			Convey("Invalid listeners requests should fail", func() {
				resp := sendRequestStatus(c, "server", "addListener", `{"event": []}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)

				resp = sendRequestStatus(c, "server", "removeListener", `{"event": []}`)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Fail)
			})
			Convey("Renotify should send back the last notifications", func() {
				err = c.WriteJSON(RequestListener{
					Object: "server",
					Action: "addListener",
					Params: ParamsListener{
						Event: simulation.RouteActivatedEvent,
						IDs:   []string{"1"},
					},
				})
				So(err, ShouldBeNil)
				var resp ResponseStatus
				err = c.ReadJSON(&resp)
				So(err, ShouldBeNil)
				So(resp.MsgType, ShouldEqual, TypeResponse)
				So(resp.Data.Status, ShouldEqual, Ok)

				err := c.WriteJSON(Request{Object: "server", Action: "renotify"})
				So(err, ShouldBeNil)
				var haveResponse, haveNotification bool
				for i := 0; i < 2; i++ {
					var r Response
					err = c.ReadJSON(&r)
					So(err, ShouldBeNil)
					switch r.MsgType {
					case TypeResponse:
						So(haveResponse, ShouldBeFalse)
						haveResponse = true
						var rd DataStatus
						err := json.Unmarshal(r.Data, &rd)
						So(err, ShouldBeNil)
						So(rd.Status, ShouldEqual, Ok)
					case TypeNotification:
						So(haveNotification, ShouldBeFalse)
						haveNotification = true
						var de DataEvent
						err := json.Unmarshal(r.Data, &de)
						So(err, ShouldBeNil)
						So(de.Name, ShouldEqual, simulation.RouteActivatedEvent)
					}
				}
			})
		})
		Reset(func() {
			err := c.Close()
			So(err, ShouldBeNil)
		})
	})
}
