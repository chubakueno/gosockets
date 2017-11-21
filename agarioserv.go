// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:9002", "http service address")

var upgrader = websocket.Upgrader{} // use default options
var index = 0
var connections = make([]*websocket.Conn, 0)
var posx = make([]float32, 1000)
var posy = make([]float32, 1000)
var killedby = make([]int, 1000)
func home(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	connections=append(connections,c)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	c.WriteMessage(websocket.TextMessage,[]byte(fmt.Sprintf("%d C", index)));
	var myid = index
	index++
	for {
		_, message, err := c.ReadMessage() //_=mt
		if err != nil {
			log.Println("read:", err)
			break
		}
		var id int
		var op string
		fmt.Sscanf(string(message),"%d%s",&id,&op)
		if (op == "P") {
			var x float32
			var y float32
			fmt.Sscanf(string(message),"%d%s%f%f",&id,&op,&x,&y)
			fmt.Printf("\n%d:%s:%f:%f\n",id,op,x,y);
			posx[id] = x;
			posy[id] = y;
		}
		if (op == "I") {
			var killed int
			fmt.Sscanf(string(message),"%d%s%d",&killed)
			killedby[killed] = id;
		}
		log.Printf("recv: %s", message)
		for i:=0; i<len(connections); i++ {
			if (i==myid) {
				continue
			}
			fmt.Printf("Notifying %d\n",i);
			err = connections[i].WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error:", err)
				break
			}
		}
	}
	upgrader.Upgrade(w, r, nil)
	//homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
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
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))