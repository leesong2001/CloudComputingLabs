package main

import (
    "flag"
    "fmt"
)

func main() {
    ip := flag.String("ip", "127.0.0.1", "What is your ip_address?")
    port := flag.Int("port", 8888, "What is the port?")
    number_thread := flag.Int("number_thread", 1, "How much is the thread number?")



    flag.Parse() //解析输入的参数

    fmt.Println("ip:", *ip)
    fmt.Println("port:", *port)
    fmt.Println("number-thread:", *number_thread)

}

