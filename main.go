package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"time"
	mail "gopkg.in/gomail.v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	host string
	port int
	username string
	password string
}

var conn *bufio.ReadWriter
var mailDailer *mail.Dialer
var alertMessage *mail.Message

func init() {
	config := Config{}
	content, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(content, config); err != nil {
		panic(err)
	}
	tcpConn, err := net.Dial("tcp", "192.168.2.155:8813")
	if err != nil {
		panic(err)
	}
	conn = bufio.NewReadWriter(bufio.NewReader(tcpConn), bufio.NewWriter(tcpConn))
	readResp(conn)
	alertMessage = mail.NewMessage()
	alertMessage.SetHeader("From", "shenchongdadi@163.com")
	alertMessage.SetHeader("To", "shenchongdadi@163.com")
	alertMessage.SetHeader("Subject", "You need to change the light intensity")
	alertMessage.SetBody("text/html", "<b>chong</b> please stand up and do you exp!")
	mailDailer = mail.NewDialer(config.host, config.port, config.username, config.password)
}

func main() {
	for i:=0; i<10; i++ {
		singleSchedule()
		go func() {
			if err := mailDailer.DialAndSend(alertMessage); err != nil{
				fmt.Println(err)
			}}()
	}
}

func readResp(conn *bufio.ReadWriter){
	fmt.Println(conn.ReadString('\n'))
}

func sendCommand(command string){
	fmt.Fprintln(conn, command)
	fmt.Println(conn.ReadString('\n'))
}

func singleSchedule() {
	sendCommand("record")
	sendCommand("run 10")
	time.Sleep(time.Second*50)
	sendCommand("schedule 1800 10 2.93 1")
	time.Sleep(time.Second*1800)
	readResp(conn) // waiting for receiving the terminal signal of schedule, otherwise, endless loop
	sendCommand("sstop")
	time.Sleep(time.Second*5)
}