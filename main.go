package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var conn *bufio.ReadWriter

func init() {
	tcpConn, err := net.Dial("tcp", "192.168.2.155:8813")
	if err != nil {
		panic(err)
	}
	conn = bufio.NewReadWriter(bufio.NewReader(tcpConn), bufio.NewWriter(tcpConn))
	readResp(conn)
}

func main() {
	sendCommand("record")
	readResp(conn)
	time.Sleep(time.Second*100)
	sendCommand("sstop")
	readResp(conn)
}

func readResp(conn *bufio.ReadWriter){
	fmt.Println(conn.ReadString('\n'))
}

func sendCommand(command string){
	fmt.Fprintln(conn, "record")
}