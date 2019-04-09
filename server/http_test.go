// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package server

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHTTP(t *testing.T) {
	// Wait for server to come up
	time.Sleep(100 * time.Millisecond)
	Convey("Testing HTTP websocket client", t, func() {
		Convey("Normal GET / ", func() {
			res, err := http.Get("http://127.0.0.1:22222")
			So(err, ShouldBeNil)
			So(res.StatusCode, ShouldEqual, http.StatusOK)
		})
		Convey("Wrong URI should fail", func() {
			res, err := http.Get("http://127.0.0.1:22222/undefined")
			So(err, ShouldBeNil)
			So(res.StatusCode, ShouldEqual, http.StatusNotFound)

		})
		Convey("Wrong method should fail", func() {
			res, err := http.Post("http://127.0.0.1:22222", "application/json", nil)
			So(err, ShouldBeNil)
			So(res.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
		})
	})
}
