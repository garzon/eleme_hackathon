package eleme

import "github.com/garyburd/redigo/redis"

type DataModel struct {
	Id string `json:"id"`
}

func (this *DataModel) generateId() {
	this.Id = genRandomString()
}

func (this *DataModel) saveRawData(redisConn redis.Conn, rawData string) {
	redisConn.Do("SET", this.Id, rawData)
}

func (this *DataModel) loadRawData(redisConn redis.Conn) string {
	ret, _ := redis.String(redisConn.Do("GET", this.Id))
	return ret
}
