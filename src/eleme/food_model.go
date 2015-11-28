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
	templateString string
}

var foodpool = list.New()

func (this *FoodModel) fetch(id string) *FoodModel {
	ret := this.MysqlModel.fetch(id)
	if ret == nil { return nil }
	return ret.(*FoodModel)
}

func (this *FoodModel) create(redisConn *redis.Conn, id, stock, price int) *FoodModel {
	ret := new(FoodModel)
	ret.id = fmt.Sprintf("Food_%d", id)
	ret.realid = id
	ret.stock = stock
	ret.price = price

	(*redisConn).Do("SET", "food_stock_of_" + ret.id, strconv.Itoa(stock))

	datapool[ret.id] = ret
	foodpool.PushBack(ret)

	ret.templateString = fmt.Sprintf("{\"id\":%d,\"price\":%d,\"stock\":%%s}", id, price)

	return ret
}

func (this *FoodModel) reserve(redisConn *redis.Conn,count int) bool {
	key := "food_stock_of_" + this.id
	ret, _ := redis.Int((*redisConn).Do("DECRBY", key, count))
	if ret < 0 {
		(*redisConn).Do("INCRBY", key, count)
		return false
	} else {
		return true	
	}
}

func (this *FoodModel) dump(redisConn *redis.Conn) string {
	foodStock, _ := redis.String((*redisConn).Do("GET", "food_stock_of_" + this.id))
	return fmt.Sprintf(this.templateString, foodStock)
}

func (this *FoodModel) dumpAll(redisConn *redis.Conn) string {
	var buf []string
	for obj := foodpool.Front(); obj != nil; obj = obj.Next() {
		buf = append(buf, obj.Value.(*FoodModel).dump(redisConn))
	}
	return "[" + strings.Join(buf, ",") + "]"
}

var foodModel *FoodModel = new(FoodModel)
