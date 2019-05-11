// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package simulation

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDelayGenerator(t *testing.T) {
	Convey("Testing DelayGenerator", t, func() {
		var dg DelayGenerator
		dg = DelayGenerator{data: []delayTuplet{
			{low: 0, high: 100, prob: 50},
			{low: 100, high: 1000, prob: 30},
			{low: 1000, high: 10000, prob: 20},
		}}
		Convey("DelayGenerator should load from JSON", func() {
			dgData := []byte(`[[0, 100, 50], [100, 1000, 30], [1000, 10000, 20]]`)
			var newDG DelayGenerator
			err := json.Unmarshal(dgData, &newDG)
			So(err, ShouldBeNil)
			So(newDG, ShouldResemble, dg)
			dgData = []byte(`0`)
			err = json.Unmarshal(dgData, &newDG)
			So(err, ShouldBeNil)
			So(newDG, ShouldResemble, DelayGenerator{data: []delayTuplet{{0, 0, 100}}})
		})
		Convey("Invalid JSON should return an error", func() {
			dgData := []byte(`"invalid"`)
			var newDG DelayGenerator
			err := json.Unmarshal(dgData, &newDG)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `DelayGenerator.UnmarshalJSON(): Unparsable JSON: "invalid"`)
		})
		Convey("DelayGenerator should marshal to JSON", func() {
			res, err := json.Marshal(dg)
			So(err, ShouldBeNil)
			So(string(res), ShouldEqual, `[[0,100,50],[100,1000,30],[1000,10000,20]]`)
		})
		Convey("DelayGenerator should yield random values", func() {
			rand.Seed(1)
			So(dg.Yield(), ShouldEqual, 9464*time.Second)
			rand.Seed(2)
			So(dg.Yield(), ShouldEqual, 3385*time.Second)
			rand.Seed(3)
			So(dg.Yield(), ShouldEqual, 65*time.Second)
		})
		Convey("Overflown delay generator should return max value", func() {
			dg2 := DelayGenerator{data: []delayTuplet{
				{low: 0, high: 100, prob: 50},
				{low: 1000, high: 10000, prob: 20},
			}}
			So(dg2.Yield(), ShouldEqual, 10000*time.Second)
		})
		Convey("Wider delay generator should not care", func() {
			dg3 := DelayGenerator{data: []delayTuplet{
				{low: 0, high: 100, prob: 50},
				{low: 100, high: 1000, prob: 30},
				{low: 1000, high: 10000, prob: 30},
			}}
			rand.Seed(1)
			So(dg3.Yield(), ShouldEqual, 9464*time.Second)
		})
		Convey("Checking null generators", func() {
			So(dg.IsNull(), ShouldBeFalse)
			So(DelayGenerator{}.IsNull(), ShouldBeTrue)
			So(DelayGenerator{data: []delayTuplet{{0, 0, 100}}}.IsNull(), ShouldBeTrue)
		})
	})
}
