package eleme

import "strconv"

type UserModel struct {
	MysqlModel
	realid int
	realidString string
	username string
	password string
	token string
}

var username2user = map[string] *UserModel {}
var token2user = map[string] *UserModel {}

func (this *UserModel) fetch(id string) *UserModel {
	ret := this.MysqlModel.fetch(id)
	if ret == nil { return nil }
	return ret.(*UserModel)
}

func (this *UserModel) create(id int, username, password string) *UserModel {
	ret := new(UserModel)
	ret.realidString = strconv.Itoa(id)
	ret.id = "User_" + ret.realidString
	ret.realid = id
	ret.username = username
	ret.password = password
	ret.token = md5_hex(password)

	datapool[ret.id] = ret
	username2user[username] = ret
	token2user[ret.token] = ret

	return ret
}

func (this *UserModel) findUserIdByToken(token string) string {
	user, ok := token2user[token]
	if !ok { return "" }
	return user.id
}

func (this *UserModel) login(username, password string) *UserModel {
	user, ok := username2user[username]
	if !ok { return nil }
	if user.password != password { return nil }
	return user
}

var userModel *UserModel

