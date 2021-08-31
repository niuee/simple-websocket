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
	Id           int
	OutputList   []*websocket.Conn
}

type Room struct {
	Name      string
	UAVConns  []*Conn
	UserConns []*Conn
}

func (c *Conn) Compare(b *Conn) int {
	return 1
}

type WSRequestMsg struct {
	Requester string `json:"requester"`
}

type UAVIdentifier struct {
	Uavid int   `json:"UAVID"`
	User  []int `json:"User"`
}

type UavUser struct {
	Uuuid   int    `json:"uuuid"`
	Name    string `json:"name"`
	UavList []int  `json:"uavlist"`
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

var connmap = map[int]*Conn{}

var conweb = map[int][]int{
	1: {0, 3, 5},
}

var UAVList = map[int]UAVIdentifier{
	5: {
		Uavid: 5,
		User:  []int{1},
	},
}

func addConnection(id int) *Conn {
	if conn, ok := connmap[id]; ok {
		return conn
	} else {
		return nil
	}
}

func wshandle(rw http.ResponseWriter, r *http.Request) {

	//fmt.Println("All connections")
	fmt.Println(connmap)
	var conn *Conn = nil
	var intVar int
	connectionid := r.Header["Connectionid"]
	intVar, err := strconv.Atoi(connectionid[0])
	if err != nil {
		log.Println(err)
		return
	}
	conn = addConnection(intVar)
	if conn == nil {
		fmt.Println("Jia la")
		connection, err := wsupgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(intVar)
		conn = &Conn{
			Connection: connection,
			Id:         intVar,
		}
		fmt.Println(conn)
	}

	go readFromConnection(conn)
}

func readFromConnection(c *Conn) {
	for {
		mt, message, err := c.Connection.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("Received Message: %s. Message Type: %d\n", message, mt)
		log.Println(time.Now().String())
	}

	defer c.Connection.Close()
	defer delete(connmap, c.Id)
	fmt.Printf("Connection %d is disconnecting\n", c.Id)

}

func notifyWSConnection(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
	http.HandleFunc("/ws", wshandle)

	http.ListenAndServe(":9999", nil)

}
