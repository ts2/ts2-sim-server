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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFollowingItem(t *testing.T) {
	Convey("Testing FollowingItem method", t, func() {
		var sim Simulation
		if err := json.Unmarshal(loadSim("testdata/demo.json"), &sim); err != nil {
			t.Errorf("Options: error while loading JSON: %s", err)
		}
		ei1, _ := sim.TrackItems["1"]
		li2, _ := sim.TrackItems["2"]
		si3, _ := sim.TrackItems["3"]
		li4, _ := sim.TrackItems["4"]
		ti6, _ := sim.TrackItems["6"]
		pi7, _ := sim.TrackItems["7"]
		ti8, _ := sim.TrackItems["8"]
		ti14, _ := sim.TrackItems["14"]
		Convey("Following items should match", func() {
			fi1, _ := li2.FollowingItem(si3, 0)
			fi1b, _ := li2.FollowingItem(ei1, 0)
			fi3, _ := si3.FollowingItem(li4, 1)
			_, nle := ei1.FollowingItem(si3, 0)
			fipr, _ := pi7.FollowingItem(ti6, 1)
			fipn, _ := pi7.FollowingItem(ti6, 0)
			ficr, _ := pi7.FollowingItem(ti14, 0)
			ficrb, _ := pi7.FollowingItem(ti14, 1)
			ficn, _ := pi7.FollowingItem(ti8, 0)
			So(fi1, ShouldEqual, ei1)
			So(fi1b, ShouldEqual, si3)
			So(fi3, ShouldEqual, li2)
			So(nle, ShouldHaveSameTypeAs, ItemsNotLinkedError{})
			So(nle.Error(), ShouldEqual, "TrackItems 1 and 3 are not linked")
			So(fipr, ShouldEqual, ti14)
			So(fipn, ShouldEqual, ti8)
			So(ficr, ShouldEqual, ti6)
			So(ficrb, ShouldEqual, ti6)
			So(ficn, ShouldEqual, ti6)
		})
	})
}
