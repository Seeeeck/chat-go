package server

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

type User struct {
	Id         string
	Username   string
	Chan       chan string
	Conn       net.Conn
	server     *Server
	CountAlive chan bool
	Alive      bool
}

func InitUser(conn net.Conn, server *Server) *User {
	user := &User{conn.RemoteAddr().String(),
		conn.RemoteAddr().String(),
		make(chan string), conn, server, make(chan bool, 5), true}
	go user.listenMessageFromChannel()
	go user.listenAlive()
	return user
}

func (u *User) listenMessageFromChannel() {
	for msg := range u.Chan {
		_, err := u.Conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("Write to conn err:", err)
			u.Alive = false
			return
		}
	}
	u.Alive = false
}

func (u *User) SendMsgToClient(msg string) {
	_, err := u.Conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("SendMsgToClient err:", err)
	}
}

func (u *User) HandleMsg(msg string) {
	msgStatus := u.isCommand(msg)
	switch msgStatus {
	case 0:
		u.server.broadCast(u, msg)
		break
	case 1:
		returnMsg := u.server.generateUserListMsg()
		u.SendMsgToClient(returnMsg)
		break
	case 2:
		username := msg[8:]
		existed := false
		u.server.lockWrapper(func() {
			for _, u := range u.server.UserMap {
				if u.Username == username {
					existed = true
					return
				}
			}
			u.Username = username
		})
		if existed {
			u.SendMsgToClient(">> The username is already exist.\n")
		} else {
			u.SendMsgToClient(">> Success!\n")
		}
		break
	case 3:
		username, message := u.parseData(msg)
		u.server.mapLock.RLock()
		user, ok := u.server.UserMap[username]
		u.server.mapLock.RUnlock()
		if ok {
			user.SendMsgToClient("[" + u.Username + "](private):" + message + "\n")
		} else {
			u.SendMsgToClient(">> User [" + username + "] is not exist\n")
		}
		break
	}
}

func (u *User) isCommand(msg string) int {
	if msg == "!who" {
		return 1
	} else if strings.HasPrefix(msg, "!rename:") && len(strings.TrimSpace(msg)) > 8 {
		return 2
	} else if u.isPrivateMsg(msg) {
		return 3
	} else {
		return 0
	}
}
func (u User) isPrivateMsg(msg string) bool {
	re := regexp.MustCompile(`^!to:(?P<name>.+?):(?P<msg>.+)`)
	if re == nil {
		fmt.Println("re err")
		return false
	}
	return re.MatchString(msg)
}
func (u User) parseData(msg string) (name, message string) {
	re := regexp.MustCompile(`^!to:(?P<name>.+?):(?P<message>.+)`)
	if re == nil {
		fmt.Println("re err")
	}
	subMatch := re.FindStringSubmatch(msg)
	return subMatch[1], subMatch[2]
}

func (u *User) listenAlive() {
	for !u.Alive {
		u.server.lockWrapper(func() {
			delete(u.server.UserMap, u.Id)
		})
		u.server.broadCast(u, "下线了.")
		close(u.Chan)
		close(u.CountAlive)
		err := u.Conn.Close()
		if err != nil {
			fmt.Println("user Conn Close err:", err)
		}
		return
	}
}
