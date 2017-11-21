package main

import (
	"fmt"
	"net"
	"bufio"
	"time"
)

func callServer(msg, sid, pid *byte, tab *[]byte) {
	conn, _ := net.Dial("tcp", "localhost:8080")
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	buff := make([]byte, 12)
	setMsg(buff, *msg)
	setSessId(buff, *sid)
	setPlayerId(buff, *pid)
	setTab(buff, *tab)
	w.Write(buff)
	w.Flush()
	r.Read(buff)
	*msg = getMsg(buff)
	*sid = getSessId(buff)
	*pid = getPlayerId(buff)
	*tab = getTab(buff)
}

func main() {
	tab := make([]byte, 9)
	var msg, sid, pid, winner byte
	var p byte
	msg = NEW
	firstTurn := true
	for winner == 0 {
		callServer(&msg, &sid, &pid, &tab)
		winner = findWinner(tab)
		switch msg {
		case WAIT:
			time.Sleep(time.Second)
			msg = UPDATE
		case TURN:
			if firstTurn {
				firstTurn = false
				p = chooseOpositeToken(tab)
				if p == 0 {
					p = pickToken()
				}
			}
			fmt.Println("Tablero actual: ")
			drawTab(tab)
			if winner == 0 {
				scanJugada(tab, p)
				drawTab(tab)
				msg = PLAY
			}
		}
	}
	fmt.Printf("Ganaron las %c!", winner)
}