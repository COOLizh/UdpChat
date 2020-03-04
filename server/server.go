package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/COOLizh/itirod/UdpChat/common"
	"github.com/sirupsen/logrus"
)

//Server ...
type Server struct {
	listener        *net.UDPConn
	users           map[string]string // key - username, value - remote address
	sendMessage     chan common.ServerResponse
	recievedMessage chan common.Message
	handleMessage   chan common.Message
	//TODO: dialogues, groups, general room
}

func (s *Server) RecieveClientMessage() {
	for {
		buff := make([]byte, 1024)
		bytes, _, err := s.listener.ReadFromUDP(buff)
		common.HandleError(err, common.ErrorFatal)
		var msg common.Message
		err = json.Unmarshal(buff[:bytes-1], &msg)
		common.HandleError(err, common.ErrorFatal)
		s.handleMessage <- msg
	}
}

func (s *Server) SendClientMessage() {
	for {
		msgGen := <-s.sendMessage
		msg := msgGen.Message
		msgJSON, err := json.Marshal(msg)
		common.HandleError(err, common.ErrorFatal)
		msgJSON = append(msgJSON, '\n')
		for _, v := range msgGen.Addrs {
			addr, err := net.ResolveUDPAddr("udp4", v)
			common.HandleError(err, common.ErrorFatal)
			s.listener.WriteToUDP(msgJSON, addr)
		}
	}
}

//HandleClientRequest ...
func (s *Server) HandleClientRequest() {
	for {
		msg := <-s.handleMessage
		switch msg.MessageHeader.MessageType {
		case common.GeneralRoom:
		case common.DialogueRoom:
		case common.GroupRoom:
		case common.Instruction:
			switch msg.MessageHeader.Function {
			case common.LogIn:
				var content string
				_, ok := s.users[msg.Author]
				if !ok {
					content = "You have been succesfully registered"
				} else {
					content = "Such user exist"
				}
				s.users[msg.Author] = msg.MessageHeader.RemoteAddr
				fmt.Println(s.users)
				addrs := make([]string, 0)
				addrs = append(addrs, msg.MessageHeader.RemoteAddr)
				var responce = common.ServerResponse{
					Message: common.Message{
						Content: content,
					},
					Addrs: addrs,
				}
				s.sendMessage <- responce
			case common.CreateDialogue:
			case common.CreateGroup:
			default:
			}
		default:
		}
	}
}

func main() {
	var server = Server{
		listener:        nil,
		users:           make(map[string]string),
		sendMessage:     make(chan common.ServerResponse),
		recievedMessage: make(chan common.Message),
		handleMessage:   make(chan common.Message),
	}
	var err error
	udpAddr := &net.UDPAddr{
		Port: 8000,
		IP:   net.ParseIP("127.0.0.1"),
	}
	server.listener, err = net.ListenUDP("udp", udpAddr)
	common.HandleError(err, common.ErrorFatal)
	defer func(s *Server) {
		fail := recover()
		if fail != nil {
			logrus.Error(fail)
		}
		s.listener.Close()
	}(&server)

	go server.HandleClientRequest()

	go server.RecieveClientMessage()

	server.SendClientMessage()
}
