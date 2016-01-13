/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package server

import (
	"encoding/json"
	"net/http"

	//"github.com/fatih/structs"
	"github.com/ajstarks/svgo"

	//"github.com/ts2/ts2-sim-server/simulation"
)

// SendJson() sets content type, and encodes the payload and sends to browser
func SendJson(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	var json_data []byte

	json_data, err = json.MarshalIndent(payload, "", "  ")
	if err != nil {
		json_data, err = json.MarshalIndent(NewErrorResponse(err), "", "")
	}
	w.Write(json_data)
}

// Creates new adhock json Payload
func NewAjaxPayload() map[string]interface{} {
	pay := make(map[string]interface{})
	pay["error"] = nil
	return pay
}

/*
H_Ajax() - handles and serves ajax requests for "/ajax".
*/
func H_AjaxIndex(w http.ResponseWriter, r *http.Request) {

	data := NewAjaxPayload()
	data["title"] = sim.Options.Title
	data["description"] = sim.Options.Description
	data["host"] = "ws://" + r.Host + "/ws"
	data["ajax_urls"] = []string{"/trains", "/foo"}

	SendJson(w, data)
}

/*
H_AjaxTrains() - handles and serves ajax requests for "/ajax/trains" - test.
*/
func H_AjaxTrains(w http.ResponseWriter, r *http.Request) {

	data := NewAjaxPayload()
	data["trains"] = sim.Trains

	SendJson(w, data)
}

/*
H_AjaxTrackItems() - handles and serves ajax requests for "/ajax/trackitems" - test.
*/
func H_AjaxTrackItems(w http.ResponseWriter, r *http.Request) {

	items := make([]interface{}, 0)
	for _, i := range sim.TrackItems {
		items = append(items, i)
	}
	data := NewAjaxPayload()
	data["trackitems"] = items

	SendJson(w, data)
}

/*
H_Svg()  experimental to show svg
*/
func H_SvgTest(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	s.Start(1000, 800)


	// Add circle for trackitem.. We Dont have __type__ ;-(
	for _, ti := range sim.TrackItems {


		switch ti.Type(){
		case "LineItem":

			s.Circle(ti.Origin().Xi(), ti.Origin().Yi(), 2, "fill:none;stroke:black")
			s.Line(ti.Origin().Xi(),ti.Origin().Yi(), 20, 20,  "fill:none;stroke:green")

			//s.Line(ti.Origin().Xi(),ti.Origin().Yi(), ti.End().Xi(), 20,  "fill:none;stroke:green")
			// WTF - how to get END!!!
			//m := structs.Map(ti)
			//xe, _ := m["Xf"].(int)
			//ye, _ := m["Yf"].(int)
			//fmt.Println("==",xe,ye)
			//s.Line(ti.Origin().Xi(),ti.Origin().Yi(), xe, ye,  "fill:none;stroke:green")

		default:
			s.Circle(ti.Origin().Xi(), ti.Origin().Yi(), 2, "fill:none;stroke:red")

		}

		s.Circle(ti.Origin().Xi(), ti.Origin().Yi(), 5, "fill:none;stroke:black")
	}

	s.End()
}