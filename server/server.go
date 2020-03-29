package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/COOLizh/itirod/UdpChat/common"
	"github.com/sirupsen/logrus"
)

// Server : contains all needful information about server
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

//RecieveClientMessage : receives a message for further processing
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

// SendClientMessage : send message to client after processing
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

//HandleClientRequest : processing message
func (s *Server) HandleClientRequest() {
	for {
		msg := <-s.handleMessage
		var response common.ServerResponse
		switch msg.MessageHeader.MessageType {
		case common.GeneralRoom:
			s.general = append(s.general, msg)
			addrs := make([]string, 0, len(s.users))
			for _, v := range s.users {
				addrs = append(addrs, v.Addr)
			}
			response = common.ServerResponse{
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
		case common.DialogueRoom:
			id := msg.MessageHeader.DestinationID
			s.dialogues[id].Messages = append(s.dialogues[id].Messages, msg)
			if len(s.dialogues[id].Users) == 1 {
				break
			}
			var receiver *common.User
			var addrs string
			for _, v := range s.dialogues[id].Users {
				if v.Addr != msg.MessageHeader.RemoteAddr {
					receiver = v
					addrs = receiver.Addr
				}
			}

			if !receiver.IsOnline {
				s.dialogues[id].LastMessageAddr = msg.MessageHeader.RemoteAddr
			} else {
				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType:    common.DialogueRoom,
							DestinationID:  id,
							ResponseStatus: common.Ok,
							RemoteAddr:     msg.MessageHeader.RemoteAddr,
						},
						Author:  msg.Author,
						Content: msg.Author + " : " + msg.Content + "\n",
					},
					Addrs: []string{addrs},
				}

				s.sendMessage <- response

				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType:    common.DialogueRoom,
							DestinationID:  id,
							ResponseStatus: common.Ok,
							RemoteAddr:     msg.MessageHeader.RemoteAddr,
						},
						Author:  msg.Author,
						Content: "*Message recieved*\n",
					},
					Addrs: []string{msg.MessageHeader.RemoteAddr},
				}
			}

		case common.GroupRoom:
			id := msg.MessageHeader.DestinationID
			s.groups[id].Messages = append(s.groups[id].Messages, msg)
			addrs := make([]string, 0)
			for _, v := range s.groups[id].Users {
				if v.Addr == msg.MessageHeader.RemoteAddr {
					continue
				}
				if v.IsOnline {
					addrs = append(addrs, v.Addr)
				}
			}

			// if nobody read message
			if len(addrs) == 0 {
				s.groups[id].LastMessageAddr = msg.MessageHeader.RemoteAddr
			} else {
				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType:    common.GroupRoom,
							DestinationID:  id,
							ResponseStatus: common.Ok,
							RemoteAddr:     msg.MessageHeader.RemoteAddr,
						},
						Author:  msg.Author,
						Content: "*Message recieved*\n",
					},
					Addrs: []string{msg.MessageHeader.RemoteAddr},
				}
				s.sendMessage <- response

				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType:    common.GroupRoom,
							DestinationID:  id,
							ResponseStatus: common.Ok,
							RemoteAddr:     msg.MessageHeader.RemoteAddr,
						},
						Author:  msg.Author,
						Content: msg.Author + " : " + msg.Content + "\n",
					},
					Addrs: addrs,
				}
			}
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
					logrus.Info("a new user was registered")
					status = common.Ok
					content = "\nYou have been succesfully registered!\n" + common.CommandsInfo + common.InputArrows
				} else {
					status = common.Fail
					content = "\nSuch user exist\n"
					logrus.Error("was not registered new user")
				}
				addrs := make([]string, 0)
				addrs = append(addrs, msg.MessageHeader.RemoteAddr)
				response = common.ServerResponse{
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
				statusStr := "Dialogue created successfully, dialogue name : " + s.dialogues[newID].Name + common.InputArrows
				response = common.ServerResponse{
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
				fmt.Println(s.dialogues[1])
			case common.CreateGroup:
				users := make(map[string]*common.User)
				for _, v := range msg.MessageHeader.RequestCreateConf.UserNames {
					users[v] = s.users[v]
				}
				newID := len(s.groups) + 1
				s.groups[newID] = &common.Conf{
					Name:            msg.MessageHeader.RequestCreateConf.Name,
					Messages:        make([]common.Message, 0),
					Users:           users,
					LastMessageAddr: "",
				}
				addrs := make([]string, 0)
				for _, v := range users {
					addrs = append(addrs, v.Addr)
				}
				statusStr := "Group created successfully, group name : " + s.groups[newID].Name + common.InputArrows
				response = common.ServerResponse{
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
				logrus.Info("a new group was created")

			case common.ConnectGroup:
				//template of response
				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType: msg.MessageHeader.MessageType,
							Function:    msg.MessageHeader.Function,
						},
						Author: msg.Author,
					},
					Addrs: []string{msg.MessageHeader.RemoteAddr},
				}

				//check if such group exists
				var isNotExist bool
				var confKey int
				for key := range s.groups {
					if s.groups[key].Name == msg.Content {
						isNotExist = true
						confKey = key
						break
					}
				}
				if !isNotExist {
					response.Message.Content = "No group with such name" + common.InputArrows
					response.Message.MessageHeader.ResponseStatus = common.Fail
					break
				}

				//check user member of conf
				_, ok := s.groups[confKey].Users[msg.Author]
				if !ok {
					response.Message.Content = "You're not member of this group" + common.InputArrows
					response.Message.MessageHeader.ResponseStatus = common.Fail
					break
				}

				//make this user online in chat
				s.groups[confKey].Users[msg.Author].IsOnline = true

				//collect all messages from group
				var content string = "*You are in " + s.groups[confKey].Name + " group*\n"
				for _, message := range s.groups[confKey].Messages {
					content += message.Author + ": " + message.Content + "\n"
				}

				//if there unread messages
				if s.groups[confKey].LastMessageAddr != "" {
					var res = common.ServerResponse{
						Message: common.Message{
							MessageHeader: common.MessageHeader{
								MessageType:    common.GroupRoom,
								DestinationID:  confKey,
								ResponseStatus: common.Ok,
								RemoteAddr:     msg.MessageHeader.RemoteAddr,
							},
							Author:  msg.Author,
							Content: "*Message recieved*\n",
						},
						Addrs: []string{s.groups[confKey].LastMessageAddr},
					}
					s.sendMessage <- res
					s.groups[confKey].LastMessageAddr = ""
				}

				response.Message.Content = content
				response.Message.MessageHeader.DestinationID = confKey
				response.Message.MessageHeader.ResponseStatus = common.Ok
				response.Message.MessageHeader.ResponseCreateConf.Name = s.groups[confKey].Name

			case common.ConnectDialogue:
				//template of response
				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType: msg.MessageHeader.MessageType,
							Function:    msg.MessageHeader.Function,
						},
						Author: msg.Author,
					},
					Addrs: []string{msg.MessageHeader.RemoteAddr},
				}

				//check if such group exists
				var isNotExist bool
				var confKey int
				for key := range s.dialogues {
					if s.dialogues[key].Name == msg.Content {
						isNotExist = true
						confKey = key
						break
					}
				}
				if !isNotExist {
					response.Message.Content = "No dialogue with such name" + common.InputArrows
					response.Message.MessageHeader.ResponseStatus = common.Fail
					break
				}

				//check user member of conf
				_, ok := s.dialogues[confKey].Users[msg.Author]
				if !ok {
					response.Message.Content = "You're not member of this dialogue" + common.InputArrows
					response.Message.MessageHeader.ResponseStatus = common.Fail
					break
				}

				//make this user online in chat
				s.dialogues[confKey].Users[msg.Author].IsOnline = true

				//collect all messages from group
				var content string = "*You are in " + s.dialogues[confKey].Name + " dialogue*\n"
				for _, message := range s.dialogues[confKey].Messages {
					content += message.Author + ": " + message.Content + "\n"
				}

				//if there unread messages
				if s.dialogues[confKey].LastMessageAddr != "" {
					var res = common.ServerResponse{
						Message: common.Message{
							MessageHeader: common.MessageHeader{
								MessageType:    common.DialogueRoom,
								DestinationID:  confKey,
								ResponseStatus: common.Ok,
								RemoteAddr:     msg.MessageHeader.RemoteAddr,
							},
							Author:  msg.Author,
							Content: "*Message recieved*\n",
						},
						Addrs: []string{s.dialogues[confKey].LastMessageAddr},
					}
					s.sendMessage <- res
					s.dialogues[confKey].LastMessageAddr = ""
				}

				response.Message.Content = content
				response.Message.MessageHeader.DestinationID = confKey
				response.Message.MessageHeader.ResponseStatus = common.Ok
				response.Message.MessageHeader.ResponseCreateConf.Name = s.dialogues[confKey].Name

			case common.InviteToDialogue:
				//template of response
				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType: msg.MessageHeader.MessageType,
							Function:    msg.MessageHeader.Function,
						},
						Author: msg.Author,
					},
					Addrs: []string{msg.MessageHeader.RemoteAddr},
				}

				//check if user with such username exists
				usr, ok := s.users[msg.Content]
				if !ok {
					response.Message.Content = "There is no such user registered.\n"
					response.Message.MessageHeader.ResponseStatus = common.Fail
					break
				}

				//check if there is more than 2 users
				if len(s.dialogues[msg.MessageHeader.DestinationID].Users) == 2 {
					response.Message.Content = "*There are also 2 users in dialogue*\n"
					response.Message.MessageHeader.ResponseStatus = common.Fail
					break
				}

				//adding new user to conf
				s.dialogues[msg.MessageHeader.DestinationID].Users[usr.Username] = usr
				response.Message.Content = "*User " + usr.Username + " sucesfully added!*\n"

			case common.InviteToGroup:
				//template of response
				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType: msg.MessageHeader.MessageType,
							Function:    msg.MessageHeader.Function,
						},
						Author: msg.Author,
					},
					Addrs: []string{msg.MessageHeader.RemoteAddr},
				}

				//check if such user exists
				usr, ok := s.users[msg.Content]
				if !ok {
					response.Message.Content = "*There is no such user registered*\n"
					response.Message.MessageHeader.ResponseStatus = common.Fail
					break
				}

				//TODO : check if user already invited

				//adding new user to conf
				s.groups[msg.MessageHeader.DestinationID].Users[usr.Username] = usr
				response.Message.Content = "*User " + usr.Username + " sucesfully added!*\n"

			case common.Disconnect:
				response = common.ServerResponse{
					Message: common.Message{
						MessageHeader: common.MessageHeader{
							MessageType: msg.MessageHeader.MessageType,
							Function:    msg.MessageHeader.Function,
						},
						Author: msg.Author,
					},
					Addrs: []string{msg.MessageHeader.RemoteAddr},
				}

				if msg.Content == string(common.CommandGroupConnect) {
					s.groups[msg.MessageHeader.DestinationID].Users[msg.Author].IsOnline = false
				} else if msg.Content == string(common.CommandDialogueConnect) {
					s.dialogues[msg.MessageHeader.DestinationID].Users[msg.Author].IsOnline = false
				}
				response.Message.Content = common.InputArrows
			}
		}
		s.sendMessage <- response
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
