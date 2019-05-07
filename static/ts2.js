



//= Client State
var STA = {};
STA.connected = false; // ws connected
STA.auth = false; // auth against server
STA.running = false; // sim running

//====================================================
// Signal Logo
var aspects = {};

function makeSignal(){
    
    var img_width = 60;
    var img_height = 270; // img height
    var CX = (img_width / 2)

    // create svg and make size (position is in css)
    var isvg = SVG('ts2_signal').size(img_width, img_height);

    // fill with debug color.....
    //isvg.rect(img_width, img_height).attr({ fill: 'green' })

    //==== make vertical pole in middle
    var poleWidth = 16;
    isvg.rect(poleWidth, img_height).radius(5).cx(CX).attr({ fill: '#555555' })

    //== Signal
    var sig_top = 30;
    var sig_h = 160; // signal height = later derived ??
    var sig_w = img_width; 
    var sig_b = 5; // border width
    
    // signal by fill background with light color as rect for border
    isvg.rect(sig_w, sig_h).radius(10).cx(CX).y(sig_top).attr({ fill: '#999999' })
    // fill the signal background color above with a darker layer above
    isvg.rect(sig_w - sig_b, sig_h - (sig_b*2)).radius(10).cx(CX).y(sig_top + sig_b).attr({ fill: '#222222' })
    
    // add aspects
    var lampRadius = 30;
    var aspDown = sig_top + sig_b + 10; // vertical down start point
    var aspSpace = 4; // space between
    aspects.yelltop = isvg.circle(lampRadius).cx(CX).y(aspDown).attr({ fill: 'yellow' });
    aspects.green = isvg.circle(lampRadius).cx(CX).y(aspDown + lampRadius + aspSpace).attr({ fill: '#62D637' });
    aspects.yellbottom = isvg.circle(lampRadius).cx(CX).y(aspDown + (lampRadius * 2) + (aspSpace * 2)).attr({ fill: 'yellow' });
    aspects.red = isvg.circle(lampRadius).cx(CX).y(aspDown + (lampRadius * 3) + (aspSpace * 3)).attr({ fill: 'red' });

    // plate ts2
    var pdown = aspects.red.y() + 50;
    var plate_width = img_width - 10;
    isvg.rect(plate_width, 30).radius(5).cx(CX).y(pdown).attr({ fill: '#efefef' });
    isvg.text("TS2").font({family: 'monospace', size: 18}).cx(CX).y(pdown + 5).attr({ fill: '#111111' })

    // Sim server plates
    pdown = aspects.red.y() + 85;
    //var ts2_server_plate_top = aspects.red.y() + 50;
    isvg.rect(plate_width, 30).radius(5).cx(CX).y(pdown + 2).attr({ fill: '#9B5410' });
    isvg.text("SIM").font({family: 'monospace', size: 11}).cx(CX).y(pdown + 5).attr({ fill: '#dddddd' });
    isvg.text("SERVER").font({family: 'monospace', size: 11}).cx(CX).y(pdown + 18).attr({ fill: '#dddddd' })

}
function updateSignalState(){
    //console.log(STA);
    var offColor = "#444444";
    aspects.yelltop.animate(300).attr({fill: STA.connected && !STA.auth && !STA.running ?  "yellow" : offColor});
    aspects.green.animate(300).attr({fill: STA.connected && STA.auth && STA.running ? "#62D637" : offColor});
    aspects.yellbottom.animate(300).attr({fill: (STA.connected && !STA.auth && !STA.running) || (STA.connected && STA.auth && !STA.running) ? "yellow" :  offColor});
    aspects.red.animate(300).attr({fill: !STA.connected ? "red" : offColor});
}

//=============================
// Updates all stuff dpending upon state STA
function updateWidgets() {
        
    updateSignalState();

    // WS stuff
    var label = $('#lblConnectedStatus');
    label.text(STA.connected ? "Connected" : "Disconnected");
    label.toggleClass("connected", STA.connected);
    label.toggleClass("not-connected", !STA.connected);

    $('#btnClose').prop("disabled", !STA.connected);
    $('#btnOpen').prop("disabled", STA.connected);

    // Sim Stuff
    var label = $('#lblRunningStatus');
    label.text((!STA.connnected && !STA.auth) ? "-------" : STA.running ? "Running" : "Paused");
    if(STA.running){
        label.removeClass("not-running");
        label.addClass("running");
    } else {
        label.removeClass("running");
        label.addClass("not-running");
    }
    
    $('#btnSend').prop("disabled", !STA.connected);
    $('#btnSendClear').prop("disabled", !STA.connected);
    $('#btnLogin').prop("disabled", !STA.connected);
    $('#btnSimStart').prop("disabled", !STA.connected && !STA.auth);
    $('#btnSimPause').prop("disabled", !STA.connected && !STA.auth);
    
    $("#action_buttons_div button").prop("disabled", !STA.connected);

    // var clockWidget = $('#ts2_clock');
    // console.log("lock=", clockWidget, STA)
    // if(STA.running){
    //     clockWidget.removeClass("clock-not-running");
    //     clockWidget.addClass("clock-running");
    // } else {
    //     clockWidget.removeClass("clock-running");
    //     clockWidget.addClass("clock-not-running");
    // }
};

function loadDataTable(data){

    var cols = [];
    var rows = [];
    var keys = Object.keys(data);
    console.log(keys)

    if(keys.length == 0){
        cols = [{ title: "-" }, { title: "-" },]
        rows.push(["No", "Rows"]);

    } else {

        var colNames = Object.keys(data[keys[0]]);
        for(var c=0; c < colNames.length; c++){
            cols.push({title: colNames[c]})
        }
        for(var ki=0; ki < keys.length; ki++){
            var drow = data[keys[ki]];
            console.log("drow=", drow);
            var row = [];
            for(var c=0; c < colNames.length; c++){
                row.push(drow[colNames[c]]);
            }
            console.log("row=", row)   
            rows.push(row);
        }
    }
     
    console.log("cols==", cols)
    console.log("rows==", rows)
    var table = $('#data_table');

    try {
        //table.destroy();
        console.log("destroyes");
    } catch(err) {
        console.log("err=", err);
    }

    table =  $('#data_table').DataTable({
        destroy: true,
        paging: false, searching: false,
        data: rows,
        columns: cols
    });


}


var ws = null;
var sentIDs = {};

function SendListAction(target){
    //{"object": "trainType", "action": "list"}
    var ts  =  new Date().getTime();
    var data = {id: ts, object: target, action: "list"}
    
    sentIDs[ts] = data;
    console.log(ts, data, sentIDs)
    var jso = JSON.stringify(data);
    print("> SENT: " + jso);
    ws.send(jso);
    //$('#input').val("");
    //input.focus();
    incrementCounter("#lblSentCount");
}
function print(message) {
    $('#output').append(message + "\n")
};
function printNotice(message) {
    $('#outputNotifications').append(message )
};
function incrementCounter(xid){
    var lbl = $(xid);
    lbl.html(parseInt(lbl.text(), 10) + 1);
}



//= Lets go !

var clockWidget;

window.addEventListener("load", function (evt) {

    makeSignal();
    clockWidget = $('#ts2_clock');
    
    var input = document.getElementById("input");
    


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
                incrementCounter("#lblTotalCount");

                switch(resp.msgType){

                    case "notification":
                    
                        incrementCounter("#lblNoticesCount")
                        printNotice("= NOTICE: " + evt.data);

                        // Clock
                        if(resp.data.name == "clock"){
                            
                            clockWidget.html(resp.data.object);
                            
                            return // get outta here

                        // Sim running or paused
                        } else if (resp.data.name == "stateChanged"){ 
                            
                            STA.running = resp.data.object.value
                            console.log("stateChanges", STA.running)
                            if(STA.running){
                                clockWidget.removeClass("clock-paused");
                                clockWidget.addClass("clock-running");
                            } else {
                                clockWidget.removeClass("clock-running");
                                clockWidget.addClass("clock-paused");
                            }
                        }
                        break;
                
                    case "response":
                        print("= RESPONSE: " + evt.data);

                        
                        if( resp.id > 0){
                            var sent = sentIDs[resp.id];
                            console.log("-----------------------------\nsent=", sent);
                            console.log("resp=", resp)
                            if(sent.action == "list"){
                                console.log("YES");
                                loadDataTable(resp.data);

                            }
                        }


                        if(resp.data.status == "OK"){
                            incrementCounter("#lblRecvOkCount");
                            
                        } else if(resp.data.status == "FAIL"){
                            incrementCounter("#lblRecvFailCount")
                        }    
                        break;
                }
                updateWidgets();
                return false;

            } catch (err) {
                print("< ERROR: " + evt.data);
                print(err)
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
        SendListAction("trackItem");
        //input.value = '{"object": "trackItem", "action": "list"}';
        //input.focus();
        return false;
    };
    document.getElementById("tiShowTmpl").onclick = function (evt) {
        input.value = '{"object": "trackItem", "action": "show", "params": {"ids": ["23", "24"]}}';
        input.focus();
        return false;
    };
    document.getElementById("plListTmpl").onclick = function (evt) {
        SendListAction("place");
        //input.value = '{"object": "place", "action": "list"}';
        //input.focus();
        return false;
    };
    document.getElementById("plShowTmpl").onclick = function (evt) {
        input.value = '{"object": "place", "action": "show", "params": {"ids": ["LFT", "STN"]}}';
        input.focus();
        return false;
    };
    document.getElementById("ttListTmpl").onclick = function (evt) {
        SendListAction("trainType");
        //input.value = '{"object": "trainType", "action": "list"}';
        //input.focus();
        return false;
    };
    document.getElementById("ttShowTmpl").onclick = function (evt) {
        input.value = '{"object": "trainType", "action": "show", "params": {"ids": ["UT"]}}';
        input.focus();
        return false;
    };
    document.getElementById("serviceListTmpl").onclick = function (evt) {
        SendListAction("service");
        //input.value = '{"object": "service", "action": "list"}';
        //input.focus();
        return false;
    };
    document.getElementById("serviceShowTmpl").onclick = function (evt) {
        input.value = '{"object": "service", "action": "show", "params": {"ids": ["S001"]}}';
        input.focus();
        return false;
    };
    document.getElementById("routeListTmpl").onclick = function (evt) {
        SendListAction("route");
        //input.value = '{"object": "route", "action": "list"}';
        //input.focus();
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
    do_resize();

    $('#data-tab').tab('show');
    loadDataTable({});


});


// resize output to bottom of page
function do_resize(){
    var ele = document.getElementById("output");
    var rect = ele.getBoundingClientRect();
    var height = window.innerHeight;
    var inpHeight = height - rect.y - 10;
    if(inpHeight < 300) {
        inpHeight  = 300
    }
    var hstyle = inpHeight + "px";
    ele.style.height = hstyle
    document.getElementById("outputNotifications").style.height = hstyle;
    document.getElementById("data_table").style.height = hstyle;
}

window.addEventListener("resize", function (evt) {
    do_resize();
});