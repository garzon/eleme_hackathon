package eleme

import "fmt"
import "github.com/garyburd/redigo/redis"

type UserModel struct {
	MysqlModel
	realid int
	username string
	password string
	token string
}

var username2user = map[string] *UserModel {}

func (this *UserModel) create(id int, username, password string) *UserModel {
	ret := new(UserModel)
	ret.id = fmt.Sprintf("User_%d", id)
	ret.realid = id
	ret.username = username
	ret.password = password
	ret.token = ""
	datapool[ret.id] = ret
	username2user[username] = ret
	ret.updateToken()
	return ret
}

func (this *UserModel) findUserIdByToken(token string) string {
	redisConn := redisPool.Get()
	ret, _ := redis.String(redisConn.Do("GET", "token2userid_" + token))
	redisConn.Close()
	return ret
}

func (this *UserModel) updateToken() string {
	this.token = genRandomString()
	redisConn := redisPool.Get()
	redisConn.Do("SET", "token2userid_" + this.token, this.id)
	redisConn.Close()
	return this.token
}

func (this *UserModel) login(username, password string) *UserModel {
	user, ok := username2user[username]
	if !ok { return nil }
	if user.password != password { return nil }
	//user.updateToken()
	return user
}

var userModel *UserModel

