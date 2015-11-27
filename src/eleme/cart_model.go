package eleme

import (
	"fmt"
	"strings"
	"encoding/json"
	"strconv"
	"github.com/garyburd/redigo/redis"
)

type CartModel struct {
	DataModel
	Userid string
	FoodIds map[string] int
	FoodCount int
	Total int
	IsBadOrder bool
	IsOrder bool
}

func (this *CartModel) save(redisConn redis.Conn) {
	raw, _ := json.Marshal(this)
	this.saveRawData(redisConn, string(raw))
}

func (this *CartModel) load(redisConn redis.Conn) bool {
	raw := this.loadRawData(redisConn)
	if raw == "" { return false }
	json.Unmarshal([]byte(raw), this)
	return true
}

func createCart(redisConn redis.Conn, userid string) string {
	ret := new(CartModel)
	ret.DataModel.generateId()
	ret.Userid = userid
	ret.FoodIds = map[string] int{}
	ret.FoodCount = 0
	ret.Total = 0
	ret.IsBadOrder = false
	ret.IsOrder = false
	ret.save(redisConn)
	return ret.Id
}

func (this *CartModel) fetch(redisConn redis.Conn, cartid string) *CartModel {
	ret := new(CartModel)
	ret.Id = cartid
	if ret.load(redisConn) == false { return nil }
	return ret
}

func (this *CartModel) addFood(redisConn redis.Conn, food *FoodModel, count int) string {
	if this.IsOrder {
		return "{\"code\":\"ORDER_LOCKED\",\"message\":\"订单已经提交\"}"
	}
	if this.IsBadOrder {
		return ""
	}
	if this.FoodCount + count > 3 {
		return "{\"code\":\"FOOD_OUT_OF_LIMIT\",\"message\":\"篮子中食物数量超过了三个\"}"
	}
	if food.reserve(redisConn, count) {
		foodid := strconv.Itoa(food.realid)
		lastCount, ok := this.FoodIds[foodid]
		if !ok { lastCount = 0 }
		if lastCount + count < 0 { 
			food.reserve(redisConn, -count)
			return ""
		}
		this.FoodCount += count
                this.Total += count * food.price
                foodid = strconv.Itoa(food.realid)	
		this.FoodIds[foodid] = lastCount + count
	} else {
		this.IsBadOrder = true
	}
	this.save(redisConn)
	return ""
}

func userid2orderid(redisConn redis.Conn, userid string) string {
	ret, _ := redis.String(redisConn.Do("GET", "userid2orderid_" + userid))
	return ret
}

func (this *CartModel) makeOrder(redisConn redis.Conn, userid string) string {
	if this.IsBadOrder {
		return "{\"code\":\"FOOD_OUT_OF_STOCK\",\"message\":\"食物库存不足\"}"
	}
	if userid2orderid(userid) != "" {
		return "{\"code\":\"ORDER_OUT_OF_LIMIT\",\"message\":\"每个用户只能下一单\"}"
	}
	this.IsOrder = true
	redisConn.Do("SADD", "orders", this.Id)
	redisConn.Do("SET", "userid2orderid_" + userid, this.Id)
	this.save(redisConn)
	return ""
}

func (this *CartModel) dump() string {
	var buf []string
	for id, count := range this.FoodIds {
		buf = append(buf, fmt.Sprintf("{\"food_id\":%s,\"count\":%d}", id, count))
	}
	ret := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":%s,", this.Id, strings.Replace(this.Userid, "User_", "", 1))
	ret += "\"items\":[" + strings.Join(buf, ",") + "],"
	ret += fmt.Sprintf("\"total\":%d}", this.Total)
	return ret
}

func (this *CartModel) dumpAll(redisConn redis.Conn) string {
	list, _ := redis.Values(redisConn.Do("SMEMBERS", "orders"))
	var buf []string
	for _, id := range list {
		buf = append(buf, cartModel.fetch(string(id.([]uint8))).dump())
	}
	return "[" + strings.Join(buf, ",") + "]"
}

var cartModel *CartModel
