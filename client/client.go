package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/COOLizh/itirod/UdpChat/common"
	"github.com/sirupsen/logrus"
)

// Client : contains all needful information about client
type Client struct {
	conn            *net.UDPConn
	addr            string
	username        string
	currChatID      int
	prevCommand     common.Command
	printMessage    chan common.Message
	sendMessage     chan common.Message
	recievedMessage chan common.Message
	dialogues       map[int][]common.Message
	groups          map[int][]common.Message
}

// SendMessage : recieves message to server
func (c *Client) SendMessage() {
	for {
		msg := <-c.sendMessage
		msgJSON, err := json.Marshal(msg)
		common.HandleError(err, common.ErrorFatal)
		msgJSON = append(msgJSON, '\n')
		c.conn.Write(msgJSON)
	}
}

// HandleRecievedMessage : processing server response
func (c *Client) HandleRecievedMessage() {
	for {
		msg := <-c.recievedMessage
		switch msg.MessageHeader.MessageType {
		case common.DialogueRoom:
		case common.GeneralRoom:
		case common.GroupRoom:
			c.currChatID = msg.MessageHeader.DestinationID
		case common.Instruction:
			switch msg.MessageHeader.Function {
			case common.CreateDialogue:
			case common.CreateGroup:
				if msg.MessageHeader.ResponseStatus == common.Ok {
					c.prevCommand = common.CommandCreateGroup
					c.currChatID = -1
				}
			case common.LogIn:
				if msg.MessageHeader.ResponseStatus == common.Ok {
					c.username = msg.Author
					c.currChatID = -1
				}
			case common.ConnectGroup:
				if msg.MessageHeader.ResponseStatus == common.Ok {
					c.prevCommand = common.CommandGroupConnect
					c.currChatID = msg.MessageHeader.DestinationID
				}
			case common.Disconnect:
				c.prevCommand = common.CommandChatDisconnect
				c.currChatID = -1
			}
		}
		c.printMessage <- msg
	}
}

// RecieveMessage : receives a message fo further processing
func (c *Client) RecieveMessage() {
	for {
		buff := make([]byte, 1024)
		bytes, _, err := c.conn.ReadFrom(buff)
		common.HandleError(err, common.ErrorFatal)
		var msg common.Message
		err = json.Unmarshal(buff[:bytes-1], &msg)
		common.HandleError(err, common.ErrorFatal)
		c.recievedMessage <- msg
	}
}

// PrintMessage : prints processed message
func (c *Client) PrintMessage() {
	for {
		msg := <-c.printMessage
		fmt.Print(msg.Content)
	}
}

// Input : provides user input
func (c *Client) Input() {
	for {
		isOkInput := true
		var command, attribute string
		fmt.Scanf("%s %s", &command, &attribute)
		msg := common.Message{
			Author: c.username,
		}
		switch command {
		case string(common.CommandCreateGroup):
			msg.MessageHeader = common.MessageHeader{
				MessageType: common.Instruction,
				Function:    common.CreateGroup,
				RequestCreateConf: common.RequestCreateConf{
					Name:      attribute,
					UserNames: []string{c.username},
				},
				RemoteAddr: c.addr,
			}
		case string(common.CommandGroupConnect):
			msg.MessageHeader = common.MessageHeader{
				MessageType: common.Instruction,
				Function:    common.ConnectGroup,
				RemoteAddr:  c.addr,
			}
			msg.Content = attribute
		case string(common.CommandInviteUser):
			isOkInput = false
			if c.prevCommand == common.CommandDialogueConnect {
				isOkInput = true
				// TODO
			} else if c.prevCommand == common.CommandGroupConnect {
				isOkInput = true
				msg.MessageHeader = common.MessageHeader{
					MessageType:   common.Instruction,
					Function:      common.InviteToGroup,
					DestinationID: c.currChatID,
					RemoteAddr:    c.addr,
				}
				msg.Content = attribute
			}
		case string(common.CommandChatDisconnect):
			msg.MessageHeader = common.MessageHeader{
				MessageType:   common.Instruction,
				Function:      common.Disconnect,
				DestinationID: c.currChatID,
				RemoteAddr:    c.addr,
			}
			msg.Content = c.username
		default:
			isOkInput = false
			if c.currChatID != -1 {
				isOkInput = true
				msg.MessageHeader = common.MessageHeader{
					MessageType:   common.GroupRoom,
					DestinationID: c.currChatID,
				}
				msg.Content = command + " " + attribute
				msg.MessageHeader.RemoteAddr = c.addr
				fmt.Print(c.username + " : " + command + " " + attribute)
			}
		}
		if isOkInput {
			c.sendMessage <- msg
		} else {
			msg.Content = common.CommandsInfo + common.InputArrows
			c.printMessage <- msg
		}
	}
}

func main() {
	config := common.GetConfig()
	var client = Client{
		dialogues:       make(map[int][]common.Message),
		groups:          make(map[int][]common.Message),
		printMessage:    make(chan common.Message),
		sendMessage:     make(chan common.Message),
		recievedMessage: make(chan common.Message),
	}
	var err error

	var username string
	fmt.Print("Enter your username: ")
	fmt.Scan(&username)

	client.username = username
	addr, err := net.ResolveUDPAddr(config.Network, config.BindAddr)
	common.HandleError(err, common.ErrorFatal)
	client.conn, err = net.DialUDP(config.Network, nil, addr)
	common.HandleError(err, common.ErrorFatal)
	defer func(c *Client) {
		fail := recover()
		if fail != nil {
			logrus.Error(fail)
		}
		c.conn.Close()
	}(&client)
	client.addr = client.conn.LocalAddr().String()
	msg := common.Message{
		MessageHeader: common.MessageHeader{
			MessageType: common.Instruction,
			Function:    common.LogIn,
			RemoteAddr:  client.addr,
		},
		Author: username,
	}

	go client.SendMessage()

	go client.RecieveMessage()

	go client.PrintMessage()

	go client.HandleRecievedMessage()

	client.sendMessage <- msg

	client.Input()
}
