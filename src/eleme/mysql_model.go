package eleme

type MysqlModel struct {
	id string
}

var datapool = map[string] interface{}{}

func (this *MysqlModel) fetch(id string) interface{} {
	v, ok := datapool[id]
	if !ok {
		return nil	
	} else {
		return v	
	}
}

var mysqlModel *MysqlModel
