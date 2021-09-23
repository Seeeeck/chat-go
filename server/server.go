package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IPAddr   string
	Port     int
	UserMap  map[string]*User
	Messages chan string
	mapLock  sync.RWMutex
}

func NewServer(IPAddr string, port int) *Server {
	userMap := make(map[string]*User, 10)
	return &Server{IPAddr, port, userMap, make(chan string, 5), sync.RWMutex{}}
}

func (s *Server) handleConn(conn net.Conn) {
	user := InitUser(conn, s)
	s.lockWrapper(func() {
		s.UserMap[user.Id] = user
	})
	s.broadCast(user, "上线.")
	go s.listenMessageFromClient(user)
	s.handleUserAlive(user)
}

func (s *Server) handleUserAlive(user *User) {
	defer func() { user.Alive = false }()
	for {
		select {
		case <-user.CountAlive:

		case <-time.After(60 * time.Second):
			//不活跃踢除群聊
			return
		}
	}
}

func (s *Server) listenMessageFromClient(user *User) {
	defer func() { user.Alive = false }()
	buf := make([]byte, 4096)
	for {
		n, err := user.Conn.Read(buf)
		if n == 0 {
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err:", err)
			return
		}
		msg := string(buf[:n-1])
		user.HandleMsg(msg)
		user.CountAlive <- true
	}
}

func (s *Server) generateUserListMsg() string {
	msg := ""
	s.mapLock.RLock()
	for _, user := range s.UserMap {
		msg += user.Username + ", "
	}
	s.mapLock.RUnlock()
	if msg != "" {
		msg = msg[:len(msg)-2]
	}
	return msg + "\n"
}

func (s *Server) lockWrapper(func1 func()) {
	s.mapLock.Lock()
	func1()
	s.mapLock.Unlock()
}

func (s *Server) broadCast(user *User, msg string) {
	sendMsg := "[" + user.Username + "]: " + msg
	s.Messages <- sendMsg
}

func (s *Server) listenMessages() {
	for msg := range s.Messages {
		s.lockWrapper(func() {
			for _, user := range s.UserMap {
				user.Chan <- msg
			}
		})
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IPAddr, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener.Close err:", err)
		}
	}(listener)
	go s.listenMessages()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		go s.handleConn(conn)
	}

}
