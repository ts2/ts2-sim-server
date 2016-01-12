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
	"fmt"
	"github.com/ts2/ts2-sim-server/simulation"
	"html/template"
	"log"
	"net/http"
)

var sim *simulation.Simulation
var hub *Hub

/*
Run starts an http server and a hub for the given simulation, on the given address and port.
*/
func Run(s *simulation.Simulation, addr, port string) {
	sim = s
	hub = &Hub{}
	go HttpdStart(addr, port)
	hub.run()
}

/*
HttpdStart starts the server which serves on the following routes:

/ : Serves a HTTP home page with the server status and information about the loaded sim.
It also includes a JavaScript WebSocket client to communicate and manage the server.

/ws : WebSocket endpoint for all TS2 clients and managers.
*/
func HttpdStart(addr, port string) {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	serverAddress := fmt.Sprintf("%s:%s", addr, port)
	log.Printf("Starting HTTP at: http://%s\n", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}

/*
serveHome() serves the html home.html page with integrated JS WebSocket client.
*/
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Title       string
		Description string
		Host        string
	}{
		sim.Options.Title,
		sim.Options.Description,
		"ws://" + r.Host + "/ws",
	}
	homeTempl.Execute(w, data)
}

var homeTempl = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <script>
        window.addEventListener("load", function (evt) {
            var output = document.getElementById("output");
            var input = document.getElementById("input");
            var ws;
            var print = function (message) {
                var d = document.createElement("div");
                d.innerHTML = message;
                output.appendChild(d);
            };
            document.getElementById("open").onclick = function (evt) {
                if (ws) {
                    return false;
                }
                ws = new WebSocket("{{.Host}}");
                ws.onopen = function (evt) {
                    print("OPEN");
                };
                ws.onclose = function (evt) {
                    print("CLOSE");
                    ws = null;
                };
                ws.onmessage = function (evt) {
                    print("RESPONSE: " + evt.data);
                };
                ws.onerror = function (evt) {
                    print("ERROR: " + evt.data);
                };
                return false;
            };
            document.getElementById("send").onclick = function (evt) {
                if (!ws) {
                    return false;
                }
                print("SEND: " + input.value);
                ws.send(input.value);
                input.value = "";
                return false;
            };
            document.getElementById("close").onclick = function (evt) {
                if (!ws) {
                    return false;
                }
                ws.close();
                return false;
            };
        });
    </script>
</head>
<body>
<h1>TS2 Sim Server</h1>
<p>
    TS2 Sim Server is running !
</p>
<h2>Simulation</h2>
<table>
    <tr>
        <th>Title:</th>
        <td>{{ .Title }}</td>
    </tr>
    <tr>
        <th>Description:</th>
        <td>{{ .Description }}</td>
    </tr>
</table>

<h2>Test WebSocket Connection</h2>
<table>
    <tr>
        <td valign="top" width="50%">
            <p>
                WebSocket server: {{ .Host }}
            </p>
            <p>
                Click "Open" to create a connection to the server,
                "Send" to send a message to the server and "Close" to close the connection.
                You can change the message and send multiple times.
            </p>
            <form>
                <p>
                    <button id="open">Open</button>
                    <button id="close">Close</button>
                </p>
                <p>
                    <textarea id="input" type="text"></textarea>
                    <button id="send">Send</button>
                </p>
            </form>
        </td>
        <td valign="top" width="50%">
            <div id="output"></div>
        </td>
    </tr>
</table>
</body>
</html>
`))
