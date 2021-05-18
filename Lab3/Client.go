package main

import (
	"fmt"
	"net"
)

func main() {
	c, err := net.Dial("tcp", "175.24.152.230:8888")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	for {
		var op string
		var s1, s2 string
		fmt.Scan(&op)
		if op == "set" {
			fmt.Scan(&s1, &s2)
			c.Write([]byte(op + " " + s1 + " " + s2))
		} else if op == "get" {
			fmt.Scan(&s1)
			c.Write([]byte(op + " " + s1))
			var data []byte
			_, err := c.Read(data)
			if err != nil {
				fmt.Println("read error", err)
				break
			}
			fmt.Println(string(data))
		}
	}
}
