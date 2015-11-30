package eleme

import (
	"time"
	"fmt"
	"net/http"
	"os"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/garyburd/redigo/redis"
)

var redisPool *redis.Pool

func newRedisPool(server string) *redis.Pool {
      return &redis.Pool{
          MaxIdle: 2000,
          MaxActive: 30000, // max number of connections
          IdleTimeout: 240 * time.Second,
          Dial: func () (redis.Conn, error) {
              c, _ := redis.Dial("tcp", server)
              return c, nil
          },
          //Wait: true,
          /*TestOnBorrow: func(c redis.Conn, t time.Time) error {
              _, err := c.Do("PING")
              return err
          },*/
      }
}

func getEnv(name, defVal string) string {
	ret := os.Getenv(name)
	if ret == "" {
		return defVal
	} else {
		return ret
	}
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

var food_cache []byte

var add_food_script_sha1 string
var make_order_script_sha1 string

func Eleme() {
	host := getEnv("APP_HOST", "localhost")
	port := getEnv("APP_PORT", "8080")
	mysqlHost := getEnv("DB_HOST", "localhost")
	mysqlPort := getEnv("DB_POST", "3306")
	mysqlDb := getEnv("DB_NAME", "eleme")
	mysqlUser := getEnv("DB_USER", "root")
	mysqlPass := getEnv("DB_PASS", "toor")

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")

	redisPool = newRedisPool(fmt.Sprintf("%s:%s", redisHost, redisPort))
	
	redisConn := redisPool.Get()
	redisConn.Do("FLUSHALL")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPass, mysqlHost, mysqlPort, mysqlDb))
	checkErr(err)

	rows, err := db.Query("SELECT * FROM user")
	checkErr(err)
	for rows.Next() {
		var id int
		var username string
		var password string
		checkErr(rows.Scan(&id, &username, &password))
		userModel.create(id, username, password)
	}

	rows, err = db.Query("SELECT * FROM food")
	checkErr(err)
	for rows.Next() {
		var id int
		var stock int
		var price int
		checkErr(rows.Scan(&id, &stock, &price))
		foodModel.create(&redisConn, id, stock, price)
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	food_cache = []byte(foodModel.dumpAll(&redisConn))

	empty_cart := createCart("", "empty_cart")
	empty_cart.save(&redisConn)

// using redis lua script to reduce the number of tcp connection to redis
	add_food_script := `
if redis.call("GET", KEYS[5]) ~= false then
  return "0"
end
local cart = redis.call("GET", KEYS[1])
local flag = true
if cart == false then
  cart = redis.call("GET", KEYS[3])
  flag = false
end
local cart_obj = cjson.decode(cart)
if flag == false then
  cart_obj['Id'] = KEYS[1]
  cart_obj['Userid'] = ARGV[1]
end
if cart_obj['FoodCount'] + ARGV[3] > 3 then
  if flag == false then
    redis.call("SET", KEYS[1], cjson.encode(cart_obj))
  end
  return "1"
end
local old_num = 0
if cart_obj['FoodIds'][ARGV[2]] ~= nil then
  old_num = cart_obj['FoodIds'][ARGV[2]]
end
old_num = old_num + ARGV[3]
if old_num < 0 then
  if flag == false then
    redis.call("SET", KEYS[1], cjson.encode(cart_obj))
  end
  return "0"
end
if redis.call("DECRBY", KEYS[2], ARGV[3]) < 0 then
  redis.call("INCRBY", KEYS[2], ARGV[3])
  if flag == false then
    redis.call("SET", KEYS[1], cjson.encode(cart_obj))
  end
  redis.call("SET", KEYS[4], "1")
  return "0"
end
cart_obj['FoodIds'][ARGV[2]] = old_num
cart_obj['FoodCount'] = cart_obj['FoodCount'] + ARGV[3]
cart_obj['Total'] = cart_obj['Total'] + ARGV[3] * ARGV[4]
redis.call("SET", KEYS[1], cjson.encode(cart_obj))
return "0"
`

	make_order_script := `
if redis.call("GET", KEYS[1]) ~= false then
  return "1"
end
local ret = redis.call("SETNX", KEYS[2], ARGV[1])
if ret ~= 1 then
  return "2"
end
return "0"
`

	add_food_script_sha1, err = redis.String(redisConn.Do("SCRIPT", "LOAD", add_food_script))
	checkErr(err)

	make_order_script_sha1, err = redis.String(redisConn.Do("SCRIPT", "LOAD", make_order_script))
	checkErr(err)

	defer redisConn.Close()

	// use a gorouting to update the food cache in the background
	go func() {
		for {
			time.Sleep(2500 * time.Millisecond)
			lock, _ := redis.Int(redisConn.Do("SETNX", "foods_cache_lock", "1"))
			ret := ""
			if lock == 1 {
				ret = foodModel.dumpAll(&redisConn)
				redisConn.Do("SET", "foods_cache", ret)
				redisConn.Do("DEL", "foods_cache_lock")
			} else {
				ret, _ = redis.String(redisConn.Do("GET", "foods_cache"))
			}
			food_cache = []byte(ret)
		}
	}()

	mux := &MyMux{}
	http.ListenAndServe(addr, mux)

}
