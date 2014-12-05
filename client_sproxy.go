package main


import (
	"fmt"
	"net"
	"time"
	"bufio"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:2233")
	if err != nil {
		fmt.Println("error")
		return
	}
	str := ` {"Cmd": "date", "Ip": ["192.168.0.1", "192.168.0.2"]}
	`
	fmt.Fprintf(conn, str)
	time.Sleep(time.Second * 1)
	status, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Printf(status)
}

