package eleme

import "github.com/garyburd/redigo/redis"

type DataModel struct {
	Id string `json:"id"`
}

func (this *DataModel) generateId() {
	this.Id = genRandomString()
}

func (this *DataModel) saveRawData(rawData string) {
	redisConn := redisPool.Get()
        defer redisConn.Close()
	redisConn.Do("SET", this.Id, rawData)
}

func (this *DataModel) loadRawData() string {
	redisConn := redisPool.Get()
        defer redisConn.Close()
	ret, _ := redis.String(redisConn.Do("GET", this.Id))
	return ret
}
