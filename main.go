package main

import (
	"bufio"
	"fmt"
	mail "gopkg.in/gomail.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Sensor struct {
	Conn *bufio.ReadWriter
}

type Pump struct {
	Conn *bufio.ReadWriter
}

type Setup struct {
	Pump
	Sensor
	MailDialer *mail.Dialer
}

func NewSetup(pump Pump, sensor Sensor, dialer *mail.Dialer) Setup{
	return Setup{
		pump, sensor, dialer,
	}
}

func readResp(conn *bufio.ReadWriter) {
	resp, err := conn.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(resp)
}

func sendCommand(conn *bufio.ReadWriter, command string) error {
	command = command + "\r\n"
	n, err := conn.WriteString(command)
	if n != len(command) {
		return fmt.Errorf("expected to write %d bytes, but wrote %d", len(command), n)
	}
	if err != nil {
		return err
	}
	if err = conn.Flush(); err!=nil {
		return err
	}
	fmt.Println(command)
	readResp(conn)
	return nil
}

func (pump *Pump) schedule(time int, rateInitial float32, volumn float32, slope float32)  {
	command := fmt.Sprintf("schedule %d %.1f %.2f %.2f", time, rateInitial, volumn, slope)
	if err := sendCommand(pump.Conn, command); err != nil {
		fmt.Println("schedule error", err)
	}
}

func (pump *Pump) run(rate float32)  {
	command := fmt.Sprintf("run %.1f", rate)
	if err := sendCommand(pump.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (pump *Pump) stop()  {
	if err := sendCommand(pump.Conn, "stop"); err != nil {
		fmt.Println("stop error", err)
	}
}
func (pump *Pump) interrupt()  {
	if err := sendCommand(pump.Conn, "interrupt"); err != nil {
		fmt.Println("interrupt error", err)
	}
}
func (sensor *Sensor) record() {
	if err := sendCommand(sensor.Conn, "record"); err != nil {
		fmt.Println("record error", err)
	}
}

func (sensor *Sensor) sstop() {
	if err := sendCommand(sensor.Conn, "sstop"); err != nil {
		fmt.Println("sstop error", err)
	}
}

func (setup *Setup) singleRun(secs int) {
	setup.record()
	setup.run(10)
	time.Sleep(time.Second * 60)
	setup.schedule(secs, 10, 2.93, 1)
	time.Sleep(time.Second * time.Duration(secs))
	readResp(setup.Pump.Conn)
	setup.sstop()
	time.Sleep(time.Second * 5)
}

var setup Setup

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
	conn := bufio.NewReadWriter(bufio.NewReader(tcpConn), bufio.NewWriter(tcpConn))
	readResp(conn)
	mailDialer := mail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	setup = NewSetup(Pump{conn}, Sensor{conn}, mailDialer)
}

func soundRemainder(text string) {
    conn, err := net.Dial("tcp", "192.168.2.155:8814")
    defer conn.Close()
    if err != nil {
        fmt.Println(err)
    }
    conn.Write([]byte(text))
}

func main() {
	alertMessage := mail.NewMessage()
	alertMessage.SetHeader("From", "shenchongdadi@163.com")
	alertMessage.SetHeader("To", "shenchongdadi@163.com")
	alertMessage.SetHeader("Subject", "You need to change the light intensity")
	alertMessage.SetBody("text/html", "<b>chong</b> please stand up and do you exp!")
	done := make(chan struct{}, 1)
	go func() {
		for i := 0; i < 5; i++ {
			setup.singleRun(1200)
			go func() {
				if err := setup.MailDialer.DialAndSend(alertMessage); err != nil {
					fmt.Println(err)
				}
			}()
            go func() {
                soundRemainder("该改变光强了")
                time.Sleep(time.Second*10)
                soundRemainder("该改变光强了")
                time.Sleep(time.Second*10)
                soundRemainder("提醒沈冲该改变光强了")
            }()
		}
		done<- struct{}{}
	}()
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, syscall.SIGINT)
	select{
	case <-abort:
		setup.interrupt()
		setup.stop()
		setup.sstop()
		os.Exit(1)
	case <-done:
		fmt.Println("Program exits without error")
		os.Exit(0)
	}
}
