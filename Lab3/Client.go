package main

import (
	"fmt"
	"net"
)

const NotExist = "? Exist?"

func main() {
	c, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	for {
		var op string
		var s1, s2 string
		fmt.Scan(&op)
		info := make([]byte, 1024)
		if op == "set" {
			fmt.Scan(&s1, &s2)
			fmt.Println(op + " " + s1 + " " + s2)
			c.Write([]byte(op + " " + s1 + " " + s2))
			len, err := c.Read(info)
			sts := string(info[:len])
			if sts == "NAK" {
				fmt.Println("操作失败")
			} else if sts == "ACK" {
				fmt.Println("操作成功")
			} else if err != nil {
				fmt.Println("远程连接断开")
			} else {
				fmt.Println("未知错误")
			}
		} else if op == "get" {
			fmt.Scan(&s1)
			c.Write([]byte(op + " " + s1))
			len, err := c.Read(info)
			if err != nil {
				fmt.Println("read error", err)
				break
			}
			s := string(info[:len])
			if s == NotExist {
				fmt.Printf("key [%s] not exist\n", s1)
			} else {
				fmt.Println(s)
			}
		}
	}
}
