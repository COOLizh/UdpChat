package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/COOLizh/itirod/UdpChat/common"
	"github.com/sirupsen/logrus"
)

type Client struct {
	conn            *net.UDPConn
	addr            string
	username        string
	prevCommand     common.Command
	printMessage    chan common.Message
	sendMessage     chan common.Message
	recievedMessage chan common.Message
	dialogues       map[int][]common.Message
	groups          map[int][]common.Message
}

func (c *Client) SendMessage() {
	for {
		msg := <-c.sendMessage
		msgJSON, err := json.Marshal(msg)
		common.HandleError(err, common.ErrorFatal)
		msgJSON = append(msgJSON, '\n')
		c.conn.Write(msgJSON)
	}
}

func (c *Client) HandleRecievedMessage() {
	for {
		msg := <-c.recievedMessage
		// do smth
		switch msg.MessageHeader.MessageType {
		case common.DialogueRoom:
		case common.GeneralRoom:
		case common.GroupRoom:
		case common.Instruction:
			switch msg.MessageHeader.Function {
			case common.CreateDialogue:
			case common.CreateGroup:
			case common.LogIn:
				if msg.MessageHeader.ResponseStatus == common.Ok {
					c.username = msg.Author
				}
			}
		}
		c.printMessage <- msg
	}
}

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

func (c *Client) PrintMessage() {
	for {
		msg := <-c.printMessage
		fmt.Print(msg.Content)
	}
}

func (c *Client) Input() {
	for {
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
			c.prevCommand = common.CommandCreateGroup
		case string(common.CommandGroupConnect):
			msg.MessageHeader = common.MessageHeader{
				MessageType: common.Instruction,
				Function:    common.ConnectGroup,
				RequestConnectConf: common.RequestConnectConf{
					ConfName: attribute,
				},
				RemoteAddr: c.addr,
			}
		}

		c.sendMessage <- msg
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
