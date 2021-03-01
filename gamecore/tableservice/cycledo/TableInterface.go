package cycledo


type TableInterface struct {
	TS tableService
}

func New(ts tableService)  *TableInterface{
	return &TableInterface{
		TS: ts,
	}
}

type tableService interface {
	AddTable(clientId uint64, tableUuid string,addFlag int32)
}

func (ti *TableInterface) AddTableTi(clientId uint64, tableUuid string,addFlag int32) {
	ti.TS.AddTable(clientId,tableUuid,addFlag)
}
