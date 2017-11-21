package main

import (
	"flag"
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/websocket"
)

const maxConn = 1000
var addr = flag.String("addr", "localhost:9002", "http service address")

var upgrader = websocket.Upgrader{} // use default options
var index = 0
var connections = make([]*websocket.Conn, 0)
var posx = make([]float32, maxConn)
var posy = make([]float32, maxConn)
var killedby = make([]int, maxConn)
var dead = make([]bool, maxConn)
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
	for i := 0; i < maxConn; i++ {
		if (posx[i]!=0||posy[i]!=0) {
			c.WriteMessage(websocket.TextMessage,[]byte(fmt.Sprintf("%d P %f %f", i,posx[i],posy[i])));
		}
		if (killedby[i]>=0) {
			c.WriteMessage(websocket.TextMessage,[]byte(fmt.Sprintf("%d I %d", killedby[i], i)));
		}
	}
	for {
		_, message, err := c.ReadMessage() //_=message type
		if err != nil {
			log.Println("read:", err)
			break
		}
		var id int
		var op string
		fmt.Printf("%s\n",string(message));
		fmt.Sscanf(string(message),"%d%s",&id,&op)
		if (op == "P") {
			var x float32
			var y float32
			fmt.Sscanf(string(message),"%d%s%f%f",&id,&op,&x,&y)
			posx[id] = x;
			posy[id] = y;
		}
		if (op == "I") {
			var killed int
			fmt.Sscanf(string(message),"%d%s%d",&killed)
			killedby[killed] = id;
		}
		for i:=0; i<len(connections); i++ {
			if (i==myid||dead[i]) {
				continue
			}
			err = connections[i].WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error:", err)
				break
			}
		}
	}
	dead[myid] = true
}

func main() {
	for i := 0; i < maxConn; i++ {
		killedby[i] = -1
    }
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}