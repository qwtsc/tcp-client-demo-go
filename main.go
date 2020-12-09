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

type Motor struct {
	Conn *bufio.ReadWriter
}

type Sampler struct {
	Conn *bufio.ReadWriter
}

type Valve struct {
	Conn *bufio.ReadWriter
}

type Setup struct {
	Pump
	Sensor
    Motor
    Valve
    Sampler
	MailDialer *mail.Dialer
}

func NewSetup(pump Pump, sensor Sensor, motor Motor, valve Valve, sampler Sampler, dialer *mail.Dialer) Setup{
	return Setup{
		pump, sensor, motor, valve, sampler, dialer,
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
	command := fmt.Sprintf("run %.2f", rate)
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

func (valve *Valve) openV(pos int) {
	command := fmt.Sprintf("relay %d", pos)
	if err := sendCommand(valve.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (valve *Valve) closeV(pos int) {
	command := fmt.Sprintf("rstop %d", pos)
	if err := sendCommand(valve.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (motor *Motor) clock(step int) {
	command := fmt.Sprintf("motor %d", step)
	if err := sendCommand(motor.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (motor *Motor) antiClock(step int) {
	command := fmt.Sprintf("antimotor %d", step)
	if err := sendCommand(motor.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (motor *Motor) mstop() {
	if err := sendCommand(motor.Conn, "stopmotor"); err != nil {
		fmt.Println("run error", err)
	}
}

func (sampler *Sampler) turnX(dist int) {
	command := fmt.Sprintf("turnx %d", dist)
	if err := sendCommand(sampler.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (sampler *Sampler) turnXY(dx int, dy int) {
	command := fmt.Sprintf("turnxy %d %d", dx, dy)
	if err := sendCommand(sampler.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (sampler *Sampler) turnY(dist int) {
	command := fmt.Sprintf("turny %d", dist)
	if err := sendCommand(sampler.Conn, command); err != nil {
		fmt.Println("run error", err)
	}
}

func (sampler *Sampler) turn0() {
	if err := sendCommand(sampler.Conn, "turn0"); err != nil {
		fmt.Println("run error", err)
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

func (setup *Setup) testAbsorbanceRun() {
    setup.closeAllValves()
    for vPos := 0; vPos < 5; vPos++ {
        setup.Valve.openV(vPos)
        setup.record()
        setup.Pump.run(10)
	    time.Sleep(time.Second * 180)
        fmt.Println("please check whether the light intensity power input is 90%")
       // for i := 0; i < 3; i++ {
       //     setup.Motor.antiClock(1000)
       //     fmt.Println("1")
       // }
       // setup.Motor.clock(3000)
        setup.stop()
        setup.sstop()
        setup.Valve.closeV(vPos)
    }
}

func (setup *Setup) testSamplerMoving() {
    setup.turn0()
    for x:=-101; x<0; x-=18 {
        for y:=83; y>0; y-=18 {
            setup.turnXY(x, y)
	        time.Sleep(time.Second * 2)
            setup.turnXY(0, 0)
	        time.Sleep(time.Second * 2)
        }
    }
}

func (setup *Setup) closeAllValves() {
    for vPos := 0; vPos < 5; vPos++ {
        setup.Valve.closeV(vPos)
    }
}

func (setup *Setup) autoRun(secs int) {
    setup.closeAllValves()
    for vPos := 0; vPos < 5; vPos++ {
        setup.Valve.openV(vPos)
        fmt.Println("please check whether the light intensity power input is 90%")
        setup.singleRun(secs)
        for i := 0; i < 3; i++ {
            setup.Motor.antiClock(1000)
            time.Sleep(60*time.Second)
            setup.singleRun(secs)
        }
        setup.Motor.clock(3000)
        time.Sleep(60*time.Second)
        setup.Valve.closeV(vPos)
    }
}

func (setup *Setup) singleRun(secs int) {
	setup.record()
	setup.run(10)
	time.Sleep(time.Second * 120)
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
	tcpConnM, err := net.Dial("tcp", "192.168.2.203:8813")
	if err != nil {
		panic(err)
	}
	conn := bufio.NewReadWriter(bufio.NewReader(tcpConn), bufio.NewWriter(tcpConn))
	connM := bufio.NewReadWriter(bufio.NewReader(tcpConnM), bufio.NewWriter(tcpConnM))
	readResp(conn)
    readResp(connM)
	mailDialer := mail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	setup = NewSetup(Pump{conn}, Sensor{conn}, Motor{connM}, Valve{conn}, Sampler{connM},mailDialer)
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
        // setup.autoRun(1800)
        setup.testSamplerMoving()
        done<- struct{}{}
    }()
//	go func() {
//		for i := 0; i < 4; i++ {
//			setup.singleRun(1800)
//			go func() {
//				if err := setup.MailDialer.DialAndSend(alertMessage); err != nil {
//					fmt.Println(err)
//				}
//			}()
//            go func() {
//                soundRemainder("该改变光强了")
//                time.Sleep(time.Second*10)
//                soundRemainder("该改变光强了")
//                time.Sleep(time.Second*10)
//                soundRemainder("提醒沈冲该改变光强了")
//            }()
//		}
//		done<- struct{}{}
//	}()
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, syscall.SIGINT)
	select{
	case <-abort:
		setup.interrupt()
		setup.stop()
		setup.sstop()
		setup.mstop()
		os.Exit(1)
	case <-done:
		fmt.Println("Program exits without error")
		os.Exit(0)
	}
}
