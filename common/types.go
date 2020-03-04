package common

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

//MessageHeader : represents message metadata
type MessageHeader struct {
	MessageType    MessageType       `json:"type"`
	DestinationID  int               `json:"destination_id"`
	Function       InstructionString `json:"function"`
	ResponseStatus ResponseStatus    `json:"response_status"`
	RemoteAddr     string            `json:"user_addr"`
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
