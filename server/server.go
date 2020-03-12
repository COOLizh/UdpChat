package main

import (
	"encoding/json"
	"net"

	"github.com/COOLizh/itirod/UdpChat/common"
	"github.com/sirupsen/logrus"
)

//Server...
type Server struct {
	listener        *net.UDPConn
	users           map[string]*common.User
	dialogues       map[int]*common.Conf
	groups          map[int]*common.Conf
	general         []common.Message
	sendMessage     chan common.ServerResponse
	recievedMessage chan common.Message
	handleMessage   chan common.Message
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
			s.general = append(s.general, msg)
			addrs := make([]string, 0, len(s.users))
			for _, v := range s.users {
				addrs = append(addrs, v.Addr)
			}
			var response = common.ServerResponse{
				Message: common.Message{
					MessageHeader: common.MessageHeader{
						MessageType:    common.GeneralRoom,
						ResponseStatus: common.Ok,
						RemoteAddr:     msg.MessageHeader.RemoteAddr,
					},
					Author:  msg.Author,
					Content: msg.Content,
				},
				Addrs: addrs,
			}
			s.sendMessage <- response
		case common.DialogueRoom:
			id := msg.MessageHeader.DestinationID
			s.dialogues[id].Messages = append(s.dialogues[id].Messages, msg)
			addrs := make([]string, 0, len(s.users))
			for _, v := range s.dialogues[id].Users {
				addrs = append(addrs, v.Addr)
			}
			var response = common.ServerResponse{
				Message: common.Message{
					MessageHeader: common.MessageHeader{
						MessageType:    common.DialogueRoom,
						DestinationID:  id,
						ResponseStatus: common.Ok,
						RemoteAddr:     msg.MessageHeader.RemoteAddr,
					},
					Author:  msg.Author,
					Content: msg.Content,
				},
				Addrs: addrs,
			}
			s.sendMessage <- response
		case common.GroupRoom:
			id := msg.MessageHeader.DestinationID
			s.groups[id].Messages = append(s.groups[id].Messages, msg)
			addrs := make([]string, 0, len(s.users))
			for _, v := range s.groups[id].Users {
				addrs = append(addrs, v.Addr)
			}
			var response = common.ServerResponse{
				Message: common.Message{
					MessageHeader: common.MessageHeader{
						MessageType:    common.GroupRoom,
						DestinationID:  id,
						ResponseStatus: common.Ok,
						RemoteAddr:     msg.MessageHeader.RemoteAddr,
					},
					Author:  msg.Author,
					Content: msg.Content,
				},
				Addrs: addrs,
			}
			s.sendMessage <- response
		case common.Instruction:
			switch msg.MessageHeader.Function {
			case common.LogIn:
				var content string
				var status common.ResponseStatus
				_, ok := s.users[msg.Author]
				if !ok {
					s.users[msg.Author] = &common.User{
						Username: msg.Author,
						Addr:     msg.MessageHeader.RemoteAddr,
					}
					logrus.Info("a new user has been registered")
					status = common.Ok
					content = "\nYou have been succesfully registered!\n" + common.CommandsInfo + common.InputArrows
				} else {
					status = common.Fail
					content = "\nSuch user exist\n"
					logrus.Error("can not register new user")
				}
				addrs := make([]string, 0)
				addrs = append(addrs, msg.MessageHeader.RemoteAddr)
				var responce = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType:    msg.MessageHeader.MessageType,
							Function:       msg.MessageHeader.Function,
							ResponseStatus: status,
						},
						Author:  msg.Author,
						Content: content,
					},
					Addrs: addrs,
				}
				s.sendMessage <- responce
			case common.CreateDialogue:
				users := make(map[string]*common.User)
				for _, v := range msg.MessageHeader.RequestCreateConf.UserNames {
					users[v] = s.users[v]
				}
				newID := len(s.dialogues) + 1
				s.dialogues[newID] = &common.Conf{
					Name:     msg.MessageHeader.RequestCreateConf.Name,
					Messages: make([]common.Message, 0),
					Users:    users,
				}
				addrs := make([]string, 0)
				for _, v := range users {
					addrs = append(addrs, v.Addr)
				}
				statusStr := "Dialogue created successfully, dialogue name : " + s.dialogues[newID].Name
				var response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType:    msg.MessageHeader.MessageType,
							Function:       msg.MessageHeader.Function,
							ResponseStatus: common.Ok,
							RemoteAddr:     msg.MessageHeader.RemoteAddr,
							ResponseCreateConf: common.ResponseCreateConf{
								Name: msg.MessageHeader.RequestCreateConf.Name,
								ID:   newID,
							},
						},
						Content: statusStr,
						Author:  msg.Author,
					},
					Addrs: addrs,
				}
				s.sendMessage <- response
			case common.CreateGroup:
				users := make(map[string]*common.User)
				for _, v := range msg.MessageHeader.RequestCreateConf.UserNames {
					users[v] = s.users[v]
				}
				newID := len(s.groups) + 1
				s.groups[newID] = &common.Conf{
					Name:     msg.MessageHeader.RequestCreateConf.Name,
					Messages: make([]common.Message, 0),
					Users:    users,
				}
				addrs := make([]string, 0)
				for _, v := range users {
					addrs = append(addrs, v.Addr)
				}
				statusStr := "Group created successfully, group name : " + s.groups[newID].Name
				var response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType:    msg.MessageHeader.MessageType,
							Function:       msg.MessageHeader.Function,
							ResponseStatus: common.Ok,
							RemoteAddr:     msg.MessageHeader.RemoteAddr,
							ResponseCreateConf: common.ResponseCreateConf{
								Name: msg.MessageHeader.RequestCreateConf.Name,
								ID:   newID,
							},
						},
						Content: statusStr,
						Author:  msg.Author,
					},
					Addrs: addrs,
				}
				s.sendMessage <- response
			}
		}
	}
}

func main() {
	config := common.GetConfig()
	var server = Server{
		listener:        nil,
		users:           make(map[string]*common.User),
		dialogues:       make(map[int]*common.Conf),
		groups:          make(map[int]*common.Conf),
		general:         make([]common.Message, 0),
		sendMessage:     make(chan common.ServerResponse),
		recievedMessage: make(chan common.Message),
		handleMessage:   make(chan common.Message),
	}
	var err error
	addr, err := net.ResolveUDPAddr(config.Network, config.BindAddr)
	common.HandleError(err, common.ErrorFatal)
	server.listener, err = net.ListenUDP(config.Network, addr)
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
