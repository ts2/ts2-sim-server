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
	"html/template"
	"net/http"

	"time"

	"os"

	"github.com/ts2/ts2-sim-server/simulation"
	log "gopkg.in/inconshreveable/log15.v2"
)

const (
	DEFAULT_ADDR      string = "0.0.0.0"
	DEFAULT_PORT      string = "22222"
	MaxHubStartupTime        = 3 * time.Second
)

var sim *simulation.Simulation
var hub *Hub
var logger log.Logger

/*
InitializeLogger creates the logger for the server module
*/
func InitializeLogger(parentLogger log.Logger) {
	logger = parentLogger.New("module", "server")
}

/*
Run() starts a http web server and websocket hub for the given simulation, on the given address and port.
*/
func Run(s *simulation.Simulation, addr, port string) {
	sim = s
	hub = &Hub{}
	hubUp := make(chan bool)
	timer := time.After(MaxHubStartupTime)
	go hub.run(hubUp)
	select {
	case <-hubUp:
		HttpdStart(addr, port)
	case <-timer:
		log.Crit("Hub did not start")
		os.Exit(1)
	}
}

/*
StartHttpd() starts the server which serves on the following routes:

    / - Serves a HTTP home page with the server status and information about the loaded sim.
        It also includes a JavaScript WebSocket client to communicate and manage the server.

    /ws - WebSocket endpoint for all TS2 clients and managers.
*/
func HttpdStart(addr, port string) {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	serverAddress := fmt.Sprintf("%s:%s", addr, port)
	logger.Info("Starting HTTP", "submodule", "http", "address", serverAddress)
	err := http.ListenAndServe(serverAddress, nil)
	logger.Crit("HTTP crashed", "submodule", "http", "error", err)
}

/*
serveHome() serves the html home.html page with integrated JS WebSocket client.
*/
func serveHome(w http.ResponseWriter, r *http.Request) {
	logger.Debug("New HTTP connection", "submodule", "http", "remote", r.RemoteAddr)
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
    <link rel="icon" href="https://ts2.github.io/favicon.ico" type="image/x-icon" />
    <link rel="shortcut icon" href="https://ts2.github.io/favicon.ico" type="image/x-icon" />
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" rel="stylesheet" integrity="sha256-7s5uDGW3AHqw6xtJmNNtr+OBRJUlgkNJEo78P4b0yRw= sha512-nNo+yCHEyn0smMxSswnf/OnX6/KwJuZTlNZBjauKhTK0c+zT+q5JOCx0UFhXQ6rJR9jg6Es8gPuD2uZcYDLqSw==" crossorigin="anonymous">
    <script src="https://code.jquery.com/jquery-2.2.0.min.js"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js" integrity="sha256-KXn5puMvxCw+dAYznun+drMdG1IFl3agK0p/pqT9KAo= sha512-2e8qq0ETcfWRI4HJBzQiA3UoyFk6tbNyG+qSaIBZLyW9Xf3sWZHN/lxe9fTh1U45DpPf07yj94KsUHHWe4Yk1A==" crossorigin="anonymous"></script>
    <script>
    	function clearMessage(){
			$('input').val("");
			$('input').focus();
    	}
        window.addEventListener("load", function (evt) {
            var output = document.getElementById("output");
            var input = document.getElementById("input");
            var ws = null;
            var print = function (message) {
                $('#output').append(message + "\n")
            };
            var showConnected = function(connected){
				print(connected ? "# WS Connected": "# WS Disconnected");
				$('#lblStatus').text(connected ? "Connected" : "Disconnected");
				$('#btnClose').prop("disabled", !connected);
				$('#btnOpen').prop("disabled", connected);
				$('#btnSend').prop("disabled", !connected);
            };
            document.getElementById("btnOpen").onclick = function (evt) {
                if (ws) {
                    return false;
                }
                ws = new WebSocket("{{.Host}}");
                ws.onopen = function (evt) {
					showConnected(true);
                };
                ws.onclose = function (evt) {
                    showConnected(false);
                    ws = null;
                };
                ws.onmessage = function (evt) {
                    print("< RESPONSE: " + evt.data);
                };
                ws.onerror = function (evt) {
                    print("< ERROR: " + evt.data);
                };
                input.focus();
                return false;
            };
            document.getElementById("btnSend").onclick = function (evt) {
                if (!ws) {
                    return false;
                }
                print("> SEND: " + input.value);
                ws.send(input.value);
                $('input').focus();
                return false;
            };
            document.getElementById("btnClose").onclick = function (evt) {
                if (ws) {
                    ws.close();
                }
                return false;
            };
            document.getElementById("btnClear").onclick = function (evt) {
            	$('#output').empty();
            	return false;
            };
            showConnected(false);
        });
        // //{"object": "Server", "action": "login", "params": {"type": "client","token": "client-secret"} }
    </script>
</head>
<body>
<nav class="navbar navbar-inverse xx-navbar-fixed-top">
    <div class="container">
        <div class="navbar-header">
            <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar" aria-expanded="false" aria-controls="navbar">
                <span class="sr-only">Toggle navigation</span>
                <span class="icon-bar"></span>
                <span class="icon-bar"></span>
                <span class="icon-bar"></span>
            </button>
            <a class="navbar-brand" href="#">TS2 Sim Server</a>
        </div>
        <div id="navbar" class="collapse navbar-collapse">
            <ul class="nav navbar-nav">
                <li class="active"><a href="/">Home</a></li>
                <li><a href="https://godoc.org/github.com/ts2/ts2-sim-server" target="_godoc">godoc</a></li>
            </ul>
        </div><!--/.nav-collapse -->
    </div>
</nav>

<div class="container">
    <table class="table table-bordered table-condensed">
        <caption>Loaded Simulation</caption>
        <tr>
            <th>Title:</th>
            <td>{{ .Title }}</td>
        </tr>
        <tr>
            <th>Description:</th>
            <td>{{ .Description }}</td>
        </tr>
        <tr>
            <th>WebSocket Server:</th>
            <td>{{ .Host }}</td>
        </tr>
    </table>

    <h3>Test WebSocket</h3>
    <p>
        Click "Open" to create a connection to the server,
        "Send" to send a message to the server and "Close" to close the connection.
        You can change the message and send multiple times.
    </p>
    <form  class="form-inline">
        <div class="form-group">
            <label id="lblStatus" style="width: 100px;">Closed</label>
            <button id="btnOpen"  type="button" class="btn btn-info">Open</button>
            <button id="btnClose"  type="button" class="btn btn-info">Close</button>
        </div>
        <div class="form-group">
            <input type="text" id="input" style="width:500px" placeHolder="Message">
            <span class="glyphicon glyphicon-remove-circle" onclick="clearMessage()"></span>
            <button id="btnSend"  type="button" class="btn btn-success">Send</button>
            <button id="btnClear"  type="button" class="btn btn-default">Clear</button>
        </div>
    </form>
</div>
<div class="container">
    <form>
        <textarea id="output" style="margin-top: 10px; width:100%; height:300px; overflow: auto;"  placeHolder="Log"></textarea>
    </form>
</div>
</body>
</html>
`))
