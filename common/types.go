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
	PrevCommand        string             `json:"prev_command"`
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

type Command string

// Commands
const (
	CommandCreateGroup     Command = "/create-group"
	CommandCreateDialogue  Command = "/create-dialogue"
	CommandInviteUser      Command = "/invite"
	ComandChatConnect      Command = "/connect"
	CommandChatDisconnect  Command = "/disconnect"
	CommandDisplayCommands Command = "/info"
)

const (
	CommandsInfo string = `There are all possible commands. Usage:
	/create-group [GroupName] - creates new group of users with unique name
	/create-dialogue [Username] - creates new dialogue between you and other user with his own username
	/invite [Username] - invites user in current group or dialogue
	/connect [ChatName or UserName] - connects to group or dialogue which you are a member of
	/disconnect - disconnects from current group or chat
	/info - displays information about possible commands`

	InputArrows string = "\n>> "
)
