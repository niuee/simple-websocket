package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	Connection   *websocket.Conn
	Type         string
	InputChannel *chan string
	Name         string
	Id           int
	OutputList   []*websocket.Conn
}

type Room struct {
	Name      string
	UAVConns  map[string]*Conn
	UserConns map[string]*Conn
}

type TempRoom struct {
	Name  string
	Conns map[string]*Conn
}

type MavMsg struct {
	Destination string
	MsgBody     string
}

func (c *Conn) Compare(b *Conn) int {
	return 1
}

type WSRequestMsg struct {
	Requester string `json:"requester"`
}

type ticketResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var user = map[string]int{
	"user":  1,
	"user2": 2,
}

var connmap = map[string]*Conn{}

var Rooms = map[string]TempRoom{}

func addConnection(name string) *Conn {
	if conn, ok := connmap[name]; ok {
		return conn
	} else {
		return nil
	}
}

func wshandle(rw http.ResponseWriter, r *http.Request) {

	//fmt.Println("All connections")
	var conn *Conn = nil
	connectionid := r.Header["Connectionid"]
	connectionName := r.Header["Connectionname"]
	connectionIdInt, err := strconv.Atoi(connectionid[0])
	if err != nil {
		log.Println(err)
		return
	}

	conn = addConnection(connectionName[0])
	if conn == nil {
		connection, err := wsupgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		conn = &Conn{
			Connection: connection,
			Name:       connectionName[0],
			Id:         connectionIdInt,
		}
	}
	if connectionIdInt%2 == 0 {
		//add to first room
		Rooms["Second Room"].Conns[conn.Name] = conn
	} else {
		Rooms["First Room"].Conns[conn.Name] = conn
	}

	go readFromConnection(conn)
	fmt.Println("Exiting WS connection establishment handler function")
}

func readFromConnection(c *Conn) {
	for {
		mt, message, err := c.Connection.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("Received Message: %s. Message Type: %d\n from %s", message, mt, c.Name)
		log.Println(time.Now().String())

	}

	defer func() {
		c.Connection.Close()
		delete(connmap, c.Name)
		if c.Id%2 == 0 {
			delete(Rooms["Second Room"].Conns, c.Name)
			fmt.Println("Removed from Room 2")
		} else {
			delete(Rooms["First Room"].Conns, c.Name)
			fmt.Println("Removed from Room 1")
		}
	}()
	fmt.Printf("Connection %s is disconnecting\n", c.Name)
	for connectionName, connection := range Rooms["Second Room"].Conns {
		fmt.Println(connectionName)
		fmt.Println(connection.Type)
	}
}

func notifyWSConnection(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Println(r.URL.Scheme)
	jsonDecoder := json.NewDecoder(r.Body)
	message := WSRequestMsg{}
	err := jsonDecoder.Decode(&message)
	if err != nil {
		log.Println(err)
		return
	}
	response := ticketResponse{}
	if user, ok := user[message.Requester]; ok {
		fmt.Printf("User: %d is permitted\n", user)
		response.Status = "success"
		response.Info = "User is allowed to connect"
	} else {
		fmt.Printf("User is not permitted\n")
		response.Status = "failed"
		response.Info = "User is not allowed to connect"
	}

	jsonByte, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		return
	}
	rw.Write(jsonByte)
}

func main() {
	http.HandleFunc("/wsticket", notifyWSConnection)
	http.HandleFunc("/websocket2/ws", wshandle)

	Rooms["First Room"] = TempRoom{
		Name:  "First Room",
		Conns: make(map[string]*Conn),
	}
	Rooms["Second Room"] = TempRoom{
		Name:  "Second Room",
		Conns: make(map[string]*Conn),
	}
	http.ListenAndServe(":9999", nil)
}
