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

package simulation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSimulationLoading(t *testing.T) {
	Convey("Loading a simulation should not create an error", t, func() {
		var sim Simulation
		err := json.Unmarshal(loadSim("testdata/demo.json"), &sim)
		So(err, ShouldBeNil)
		Convey("Options should be all loaded", func() {
			So(sim.Options.CurrentScore, ShouldEqual, 0)
			So(sim.Options.CurrentTime, ShouldResemble, ParseTime("06:00:00"))
			So(sim.Options.DefaultDelayAtEntry.Equals(DelayGenerator{[]delayTuplet{{0, 0, 100}}}), ShouldBeTrue)
			So(sim.Options.DefaultMinimumStopTime.Equals(DelayGenerator{[]delayTuplet{{20, 40, 90}, {40, 120, 10}}}), ShouldBeTrue)
			So(sim.Options.DefaultMaxSpeed, ShouldEqual, 18.06)
			So(sim.Options.DefaultSignalVisibility, ShouldEqual, 100.0)
			So(sim.Options.Description, ShouldEqual, "This is a developers test/demo simulation!")
			So(sim.Options.Title, ShouldEqual, "TS2 - Demo & Test Sim")
			So(sim.Options.TimeFactor, ShouldEqual, 5)
			So(sim.Options.Version, ShouldEqual, "0.7")
			So(sim.Options.WarningSpeed, ShouldEqual, 8.34)
			So(sim.Options.TrackCircuitBased, ShouldEqual, false)
		})
		Convey("Routes should be correctly loaded", func() {
			So(sim.Routes, ShouldHaveLength, 5)

			So(sim.Routes, ShouldContainKey, "1")
			r1 := sim.Routes["1"]
			So(r1.ID(), ShouldEqual, "1")
			si5, _ := sim.TrackItems["5"].(*SignalItem)
			si101, _ := sim.TrackItems["101"].(*SignalItem)
			So(r1.BeginSignal(), ShouldEqual, si5)
			So(r1.EndSignal(), ShouldEqual, si101)
			items := []string{"5", "6", "7", "8", "9", "10", "101"}
			for i, pos := range r1.Positions {
				So(pos.TrackItem().ID(), ShouldEqual, items[i])
			}
			d1, ok := r1.Directions["7"]
			So(ok, ShouldBeTrue)
			So(d1, ShouldEqual, DirectionNormal)
			So(r1.InitialState, ShouldEqual, Activated)
			So(r1.State(), ShouldEqual, Activated)

			So(sim.Routes, ShouldContainKey, "4")
			r4, ok := sim.Routes["4"]
			So(ok, ShouldBeTrue)
			si15, _ := sim.TrackItems["15"].(*SignalItem)
			si3, _ := sim.TrackItems["3"].(*SignalItem)
			So(r4.BeginSignal(), ShouldEqual, si15)
			So(r4.EndSignal(), ShouldEqual, si3)
			items = []string{"15", "14", "7", "6", "5", "4", "3"}
			for i, pos := range r4.Positions {
				So(pos.TrackItem().ID(), ShouldEqual, items[i])
			}
			So(r4.InitialState, ShouldEqual, Deactivated)
			So(r4.State(), ShouldEqual, Deactivated)
		})
		Convey("TrackItems loading", func() {
			Convey("TrackItems links should be ok", func() {
				err := sim.checkTrackItemsLinks()
				So(err, ShouldBeNil)
			})
			Convey("All 25 items should be loaded", func() {
				items := map[string]TrackItem{
					"1":   new(EndItem),
					"2":   new(LineItem),
					"3":   new(SignalItem),
					"4":   new(LineItem),
					"5":   new(SignalItem),
					"6":   new(LineItem),
					"7":   new(PointsItem),
					"8":   new(LineItem),
					"9":   new(SignalItem),
					"10":  new(LineItem),
					"101": new(SignalItem),
					"102": new(LineItem),
					"103": new(InvisibleLinkItem),
					"104": new(LineItem),
					"11":  new(SignalItem),
					"12":  new(LineItem),
					"13":  new(EndItem),
					"14":  new(LineItem),
					"15":  new(SignalItem),
					"16":  new(LineItem),
					"17":  new(SignalItem),
					"18":  new(EndItem),
					"22":  new(PlatformItem),
					"23":  new(PlatformItem),
					"24":  new(TextItem),
					"25":  new(TextItem),
				}
				for id, typ := range items {
					it, ok := sim.TrackItems[id]
					So(ok, ShouldBeTrue)
					So(it, ShouldHaveSameTypeAs, typ)
				}
			})
			Convey("All places should be loaded too", func() {
				places := []string{"LFT", "STN", "RGT"}
				for _, place := range places {
					pl, ok := sim.Places[place]
					So(ok, ShouldBeTrue)
					So(pl, ShouldHaveSameTypeAs, new(Place))
				}
			})
			Convey("Checking a items properties", func() {
				So(sim.TrackItems["1"].Name(), ShouldEqual, "")
				So(sim.TrackItems["1"].NextItem(), ShouldEqual, nil)
				So(sim.TrackItems["1"].PreviousItem(), ShouldEqual, sim.TrackItems["2"])
				So(sim.TrackItems["1"].Origin(), ShouldResemble, Point{0.0, 0.0})
				So(sim.TrackItems["1"].ID(), ShouldEqual, "1")
				So(sim.TrackItems["2"].PreviousItem(), ShouldEqual, sim.TrackItems["1"])
				So(sim.TrackItems["2"].(*LineItem).TrackCode(), ShouldEqual, "")
				So(sim.TrackItems["2"].Place(), ShouldEqual, sim.Places["LFT"])
				So(sim.TrackItems["2"].MaxSpeed(), ShouldEqual, 27.77)
				So(sim.TrackItems["2"].RealLength(), ShouldEqual, 400.0)
				So(sim.TrackItems["3"].Origin(), ShouldResemble, sim.TrackItems["4"].Origin())
				So(sim.TrackItems["4"].Name(), ShouldEqual, "Sample Name")
				So(sim.TrackItems["4"].MaxSpeed(), ShouldEqual, 18.06)
				So(sim.TrackItems["6"].MaxSpeed(), ShouldEqual, 10.0)
				So(sim.TrackItems["6"].RealLength(), ShouldEqual, 200.0)
				So(sim.TrackItems["6"].Origin(), ShouldResemble, Point{200.0, 0.0})
				So(sim.TrackItems["6"].PreviousItem(), ShouldEqual, sim.TrackItems["5"])
				So(sim.TrackItems["6"], ShouldEqual, sim.TrackItems["7"].PreviousItem())
				So(sim.TrackItems["8"], ShouldEqual, sim.TrackItems["7"].NextItem())
				So(sim.TrackItems["9"].(*SignalItem).Reversed(), ShouldBeTrue)
				So(sim.TrackItems["10"].Place(), ShouldEqual, sim.Places["STN"])
				So(sim.TrackItems["10"].(*LineItem).TrackCode(), ShouldEqual, "1")
				So(sim.TrackItems["11"].(*SignalItem).SignalType(), ShouldEqual, sim.SignalLib.Types["UK_2_AUTOMATIC"])
				So(sim.TrackItems["12"].Place(), ShouldEqual, sim.Places["RGT"])
				So(sim.TrackItems["12"].(*LineItem).TrackCode(), ShouldEqual, "")
				So(sim.TrackItems["13"].Origin(), ShouldResemble, Point{600.0, 0.0})
				So(sim.TrackItems["7"].(*PointsItem).ReverseItem(), ShouldEqual, sim.TrackItems["14"])
				So(sim.TrackItems["7"].(*PointsItem).CommonEnd(), ShouldResemble, Point{-5.0, 0.0})
				So(sim.TrackItems["7"].(*PointsItem).ReverseEnd(), ShouldResemble, Point{5.0, 5.0})
				So(sim.TrackItems["7"].(*PointsItem).NormalEnd(), ShouldResemble, Point{5.0, 0.0})
				So(sim.TrackItems["15"].(*SignalItem).Reversed(), ShouldBeTrue)
				So(sim.TrackItems["15"].PreviousItem(), ShouldEqual, sim.TrackItems["16"])
				So(sim.TrackItems["16"].Place(), ShouldEqual, sim.Places["STN"])
				So(sim.TrackItems["16"].(*LineItem).TrackCode(), ShouldEqual, "2")
				So(sim.TrackItems["17"].(*SignalItem).Reversed(), ShouldBeFalse)
				So(sim.TrackItems["17"].(*SignalItem).SignalType(), ShouldEqual, sim.SignalLib.Types["BUFFER"])
				So(sim.TrackItems["18"].PreviousItem(), ShouldEqual, sim.TrackItems["17"])
				So(sim.TrackItems["18"].NextItem(), ShouldBeNil)
				So(sim.TrackItems["22"].Origin(), ShouldResemble, Point{300, 35})
				So(sim.TrackItems["22"].End(), ShouldResemble, Point{390, 50})
				So(sim.TrackItems["23"].Place(), ShouldEqual, sim.Places["STN"])
				So(sim.TrackItems["23"].(*PlatformItem).TrackCode(), ShouldEqual, "1")
				So(sim.TrackItems["24"].Name(), ShouldEqual, "2")
				So(sim.TrackItems["25"].Name(), ShouldEqual, "1")
				So(sim.TrackItems["5"].CustomProperty("ROUTES_SET")["UK_DANGER"], ShouldHaveLength, 1)
				So(sim.TrackItems["5"].CustomProperty("ROUTES_SET")["UK_DANGER"][0], ShouldEqual, "2")
				So(sim.TrackItems["5"].CustomProperty("TRAIN_NOT_PRESENT_ON_ITEMS")["UK_DANGER"], ShouldHaveLength, 2)
				So(sim.TrackItems["5"].CustomProperty("TRAIN_NOT_PRESENT_ON_ITEMS")["UK_DANGER"][0], ShouldEqual, "4")
				So(sim.TrackItems["5"].CustomProperty("TRAIN_NOT_PRESENT_ON_ITEMS")["UK_DANGER"][1], ShouldEqual, "3")
			})
		})
		Convey("TrainTypes should be correctly loaded", func() {
			So(sim.TrainTypes, ShouldHaveLength, 2)
			So(sim.TrainTypes, ShouldContainKey, "UT")
			So(sim.TrainTypes, ShouldContainKey, "UT2")
			tt := sim.TrainTypes["UT"]
			tt2 := sim.TrainTypes["UT2"]
			So(tt.Description, ShouldEqual, "Underground train")
			So(tt.EmergBraking, ShouldEqual, 1.5)
			So(tt.Length, ShouldEqual, 70.0)
			So(tt.MaxSpeed, ShouldEqual, 25.0)
			So(tt2.Elements()[0], ShouldEqual, tt)
			So(tt2.Elements()[1], ShouldEqual, tt)
		})
		Convey("Services should all be loaded", func() {
			So(sim.Services, ShouldHaveLength, 3)
			So(sim.Services, ShouldContainKey, "S001")
			So(sim.Services, ShouldContainKey, "S002")
			So(sim.Services, ShouldContainKey, "S003")
			s1 := sim.Services["S001"]
			s2 := sim.Services["S002"]
			So(s1.Description, ShouldEqual, "LEFT->STATION")
			So(s1.PlannedTrainType(), ShouldEqual, sim.TrainTypes["UT"])
			So(s1.Lines, ShouldHaveLength, 2)
			So(s1.Lines[0].MustStop, ShouldBeFalse)
			So(s1.Lines[0].Place(), ShouldEqual, sim.Places["LFT"])
			So(s1.Lines[0].ScheduledArrivalTime, ShouldResemble, Time{})
			So(s1.Lines[0].ScheduledDepartureTime, ShouldResemble, ParseTime("06:00:30"))
			So(s1.Lines[0].TrackCode, ShouldBeEmpty)
			So(s1.PostActions, ShouldHaveLength, 2)
			So(s1.PostActions[0].ActionCode, ShouldEqual, actionSetService)
			So(s1.PostActions[0].ActionParam, ShouldEqual, "S002")
			So(s1.PostActions[1].ActionCode, ShouldEqual, actionReverse)
			So(s1.PostActions[1].ActionParam, ShouldBeEmpty)
			So(s2.Description, ShouldEqual, "STATION->LEFT")
			So(s2.PostActions, ShouldHaveLength, 0)
		})
		Convey("Trains loading should be Ok", func() {
			So(sim.Trains, ShouldHaveLength, 2)
			tr := sim.Trains[0]
			So(tr.Service(), ShouldEqual, sim.Services["S001"])
			So(tr.TrainType(), ShouldEqual, sim.TrainTypes["UT"])
			So(tr.TrainHead.Equals(Position{&sim, sim.TrackItems["2"].ID(), sim.TrackItems["1"].ID(), 3.0}), ShouldBeTrue)
			So(tr.AppearTime, ShouldResemble, ParseTime("06:00:00"))
			So(tr.InitialDelay.Equals(DelayGenerator{[]delayTuplet{{0, 0, 100}}}), ShouldBeTrue)
			So(tr.InitialSpeed, ShouldEqual, 5.0)
			So(tr.Speed, ShouldEqual, 5)
			So(tr.NextPlaceIndex, ShouldEqual, 0)
			So(tr.Status, ShouldEqual, Inactive)
			So(tr.StoppedTime, ShouldEqual, 0)
		})
		Convey("MessageLogger should be fully loaded", func() {
			So(sim.MessageLogger.Messages, ShouldHaveLength, 1)
			So(sim.MessageLogger.Messages[0], ShouldResemble, Message{playerWarningMsg, "Test message"})
		})
		Convey("SignalLibrary should be correctly loaded", func() {
			So(sim.SignalLib.Types, ShouldHaveLength, 3)
			So(sim.SignalLib.Aspects, ShouldHaveLength, 4)
			So(sim.SignalLib.Aspects, ShouldContainKey, "BUFFER")
			bufferAspect := sim.SignalLib.Aspects["BUFFER"]
			So(bufferAspect.Name, ShouldEqual, "BUFFER")
			So(bufferAspect.LineStyle, ShouldEqual, bufferStyle)
			So(bufferAspect.OuterShapes, ShouldEqual, [6]signalShape{noneShape, noneShape, noneShape, noneShape, noneShape, noneShape})
			black, _ := FromHex("#000000")
			So(bufferAspect.OuterColors, ShouldEqual, [6]Color{black, black, black, black, black, black})
			So(bufferAspect.Shapes, ShouldEqual, [6]signalShape{noneShape, noneShape, noneShape, noneShape, noneShape, noneShape})
			So(bufferAspect.ShapesColors, ShouldEqual, [6]Color{black, black, black, black, black, black})
			So(bufferAspect.Actions, ShouldHaveLength, 1)
			So(bufferAspect.Actions[0].Target, ShouldEqual, BeforeThisSignal)
			So(bufferAspect.Actions[0].Speed, ShouldEqual, 0.0)

			So(sim.SignalLib.Aspects, ShouldContainKey, "UK_DANGER")
			dangerAspect := sim.SignalLib.Aspects["UK_DANGER"]
			So(dangerAspect.LineStyle, ShouldEqual, lineStyle)
			So(dangerAspect.Shapes, ShouldEqual, [6]signalShape{circleShape, noneShape, noneShape, noneShape, noneShape, noneShape})
			red, _ := FromHex("#FF0000")
			So(dangerAspect.ShapesColors, ShouldEqual, [6]Color{red, black, black, black, black, black})

			So(sim.SignalLib.Aspects, ShouldContainKey, "UK_CAUTION")
			cautionAspect := sim.SignalLib.Aspects["UK_CAUTION"]
			So(cautionAspect.Actions[0].Target, ShouldEqual, BeforeNextSignal)
			So(cautionAspect.Actions[0].Speed, ShouldEqual, 0.0)
		})
	})
	Convey("Testing simulation loading errors", t, func() {
		var sim Simulation
		Convey("Loading wrong JSON should fail", func() {
			err := json.Unmarshal([]byte(`{"this": ["is": "erroneous JSON"]}`), &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid character ':' after array element")

			err = json.Unmarshal([]byte(`{"routes": []}`), &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unable to decode simulation JSON: json: cannot unmarshal array into Go struct field auxSim.routes of type map[string]*simulation.Route")
		})
		Convey("Loading simulation with wrong version should fail", func() {
			err := json.Unmarshal([]byte(`{"options": {"version": "0.6"}, "signalLibrary": {}}`), &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, fmt.Sprintf("version mismatch: server: %s / file: 0.6", Version))
		})
		Convey("Loading with wrong signalLibrary should fail", func() {
			err := json.Unmarshal([]byte(`
{"options": {
	"version": "0.7"
},
"signalLibrary": {
	"signalTypes": {
		"BUFFER": {
			"states": [
				{
					"aspectName": "BUFFER",
					"conditions": {}
				}
			]
		}
	}
}}`), &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "error initializing signal Library: no aspect with code BUFFER found")
		})
		Convey("Wrong trackItem type should fail", func() {
			err := json.Unmarshal([]byte(`
{
	"options": {
		"version": "0.7"
	},
	"trackItems": {
		"3": {
			"__type__": "undefined"
		}
	}
}`), &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `unknown TrackItem type: "undefined"`)
		})
		Convey("Wrong trackItem definition should fail", func() {
			err := json.Unmarshal([]byte(`
{
	"options": {
		"version": "0.7"
	},
	"trackItems": {
		"3": {
			"__type__": "SignalItem",
			"name": []
		}
	}
}`), &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `unable to decode "SignalItem": {
			"__type__": "SignalItem",
			"name": []
		}. json: cannot unmarshal array into Go struct field SignalItem.name of type string`)
		})
		Convey("Simulation with wrong links should fail loading", func() {
			data, _ := ioutil.ReadFile("testdata/badlinks.json")
			err := json.Unmarshal(data, &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldBeIn, []string{
				"inconsistent link at (0.000000, 0.000000) between 1 and 3",
				"inconsistent link at (90.000000, 0.000000) between 2 and 1",
			})
		})
		Convey("Simulation with wrong routes should fail loading", func() {
			data, _ := ioutil.ReadFile("testdata/badroutes.json")
			err := json.Unmarshal(data, &sim)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "error initializing route 1: route Error: unable to link signal 11 to signal 5")
		})
	})
}
