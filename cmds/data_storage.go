package cmds

type StorageCommand struct {
	CreateData    CreateDataCommand    `cmd:"" name:"create-data" help:"create new storage data"`
	UpdateData    UpdateDataCommand    `cmd:"" name:"update-data" help:"update storage data"`
	DeleteData    DeleteDataCommand    `cmd:"" name:"delete-data" help:"delete storage data"`
	RegisterModel RegisterModelCommand `cmd:"" name:"register-model" help:"register storage model"`
}
