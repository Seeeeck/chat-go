package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Conn       net.Conn
	flag       int
	Username   string
}

func (c *Client) Run() {
	go c.handleResponse()
	c.handleRequest()
}

func (c *Client) handleResponse() {
	_, err := io.Copy(os.Stdout, c.Conn)
	if err != nil {
		return
	}
}

func (c *Client) handleRequest() {
	for c.flag != 0 {
		for c.menu() != true {
		}
		switch c.flag {
		case 1:
			flag := c.handlePublicChat()
			if flag != 0 {
				return
			}
			break
		case 2:
			fmt.Println("private")
			break
		case 3:
			fmt.Println(">> Input username:")
			c.rename()
			break
		}
	}
	c.release()
}

func (c *Client) rename() {
	var name string
	fmt.Scanln(&name)
	command := "!rename:" + name + "\n"
	_, err := c.Conn.Write([]byte(command))
	if err != nil {
		fmt.Println("Conn.Write err", err)
		return
	} else {
		c.Username = name
	}
}

func (c *Client) handlePublicChat() int {
	var msg string
	fmt.Println(">> Input the message,\"!exit\" to exit")
	fmt.Scanln(&msg)
	for msg != "!exit" {
		if len(msg) != 0 {
			_, err := c.Conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Println("Conn Write err:", err)
				break
			}
		}
		fmt.Scanln(&msg)
	}
	if msg == "!exit" {
		return 0
	} else {
		return -1
	}
}

func (c *Client) release() {
	err := c.Conn.Close()
	if err != nil {
		return
	}
}

func (c *Client) menu() bool {
	var flag int
	fmt.Println(">> Input number to select menu.")
	fmt.Println("1.public")
	fmt.Println("2.private")
	fmt.Println("3.rename")
	fmt.Println("0.exit")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println("Input valid number")
		return false
	}
}

func NewClient(serverIP string, serverPort int) *Client {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	return &Client{serverIP, serverPort, conn, 999, conn.LocalAddr().String()}
}
