package eleme

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/garyburd/redigo/redis"
	"container/list"
)

type FoodModel struct {
	MysqlModel
	realid int
	stock int
	price int
}

var foodpool = list.New()

func (this *FoodModel) fetch(id string) *FoodModel {
	ret := this.MysqlModel.fetch(id)
	if ret == nil { return nil }
	return ret.(*FoodModel)
}

func (this *FoodModel) create(id, stock, price int) *FoodModel {
	ret := new(FoodModel)
	ret.id = fmt.Sprintf("Food_%d", id)
	ret.realid = id
	ret.stock = stock
	ret.price = price
	redisConn := redisPool.Get()
        defer redisConn.Close()
	redisConn.Do("SET", "food_stock_of_" + ret.id, strconv.Itoa(stock))

	datapool[ret.id] = ret
	foodpool.PushBack(ret)
	return ret
}

func (this *FoodModel) reserve(count int) bool {
	key := "food_stock_of_" + this.id
	redisConn := redisPool.Get()
        defer redisConn.Close()
	ret, _ := redis.Int(redisConn.Do("DECRBY", key, count))
	if ret < 0 {
		redisConn.Do("INCRBY", key, count)
		return false
	} else {
		return true	
	}
}

func (this *FoodModel) dump() string {
	redisConn := redisPool.Get()
        defer redisConn.Close()
	foodStock, _ := redis.String(redisConn.Do("GET", "food_stock_of_" + this.id))
	return fmt.Sprintf("{\"id\":%d,\"price\":%d,\"stock\":%s}", this.realid, this.price, foodStock)
}

func (this *FoodModel) dumpAll() string {
	var buf []string
	for obj := foodpool.Front(); obj != nil; obj = obj.Next() {
		buf = append(buf, obj.Value.(*FoodModel).dump())
	}
	return "[" + strings.Join(buf, ",") + "]"
}

var foodModel *FoodModel = new(FoodModel)