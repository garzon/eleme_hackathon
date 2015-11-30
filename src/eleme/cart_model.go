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
	return &CartModel{cartid, userid, map[string]int{}, 0, 0}
}

func (this *CartModel) fetch(redisConn *redis.Conn, cartid string) *CartModel {
	ret := new(CartModel)
	ret.Id = cartid
	if ret.load(redisConn) == false { return nil }
	return ret
}

func addFood(redisConn *redis.Conn, cartid, userid string, food *FoodModel, count int) string {
	ret, _ := redis.String((*redisConn).Do("EVALSHA", add_food_script_sha1, 5, cartid, "food_stock_of_" + food.id, "empty_cart", cartid + "_is_bad_order", "userid2orderid_" + userid, userid, food.realidString, count, food.price))
	return ret
}

func userid2orderid(redisConn *redis.Conn, userid string) string {
	ret, _ := redis.String((*redisConn).Do("GET", "userid2orderid_" + userid))
	return ret
}

func makeOrder(redisConn *redis.Conn, cartid, userid string) string {
	ret, _ := redis.String((*redisConn).Do("EVALSHA", make_order_script_sha1, 2, cartid + "_is_bad_order", "userid2orderid_" + userid, cartid))
	switch ret {
		case "1":
			return "{\"code\":\"FOOD_OUT_OF_STOCK\",\"message\":\"食物库存不足\"}"
		case "2":
			return "{\"code\":\"ORDER_OUT_OF_LIMIT\",\"message\":\"每个用户只能下一单\"}"
		default:
			return ""
	}
}

func (this *CartModel) dump() string {
	var buf []string
	for id, count := range this.FoodIds {
		buf = append(buf, fmt.Sprintf("{\"food_id\":%s,\"count\":%d}", id, count))
	}
	ret := "{\"id\":\"" + this.Id + "\",\"user_id\":" + this.Userid[5:] + ","
	ret += "\"items\":[" + strings.Join(buf, ",") + "],"
	ret += fmt.Sprintf("\"total\":%d}", this.Total)
	return ret
}

func (this *CartModel) dumpAll(redisConn *redis.Conn) string {
	var buf []string
	for _, user := range username2user {
		orderId := userid2orderid(redisConn, user.id)
		if orderId != "" {
			cart := cartModel.fetch(redisConn, orderId)
			if cart != nil {
				buf = append(buf, cart.dump())
			} else {
				buf = append(buf, "{\"id\":\"" + this.Id + "\",\"user_id\":" + this.Userid[5:] + ",\"items\":[],\"total\":0}")
			}
		}
	}
	return "[" + strings.Join(buf, ",") + "]"
}

var cartModel *CartModel
