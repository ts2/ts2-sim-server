var STA = {}
STA.running = false;
STA.auth = false;
STA.connected = false;

var aspects = {};

function makeSignal(){
    
    var w = 60;
    var h = 150;
    var bw = 5;
    
    var circleRadius = 30;
    var cent = (w / 2) - (circleRadius / 2);

    var Logo = SVG('ts2_signal').size(w, h);
    Logo.rect(w, h).radius(10).attr({ fill: '#999999' })
    Logo.rect(w - bw, h - bw).radius(10).move(bw / 2, bw / 2).attr({ fill: '#222222' })

    var down = 10;
    var space = 4;
    aspects.yelltop = Logo.circle(circleRadius).move(cent, down).attr({ fill: 'yellow' });
    aspects.green = Logo.circle(circleRadius).move(cent, down + circleRadius + space).attr({ fill: '#62D637' });
    aspects.yellbottom = Logo.circle(circleRadius).move(cent, down + (circleRadius * 2) + (space * 2)).attr({ fill: 'yellow' });
    aspects.red = Logo.circle(circleRadius).move(cent, down + (circleRadius * 3) + (space * 3)).attr({ fill: 'red' });
}

function updateSignalState(){
    //console.log(STA);
    var offColor = "#444444";
    aspects.yelltop.attr({fill: STA.connected && !STA.auth && !STA.running ?  "yellow" : offColor});
    aspects.green.attr({fill: STA.connected && STA.auth && STA.running ? "#62D637" : offColor});
    aspects.yellbottom.attr({fill: (STA.connected && !STA.auth && !STA.running) || (STA.connected && STA.auth && !STA.running) ? "yellow" :  offColor});
    aspects.red.attr({fill: !STA.connected ? "red" : offColor});
}


window.addEventListener("load", function (evt) {


    makeSignal();

    // ----
    var input = document.getElementById("input");
    var ws = null;
    var print = function (message) {
        $('#output').append(message + "\n")
    };
    var updateWidgets = function () {
        

        var label = $('#lblConnectedStatus');
        label.text(STA.connected ? "Connected" : "Disconnected");
        label.toggleClass("connected", STA.connected);
        label.toggleClass("not-connected", !STA.connected);

        var label = $('#lblRunningStatus');
        label.text((!STA.connnected && !STA.auth) ? "-------" : STA.running ? "Running" : "Paused");
        if(STA.running){
            label.removeClass("not-running");
            label.addClass("running");
        } else {
            label.removeClass("running");
            label.addClass("not-running");
        }
        

        $('#btnClose').prop("disabled", !STA.connected);
        $('#btnOpen').prop("disabled", STA.connected);
        $('#btnSend').prop("disabled", !STA.connected);
        $('#btnSendClear').prop("disabled", !STA.connected);
        $('#btnLogin').prop("disabled", !STA.connected);
        $('#btnSimStart').prop("disabled", !STA.connected);
        $('#btnSimPause').prop("disabled", !STA.connected);
        updateSignalState();
       
        $("#action_buttons_div button").prop("disabled", !STA.connected);
    };


    function incrementCounter(xid){
        var lbl = $(xid);
        lbl.html(parseInt(lbl.text(), 10) + 1);
    }

    document.getElementById("btnOpen").onclick = function (evt) {
        if (ws) {
            return false;
        }
        //ws = new WebSocket(TS2_WEBSOCKET_HOST);
        ws = new ReconnectingWebSocket(TS2_WEBSOCKET_HOST);
        ws.debug = true;
        ws.timeoutInterval = 5000;
        ws.maxReconnectInterval = 20000;
        ws.maxReconnectAttempts = 30;
        ws.reconnectDecay = 1.5;
        ws.automaticOpen = true;

        ws.onopen = function (evt) {
            STA.connected = true;
            updateWidgets();
            
        };
        ws.onclose = function (evt) {
            STA.connected = false;
            STA.auth = false;
            STA.running = false;
            updateWidgets();
            ws = null;
        };
        ws.onmessage = function (evt) {
            
            try {
                var resp = JSON.parse(evt.data);

                switch(resp.msgType){
                    case "notification":
                        if(resp.data.name == "clock"){
                            
                            var lbl = $("#clock");
                            lbl.html(resp.data.object);
                            return // get outta here

                        } else if (resp.data.name == "stateChanged"){
                            print("= RESPONSE: " + evt.data);
                            // {"msgType":"notification","data":{"name":"stateChanged","object":{"value":false}}}
                            console.log("RUNNING============", resp.data.object.value, resp.data)
                            STA.running = resp.data.object.value
                        }
                        incrementCounter("#lblNoticesCount")

                        break;
                
                    case "response":
                    
                        if(resp.data.status == "OK"){
                            print("= RESPONSE: " + evt.data);
                            //setLogoState(true, true); // workaround
                            incrementCounter("#lblRecvOkCount")
                            //var lbl = $("#lblRecvOkCount");
                            //lbl.html(parseInt(lbl.text(), 10) + 1);

                        } else if(resp.data.status == "FAIL"){
                            print("= RESPONSE: " + evt.data);
                            //var lbl = $("#lblRecvFailCount");
                            //lbl.html(parseInt(lbl.text(), 10) + 1);
                            incrementCounter("#lblRecvFailCount")
                    }    
                    break;
                }
                updateWidgets();
                return false;

            } catch (err) {
                print(err)
                print("< ERROR decoding json: " + evt.data);
                return false; 
            }
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
        var vv = input.value.trim();
        if(vv.length < 10){
            input.focus();
            return
        }
        print("> SENT: " + vv);
        ws.send(vv);
        $('#input').val("");
        input.focus();

        incrementCounter("#lblSentCount");

        return false;
    };
    document.getElementById("btnSendClear").onclick = function (evt) {
        if (!ws) {
            return false;
        }
        document.getElementById("btnClearLog").click();
        document.getElementById("btnSend").click();
    };
    document.getElementById("btnClose").onclick = function (evt) {
        if (ws) {
            ws.close();
        }
        return false;
    };
    document.getElementById("btnClearLog").onclick = function (evt) {
        $('#output').empty();
        return false;
    };
    document.getElementById("btnClearInput").onclick = function (evt) {
        $('#input').val("");
        input.focus();
        return false;
    };
    //#####################################
    document.getElementById("btnLogin").onclick = function (evt) {
        var btnSend = document.getElementById("btnSend")
        document.getElementById("loginTmpl").click();
        btnSend.click();
        document.getElementById("addListenerTmpl").click();
        btnSend.click();
        $("#input").val('{"object": "server", "action": "addListener", "params": {"event": "stateChanged"}}');
        btnSend.click();
        //document.getElementById("simStartTmpl").click();
        //btnSend.click();
        STA.auth = true;
        
    }
    document.getElementById("btnSimStart").onclick = function (evt) {
        document.getElementById("simStartTmpl").click();
        document.getElementById("btnSend").click();
    }
    document.getElementById("btnSimPause").onclick = function (evt) {
        document.getElementById("simPauseTmpl").click();
        document.getElementById("btnSend").click();
    }
    // Templates
    document.getElementById("loginTmpl").onclick = function (evt) {
        input.value = '{"object": "server", "action": "register", "params": {"type": "client", "token": "client-secret"}}';
        input.focus();
        return false;
    };
    document.getElementById("addListenerTmpl").onclick = function (evt) {
        input.value = '{"object": "server", "action": "addListener", "params": {"event": "clock", "ids": []}}';
        input.focus();
        return false;
    };
    document.getElementById("removeListenerTmpl").onclick = function (evt) {
        input.value = '{"object": "server", "action": "removeListener", "params": {"event": "clock"}}';
        input.focus();
        return false;
    };
    document.getElementById("renotifyTmpl").onclick = function (evt) {
        input.value = '{"object": "server", "action": "renotify"}';
        input.focus();
        return false;
    };
    document.getElementById("simStartTmpl").onclick = function (evt) {
        input.value = '{"object": "simulation", "action": "start"}';
        input.focus();
        return false;
    };
    document.getElementById("simPauseTmpl").onclick = function (evt) {
        input.value = '{"object": "simulation", "action": "pause"}';
        input.focus();
        return false;
    };
    document.getElementById("simDumpTmpl").onclick = function (evt) {
        input.value = '{"object": "simulation", "action": "dump"}';
        input.focus();
        return false;
    };
    document.getElementById("optionsListTmpl").onclick = function (evt) {
        input.value = '{"object": "option", "action": "list"}';
        input.focus();
        return false;
    };
    document.getElementById("optionsSetTmpl").onclick = function (evt) {
        input.value = '{"object": "option", "action": "set", "params": {"name": "description", "value": "Demo Simulation"}}';
        input.focus();
        return false;
    };
    document.getElementById("tiListTmpl").onclick = function (evt) {
        input.value = '{"object": "trackItem", "action": "list"}';
        input.focus();
        return false;
    };
    document.getElementById("tiShowTmpl").onclick = function (evt) {
        input.value = '{"object": "trackItem", "action": "show", "params": {"ids": ["23", "24"]}}';
        input.focus();
        return false;
    };
    document.getElementById("plListTmpl").onclick = function (evt) {
        input.value = '{"object": "place", "action": "list"}';
        input.focus();
        return false;
    };
    document.getElementById("plShowTmpl").onclick = function (evt) {
        input.value = '{"object": "place", "action": "show", "params": {"ids": ["LFT", "STN"]}}';
        input.focus();
        return false;
    };
    document.getElementById("ttListTmpl").onclick = function (evt) {
        input.value = '{"object": "trainType", "action": "list"}';
        input.focus();
        return false;
    };
    document.getElementById("ttShowTmpl").onclick = function (evt) {
        input.value = '{"object": "trainType", "action": "show", "params": {"ids": ["UT"]}}';
        input.focus();
        return false;
    };
    document.getElementById("serviceListTmpl").onclick = function (evt) {
        input.value = '{"object": "service", "action": "list"}';
        input.focus();
        return false;
    };
    document.getElementById("serviceShowTmpl").onclick = function (evt) {
        input.value = '{"object": "service", "action": "show", "params": {"ids": ["S001"]}}';
        input.focus();
        return false;
    };
    document.getElementById("routeListTmpl").onclick = function (evt) {
        input.value = '{"object": "route", "action": "list"}';
        input.focus();
        return false;
    };
    document.getElementById("routeShowTmpl").onclick = function (evt) {
        input.value = '{"object": "route", "action": "show", "params": {"ids": ["1", "3"]}}';
        input.focus();
        return false;
    };
    document.getElementById("routeActivateTmpl").onclick = function (evt) {
        input.value = '{"object": "route", "action": "activate", "params": {"id": "1"}}';
        input.focus();
        return false;
    };
    document.getElementById("routeDeactivateTmpl").onclick = function (evt) {
        input.value = '{"object": "route", "action": "deactivate", "params": {"id": "1"}}';
        input.focus();
        return false;
    };
    document.getElementById("trainListTmpl").onclick = function (evt) {
        input.value = '{"object": "train", "action": "list"}';
        input.focus();
        return false;
    };
    document.getElementById("trainShowTmpl").onclick = function (evt) {
        input.value = '{"object": "train", "action": "show", "params": {"ids": [0]}}';
        input.focus();
        return false;
    };
    updateWidgets();
});

