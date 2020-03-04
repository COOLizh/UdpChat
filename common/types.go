package common

// TODO : add time stamp to message struct

type MessageType int

const (
	GeneralRoom MessageType = iota + 1
	DialogueRoom
	GroupRoom
	Instruction //instruction to server: login, create smth, ...
)

type InstructionString string

const (
	CreateDialogue InstructionString = "$create_dialogue"
	CreateGroup    InstructionString = "$create_group"
	LogIn          InstructionString = "$login"
)

type ErrorType uint

const (
	ErrorFatal ErrorType = iota + 1
	ErrorError
	ErrorInfo
	ErrorDebug
)

type ResponseStatus uint

const (
	Ok ResponseStatus = iota + 1
	Fail
)

type User struct {
	Username string
	Addr     string
}

type Conf struct {
	Name     string
	Messages []Message
	Users    map[string]*User
}

type RequestCreateConf struct {
	Name      string   `json:"name"`
	UserNames []string `json:"usernames"`
}

type ResponseCreateConf struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

//MessageHeader : represents message metadata
type MessageHeader struct {
	MessageType        MessageType        `json:"type"`
	DestinationID      int                `json:"destination_id"`
	Function           InstructionString  `json:"function"`
	ResponseStatus     ResponseStatus     `json:"response_status"`
	RemoteAddr         string             `json:"user_addr"`
	RequestCreateConf  RequestCreateConf  `json:"req_create_conf"`
	ResponseCreateConf ResponseCreateConf `json:"resp_create_conf"`
}

//Message : represents all needful information about message
type Message struct {
	MessageHeader MessageHeader `json:"header"`
	Author        string        `json:"author"`
	Content       string        `json:"content"`
}

type ServerResponse struct {
	Message Message
	Addrs   []string
}
