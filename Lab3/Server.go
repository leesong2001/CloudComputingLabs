package main

import (
	"fmt"
	"net"
	"strings"
)

var d map[string]string

func ClientWork(c net.Conn) {
	fmt.Println("receive ", c)
	defer c.Close()
	for {
		var data []byte
		_, err := c.Read(data)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		s := strings.Split(string(data), " ")
		fmt.Println(s)
		if s[0] == "set" {
			d[s[1]] = s[2]
		} else if s[0] == "get" {
			c.Write([]byte(d[s[1]]))
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
