//var Logo;
var aspects = {};
function makeLogo(){
    
    var w = 60;
    var h = 120;
    var bw = 5;
    
    var circleRadius = 30;
    var cent = (w / 2) - (circleRadius / 2);

    var Logo = SVG('ts2_logo').size(w, h);
    Logo.rect(w, h).radius(10).attr({ fill: '#999999' })
    Logo.rect(w - bw, h - bw).radius(10).move(bw / 2, bw / 2).attr({ fill: '#222222' })

    var down = 10;
    var space = 4;
    aspects.green = Logo.circle(circleRadius).move(cent, down).attr({ fill: 'green' });
    aspects.amber = Logo.circle(circleRadius).move(cent, down + circleRadius + space).attr({ fill: 'orange' });
    aspects.red = Logo.circle(circleRadius).move(cent, down + (circleRadius * 2) + space + space).attr({ fill: 'red' });
}

function setLogoState(connected, auth){

    var offColor = "#444444";
    aspects.green.attr({fill: !connected ? offColor : auth ? "#62D637" : offColor});
    aspects.amber.attr({fill: connected && !auth ? "orange" :  offColor});
    aspects.red.attr({fill: !connected ? "red" : offColor});

}


window.addEventListener("load", function (evt) {


    makeLogo();


    // ----
    var input = document.getElementById("input");
    var ws = null;
    var print = function (message) {
        $('#output').append(message + "\n")
    };
    var showConnected = function (connected) {
        print(connected ? "# WS Connected" : "# WS Disconnected");
        var label = $('#lblConnectedStatus');
        label.text(connected ? "Connected" : "Disconnected");
        label.toggleClass("connected", connected);
        label.toggleClass("not-connected", !connected);
        $('#btnClose').prop("disabled", !connected);
        $('#btnOpen').prop("disabled", connected);
        $('#btnSend').prop("disabled", !connected);
        $('#btnSendClear').prop("disabled", !connected);
        $('#btnSimStartLogin').prop("disabled", !connected);
        $('#btnSimStart').prop("disabled", !connected);
        $('#btnSimPause').prop("disabled", !connected);
        setLogoState(connected, false);
       
        $("#action_buttons_div button").prop("disabled", !connected);

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
            showConnected(true);
        };
        ws.onclose = function (evt) {
            showConnected(false);
            ws = null;
        };
        ws.onmessage = function (evt) {
            print("= RESPONSE: " + evt.data);
            try {
                var resp = JSON.parse(evt.data);

                switch(resp.msgType){
                    case "notification":
                        if(resp.data.name == "clock"){
                            setLogoState(true, true); // workaround
                            var lbl = $("#clock");
                            lbl.html(resp.data.object);
                            incrementCounter("#lblRecvOkCount")
                        }   
                        break;
                
                    case "response":
                    
                        if(resp.data.status == "OK"){
                            setLogoState(true, true); // workaround
                            incrementCounter("#lblRecvOkCount")
                            //var lbl = $("#lblRecvOkCount");
                            //lbl.html(parseInt(lbl.text(), 10) + 1);

                        } else if(resp.data.status == "FAIL"){
                            //var lbl = $("#lblRecvFailCount");
                            //lbl.html(parseInt(lbl.text(), 10) + 1);
                            incrementCounter("#lblRecvFailCount")
                    }    
                    break;
                }
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
    document.getElementById("btnSimStartLogin").onclick = function (evt) {
        var btnSend = document.getElementById("btnSend")
        document.getElementById("loginTmpl").click();
        btnSend.click();
        document.getElementById("addListenerTmpl").click();
        btnSend.click();
        //$("#input").val('{"object": "server", "action": "addListener", "params": {"event": "stateChanged"}}');
        //btnSend.click();
        document.getElementById("simStartTmpl").click();
        btnSend.click();
        
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
    showConnected(false);
});

