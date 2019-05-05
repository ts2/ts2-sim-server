
//= Client State
var STA = {};
STA.connected = false; // ws connected
STA.auth = false; // auth against server
STA.running = false; // sim running

//====================================================
// Signal Logo
var aspects = {};

function makeSignal(){
    
    var w = 60;
    var ih = 200; // img height
    var sh = 150; // signal height
    var bw = 5; // border width
    
    var circleRadius = 30;
    var vCenter = (w / 2) - (circleRadius / 2);

    // create svg and make size (position is in css)
    var sigi = SVG('ts2_signal').size(w, ih);

    sigi.rect(w, ih).attr({ fill: 'green' })

    // make vertical pole in middle
    //var poleSize = 16;
    //sigi.rect(poleSize, h).move(vCenter - (poleSize/2), 0).attr({ fill: 'red' })

    // fill background with light color as rect for border
    sigi.rect(w, sh).radius(10).attr({ fill: '#999999' })
    // fill the signal background color above with a darker layer above
    sigi.rect(w - bw, sh - bw).radius(10).move(bw / 2, bw / 2).attr({ fill: '#222222' })

    // add aspects
    var down = 10; // vertical down start point
    var space = 4; // space between
    aspects.yelltop = sigi.circle(circleRadius).move(vCenter, down).attr({ fill: 'yellow' });
    aspects.green = sigi.circle(circleRadius).move(vCenter, down + circleRadius + space).attr({ fill: '#62D637' });
    aspects.yellbottom = sigi.circle(circleRadius).move(vCenter, down + (circleRadius * 2) + (space * 2)).attr({ fill: 'yellow' });
    aspects.red = sigi.circle(circleRadius).move(vCenter, down + (circleRadius * 3) + (space * 3)).attr({ fill: 'red' });

    // plates
    var plate_width = w -   20;
    sigi.circle(circleRadius).move(vCenter, down + (circleRadius * 4) + (space * 4)).attr({ fill: 'blue' });
    sigi.rect(plate_width, 30).radius(5).move(vCenter - (plate_width), down + (circleRadius * 4) + (space * 6)).attr({ fill: '#cccccc' })

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
};

function loadDataTable(dict){

    //for

    $('#data_table').DataTable({
        data: [[1,2,3]],
        columns: [
            { title: "Name" },
            { title: "Name" },
            { title: "Name" },
        ]
    });

    // )
    // $(document).ready(function() {
    //     $('#example').DataTable( {
    //         data: dataSet,
    //         columns: [
    //             { title: "Name" },
    //             { title: "Position" },
    //             { title: "Office" },
    //             { title: "Extn." },
    //             { title: "Start date" },
    //             { title: "Salary" }
    //         ]
    //     } );
    // } );
}


//= Lets go !
window.addEventListener("load", function (evt) {

    makeSignal();

    
    var input = document.getElementById("input");
    var ws = null;
    var print = function (message) {
        $('#output').append(message + "\n")
    };
    var printNotice = function (message) {
        $('#outputNotifications').append(message + "\n")
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
                    
                        incrementCounter("#lblNoticesCount")
                        printNotice("= NOTICE: " + evt.data);

                        // Clock
                        if(resp.data.name == "clock"){
                            
                            var lbl = $("#clock");
                            lbl.html(resp.data.object);
                            return // get outta here

                        // Sim running or paused
                        } else if (resp.data.name == "stateChanged"){ 
                            
                            STA.running = resp.data.object.value
                        }
                        break;
                
                    case "response":
                        print("= RESPONSE: " + evt.data);

                        if(resp.data.status == "OK"){
                            incrementCounter("#lblRecvOkCount")

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
    do_resize();

    //$('#data-view').tab('show')
    //loadDataTable();


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
    document.getElementById("data-table").style.height = hstyle;
}

window.addEventListener("resize", function (evt) {
    do_resize();
});