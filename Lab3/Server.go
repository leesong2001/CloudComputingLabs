package main

import (
	"fmt"
	"net"
	"strings"
)

var d map[string]string = make(map[string]string)

const debug = true
const NotExist = "? Exist?"

func work(c net.Conn) {
	if debug {
		fmt.Println("receive ", c)
	}
	defer c.Close()
	for {
		data := make([]byte, 1024)
		n, err := c.Read(data)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		s := strings.Split(string(data[:n]), " ")
		if s[0] == "set" {
			if debug {
				fmt.Println(s)
			}
			d[s[1]] = s[2]
			c.Write([]byte("ACK"))
		} else if s[0] == "get" {
			info, find := d[s[1]]
			if !find {
				if debug {
					fmt.Println("NOT FOUND!")
				}
				c.Write([]byte(NotExist))
			} else {
				data = []byte(info)
				if debug {
					fmt.Println(s)
					fmt.Println(string(data))
				}
				c.Write(data)
			}
		}
	}
}
func main() {
	l, err := net.Listen("tcp", ":11111")
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
		go work(c)
	}
}
