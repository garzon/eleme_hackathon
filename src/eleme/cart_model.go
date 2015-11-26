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

func (this *CartModel) save() {
	raw, _ := json.Marshal(this)
	this.saveRawData(string(raw))
}

func (this *CartModel) load() bool {
	raw := this.loadRawData()
	if raw == "" { return false }
	json.Unmarshal([]byte(raw), this)
	return true
}

func createCart(userid string) string {
	ret := new(CartModel)
	ret.DataModel.generateId()
	ret.Userid = userid
	ret.FoodIds = map[string] int{}
	ret.FoodCount = 0
	ret.Total = 0
	ret.IsBadOrder = false
	ret.IsOrder = false
	ret.save()
	return ret.Id
}

func (this *CartModel) fetch(cartid string) *CartModel {
	ret := new(CartModel)
	ret.Id = cartid
	if ret.load() == false { return nil }
	return ret
}

func (this *CartModel) addFood(food *FoodModel, count int) string {
	if this.IsOrder == true {
		return "{\"code\":\"ORDER_LOCKED\",\"message\":\"订单已经提交\"}"
	}
	if this.FoodCount + count > 3 {
		return "{\"code\":\"FOOD_OUT_OF_LIMIT\",\"message\":\"篮子中食物数量超过了三个\"}"
	}
	if food.reserve(count) {
		foodid := strconv.Itoa(food.realid)
		lastCount, ok := this.FoodIds[foodid]
		if !ok { lastCount = 0 }
		if lastCount + count < 0 { 
			food.reserve(-count)
			return ""
		}
		this.FoodCount += count
                this.Total += count * food.price
                foodid = strconv.Itoa(food.realid)	
		this.FoodIds[foodid] = lastCount + count
	} else {
		this.IsBadOrder = true
	}
	this.save()
	return ""
}

func userid2orderid(userid string) string {
	redisConn := redisPool.Get()
	defer redisConn.Close()
	ret, _ := redis.String(redisConn.Do("GET", "userid2orderid_" + userid))
	return ret
}

func (this *CartModel) makeOrder(userid string) string {
	if userid2orderid(userid) != "" {
		return "{\"code\":\"ORDER_OUT_OF_LIMIT\",\"message\":\"每个用户只能下一单\"}"
	}
	if this.IsBadOrder {
		return "{\"code\":\"FOOD_OUT_OF_STOCK\",\"message\":\"食物库存不足\"}"
	}
	redisConn := redisPool.Get()
        defer redisConn.Close()
	this.IsOrder = true
	redisConn.Do("SADD", "orders", this.Id)
	redisConn.Do("SET", "userid2orderid_" + userid, this.Id)
	this.save()
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

func (this *CartModel) dumpAll() string {
	redisConn := redisPool.Get()
        defer redisConn.Close()
	list, _ := redis.Values(redisConn.Do("SMEMBERS", "orders"))
	var buf []string
	for _, id := range list {
		buf = append(buf, cartModel.fetch(string(id.([]uint8))).dump())
	}
	return "[" + strings.Join(buf, ",") + "]"
}

var cartModel *CartModel
