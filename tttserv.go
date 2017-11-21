package main

import (
	"fmt"
	"net"
	"bufio"
)

type Sess struct {
	turn byte
	play bool
	tab []byte
}

var currentSessId byte
var sessions map[byte]*Sess

func sesion(conn net.Conn, chFirstPlayer chan bool) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	buff := make([]byte, 12)

	r.Read(buff)
	msg := getMsg(buff)
	sid := getSessId(buff)
	pid := getPlayerId(buff)
	tab := getTab(buff)
	if msg == NEW {
		if <- chFirstPlayer {
			sessions[currentSessId] = &Sess{}
			setSessId(buff, currentSessId)
			setPlayerId(buff, 0)
			setMsg(buff, WAIT)
			fmt.Printf("Jugador 1 sesión %d conectado.\n", currentSessId)
			chFirstPlayer <- false
		} else {
			setSessId(buff, currentSessId)
			setPlayerId(buff, 1)
			setMsg(buff, TURN)
			sessions[currentSessId].turn = 1
			fmt.Printf("Jugador 2 sesión %d conectado.\n", currentSessId)
			currentSessId++
			chFirstPlayer <- true
		}
	} else if msg == PLAY {
		sessions[sid].tab = tab
		sessions[sid].play = true
		setMsg(buff, WAIT)
	} else if msg == UPDATE {
		if sessions[sid].turn != pid && sessions[sid].play {
			setTab(buff, sessions[sid].tab)
			setMsg(buff, TURN)
			sessions[sid].play = false
			sessions[sid].turn = (sessions[sid].turn + 1) % 2
		} else {
			setMsg(buff, WAIT)
		}
	}
	w.Write(buff)
	w.Flush()
}

func main() {
	chFirstPlayer := make(chan bool, 1)
	chFirstPlayer <- true
	sessions = make(map[byte]*Sess)
	lstnr, _ := net.Listen("tcp", "localhost:8080")
	defer lstnr.Close()
	for {
		conn, _ := lstnr.Accept()
		go sesion(conn, chFirstPlayer)
	}
}
