window.addEventListener("load", function (evt) {

    var lw = 40;
    var lh = 70;
    var bw = 3;
   
    var cr = 18;
    var cent = 12;

    var Logo = SVG('ts2_logo').size(lw, lh);
    Logo.rect(lw, lh).radius(10).attr({ fill: '#666666' })
    Logo.rect(lw - bw, lh - bw).radius(10).move(bw / 2, bw / 2).attr({ fill: '#222222' })

    var down = 5;
    var aspects = {};
    aspects.green = Logo.circle(cr).move(cent, down).attr({ fill: 'green' });
    aspects.amber = Logo.circle(cr).move(cent, down + cr + 2).attr({ fill: 'orange' });
    aspects.red = Logo.circle(cr).move(cent, down + (cr * 2) + 4).attr({ fill: 'red' });

    // ----
    var Auth = false;
    var input = document.getElementById("input");
    var ws = null;
    var print = function (message) {
        $('#output').append(message + "\n")
    };
    var showConnected = function (connected) {
        print(connected ? "# WS Connected" : "# WS Disconnected");
        var label = $('#lblStatus');
        label.text(connected ? "Connected" : "Disconnected");
        label.toggleClass("badge-success", connected);
        label.toggleClass("badge-danger", !connected);
        $('#btnClose').prop("disabled", !connected);
        $('#btnOpen').prop("disabled", connected);
        $('#btnSend').prop("disabled", !connected);

        
        var offColor = "#333333";
        aspects.green.attr({fill: !connected ? offColor : Auth ? "green" : offColor});
        aspects.amber.attr({fill: connected && !Auth ? "orange" :  offColor});
        aspects.red.attr({fill: !connected ? "red" : offColor});


    };
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
        $('#input').val("");
        input.focus();
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
    document.getElementById("btnClearInput").onclick = function (evt) {
        $('#input').val("");
        input.focus();
        return false;
    };
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

