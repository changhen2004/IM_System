package main

import "net"
import "strings"


type User struct{
	Name string
	Addr string
	Conn net.Conn	
	C chan string
	server *Server
}
//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User{
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		C: make(chan string),
		server: server,
	}
	// 启动监听消息的goroutine
	go user.ListenMessage()
	return user
}
//用户上线业务
func(u *User) OnLine(){
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Addr] = u	
	u.server.mapLock.Unlock()
	u.server.Broadcast(u, "用户" + u.Addr + "已上线")
}
//用户下线业务
func(u *User) OffLine(){
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Addr)
	u.server.mapLock.Unlock()
	u.server.Broadcast(u, "用户" + u.Addr + "已下线")
}
//用户处理消息业务
func(u *User) DoMessage(msg string){
	//查询所有在线用户业务
	if msg == "who"{
		u.server.mapLock.Lock()
		for _,user := range u.server.OnlineMap{
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			u.C <- onlineMsg
		}
		u.server.mapLock.Unlock()
	//重命名业务
	}else if len(msg) > 7 && msg[:7] == "rename"{
		//消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		_ , ok := u.server.OnlineMap[newName]
		if ok{
			u.C <- "该用户名已存在\n"
			return
		}else{
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()
			u.Name = newName
			u.C <- "用户名已更新为：" + newName + "\n"
			return
		}
	}else if len(msg) > 4 && msg[:4] == "to|" {
		//消息格式：to|张三|消息内容
		recvUser := strings.Split(msg, "|")[1]
		sendMsg := strings.Split(msg, "|")[2]
		// 检查接收用户是否存在
		recvUserObj, ok := u.server.OnlineMap[recvUser]
		if !ok {
			u.C <- "接收用户不存在\n"
			return
		}
		// 发送消息给接收用户
		recvUserObj.C <- sendMsg
		u.C <- "发送成功\n"
		return
	}else{
		u.server.Broadcast(u, msg)
	}
}


//监听当前User channel的方法，一旦有消息，就接收并发送给客户端
func(u *User) ListenMessage(){
	for{
		msg := <-u.C
		u.Conn.Write([]byte(msg))
	}
}
