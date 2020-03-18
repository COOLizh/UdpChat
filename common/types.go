package common

// TODO : add time stamp to message struct

// MessageType : needed to know where command inputed
type MessageType int

// MessageTypes
const (
	GeneralRoom MessageType = iota + 1
	DialogueRoom
	GroupRoom
	Instruction //instruction to server: login, create smth, ...
)

// InstructionString : needed for server to know what to do
type InstructionString string

// InstructionStrings
const (
	CreateDialogue   InstructionString = "$create_dialogue"
	CreateGroup      InstructionString = "$create_group"
	LogIn            InstructionString = "$login"
	ConnectGroup     InstructionString = "$connect_group"
	ConnectDialogue  InstructionString = "$connect_dialogue"
	InviteToDialogue InstructionString = "$invite_to_dialogue"
	InviteToGroup    InstructionString = "$invite_to_group"
)

// ErrorType : needed to know what kind of error happend
type ErrorType uint

//ErrorTypes
const (
	ErrorFatal ErrorType = iota + 1
	ErrorError
	ErrorInfo
	ErrorDebug
)

// ResponseStatus : needed to know was the request successful or not
type ResponseStatus uint

// ResponseStatuses
const (
	Ok ResponseStatus = iota + 1
	Fail
)

// User : contains all needful information about user in dialogue or group
type User struct {
	Username string
	Addr     string
	IsOnline bool
}

// Conf : contains all needful information about group
type Conf struct {
	Name     string
	Messages []Message
	Users    map[string]*User
}

// RequestCreateConf : needed for MesageHeader
type RequestCreateConf struct {
	Name      string   `json:"name"`
	UserNames []string `json:"usernames"`
}

// ResponseCreateConf ...
type ResponseCreateConf struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// MessageHeader : represents message metadata
type MessageHeader struct {
	MessageType        MessageType        `json:"type"`
	DestinationID      int                `json:"destination_id"`
	Function           InstructionString  `json:"function"`
	ResponseStatus     ResponseStatus     `json:"response_status"`
	RemoteAddr         string             `json:"user_addr"`
	RequestCreateConf  RequestCreateConf  `json:"req_create_conf"`
	ResponseCreateConf ResponseCreateConf `json:"resp_create_conf"`
}

// Message : represents all needful information about message
type Message struct {
	MessageHeader MessageHeader `json:"header"`
	Author        string        `json:"author"`
	Content       string        `json:"content"`
}

// ServerResponse : contains all needful inforation about server response
type ServerResponse struct {
	Message Message
	Addrs   []string
}

// Command : describes all possible commands
type Command string

// Commands
const (
	CommandCreateGroup     Command = "/create-group"
	CommandCreateDialogue  Command = "/create-dialogue"
	CommandInviteUser      Command = "/invite"
	CommandGroupConnect    Command = "/connect-group"
	CommandDialogueConnect Command = "/connect-dialogue"
	CommandChatDisconnect  Command = "/disconnect"
	CommandDisplayCommands Command = "/info"
)

// Additional information to client
const (
	CommandsInfo string = `There are all possible commands. Usage:
	/create-group [GroupName] - creates new group of users with unique name
	/create-dialogue [Username] - creates new dialogue between you and other user with his own username
	/invite [Username] - can be used after connection to group or dialogue, invites user in current group or dialogue
	/connect-group [GroupName] - connects to group which you are a member of
	/connect-dialogue [Username] - connects to dialogue
	/disconnect - can be used after connection to group or dialogue, disconnects from current group or chat
	/info - displays information about possible commands`

	InputArrows string = "\n>> "
)
