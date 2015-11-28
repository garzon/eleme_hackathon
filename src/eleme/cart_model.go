package eleme

import (
	"fmt"
	"strings"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type CartModel struct {
	Id string
	Userid string
	FoodIds map[string] int
	FoodCount int
	Total int
	IsBadOrder bool
}

func (this *CartModel) save(redisConn *redis.Conn) {
	raw, _ := json.Marshal(this)
	(*redisConn).Do("SET", this.Id, string(raw))
}

func (this *CartModel) load(redisConn *redis.Conn) bool {
	raw, _ := redis.String((*redisConn).Do("GET", this.Id))
	if raw == "" { return false }
	json.Unmarshal([]byte(raw), this)
	return true
}

func createCart(userid, cartid string) *CartModel {
	return &CartModel{cartid, userid, map[string]int{}, 0, 0, false}
}

func (this *CartModel) fetch(redisConn *redis.Conn, cartid string) *CartModel {
	ret := new(CartModel)
	ret.Id = cartid
	if ret.load(redisConn) == false { return nil }
	return ret
}

func (this *CartModel) addFood(redisConn *redis.Conn, food *FoodModel, count int) string {
	/*if this.IsOrder {
		return "{\"code\":\"ORDER_LOCKED\",\"message\":\"订单已经提交\"}"
	}*/
	if this.IsBadOrder {
		return ""
	}
	if this.FoodCount + count > 3 {
		return "{\"code\":\"FOOD_OUT_OF_LIMIT\",\"message\":\"篮子中食物数量超过了三个\"}"
	}
	lastCount, ok := this.FoodIds[food.realidString]
	if !ok { lastCount = 0 }
	lastCount += count
	if lastCount < 0 {
		return ""
	}
	if food.reserve(redisConn, count) {
		this.FoodCount += count
		this.Total += count * food.price
		this.FoodIds[food.realidString] = lastCount
	} else {
		this.IsBadOrder = true
	}
	this.save(redisConn)
	return ""
}

func userid2orderid(redisConn *redis.Conn, userid string) string {
	ret, _ := redis.String((*redisConn).Do("GET", "userid2orderid_" + userid))
	return ret
}

func (this *CartModel) makeOrder(redisConn *redis.Conn, userid string) string {
	if this.IsBadOrder {
		return "{\"code\":\"FOOD_OUT_OF_STOCK\",\"message\":\"食物库存不足\"}"
	}
	ret, _ := redis.Int((*redisConn).Do("SETNX", "userid2orderid_" + userid, this.Id))
	if ret != 1 {
		return "{\"code\":\"ORDER_OUT_OF_LIMIT\",\"message\":\"每个用户只能下一单\"}"
	}
	//this.IsOrder = true
	//redisConn.Do("SADD", "orders", this.Id)
	//this.save(redisConn)
	return ""
}

func (this *CartModel) dump() string {
	var buf []string
	for id, count := range this.FoodIds {
		buf = append(buf, fmt.Sprintf("{\"food_id\":%s,\"count\":%d}", id, count))
	}
	ret := fmt.Sprintf("{\"id\":\"" + this.Id + "\",\"user_id\":" + this.Userid[5:] + ",")
	ret += "\"items\":[" + strings.Join(buf, ",") + "],"
	ret += fmt.Sprintf("\"total\":%d}", this.Total)
	return ret
}

func (this *CartModel) dumpAll(redisConn *redis.Conn) string {
	var buf []string
	/*
	list, _ := redis.Values(redisConn.Do("SMEMBERS", "orders"))
	for _, id := range list {
		buf = append(buf, cartModel.fetch(redisConn, string(id.([]uint8))).dump())
	}
	*/
	for _, user := range username2user {
		orderId := userid2orderid(redisConn, user.id)
		if orderId != "" {
			buf = append(buf, cartModel.fetch(redisConn, orderId).dump())
		}
	}
	return "[" + strings.Join(buf, ",") + "]"
}

var cartModel *CartModel
