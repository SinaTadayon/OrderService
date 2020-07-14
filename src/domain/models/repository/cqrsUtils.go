package repository

type RepoType string
type CmdType string

const (
	OrderRepo  RepoType = "ORDER"
	PkgRepo    RepoType = "PKG"
	SubPkgRepo RepoType = "SUB_PKG"
)

const (
	SaveCmd          CmdType = "SAVE"
	SaveAllCmd       CmdType = "SAVE_ALL"
	UpdateCmd        CmdType = "UPDATE"
	UpdateAllCmd     CmdType = "UPDATE_ALL"
	InsertCmd        CmdType = "INSERT"
	InsertAllCmd     CmdType = "INSERT_ALL"
	DeleteCmd        CmdType = "DELETE"
	DeletePartialCmd CmdType = "DELETE_PARTIAL"
	DeleteAllCmd     CmdType = "DELETE_ALL"
	RemoveCmd        CmdType = "REMOVE"
	RemoveEntityCmd  CmdType = "REMOVE_ENTITY"
	RemovePartialCmd CmdType = "REMOVE_PARTIAL"
	RemoveAllCmd     CmdType = "REMOVE_ALL"
)

type CommandData struct {
	Repository RepoType
	Command    CmdType
	Data       interface{}
}

type CommandReaderStream <-chan *CommandData
type CommandWriterStream chan<- *CommandData
type CommandStream chan *CommandData
