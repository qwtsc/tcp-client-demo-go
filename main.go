package main

import (
	"bufio"
	"fmt"
	mail "gopkg.in/gomail.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"time"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var conn *bufio.ReadWriter
var mailDialer *mail.Dialer
var alertMessage *mail.Message

func init() {
	config := Config{}
	content, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(content, &config); err != nil {
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
	mailDialer = mail.NewDialer(config.Host, config.Port, config.Username, config.Password)
}

func main() {
	for i := 0; i < 10; i++ {
		singleSchedule()
		go func() {
			if err := mailDialer.DialAndSend(alertMessage); err != nil {
				fmt.Println(err)
			}
		}()
	}
}

func readResp(conn *bufio.ReadWriter) {
	resp, err := conn.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)

}

func sendCommand(command string) {
	fmt.Fprintln(conn, command)
	conn.Flush()
	fmt.Println(command)
	readResp(conn)
}

func singleSchedule() {
	sendCommand("record")
	sendCommand("run 10")
	time.Sleep(time.Second * 50)
	sendCommand("schedule 1800 10 2.93 1")
	time.Sleep(time.Second * 1800)
	readResp(conn) // waiting for receiving the terminal signal of schedule, otherwise, endless loop
	sendCommand("sstop")
	time.Sleep(time.Second * 5)
}
