package main

import (
	"fmt"
	"net"
	"strings"
)

var d map[string]string = make(map[string]string)

func ClientWork(c net.Conn) {
	fmt.Println("receive ", c)
	defer c.Close()
	for {
		data := make([]byte, 1024)
		_, err := c.Read(data)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		s := strings.Split(string(data), " ")
		if s[0] == "set" {
			fmt.Println(s)
			d[s[1]] = s[2]
		} else if s[0] == "get" {
			fmt.Println(s)
			data = []byte(d[s[1]])
			fmt.Println(string(data))
			c.Write(data)
		}
	}
}
func main() {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			return
		}
		go ClientWork(c)
	}
}
