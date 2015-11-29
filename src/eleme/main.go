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
var is_bad_order map[string] bool = map[string] bool {}
var userid_to_orderid map[string] string = map[string] string{}

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

	defer redisConn.Close()

	psc := redis.PubSubConn{Conn: redisConn}

	psc.Subscribe("bad_order")

	go func() {
		for {
			switch n := psc.Receive().(type) {
			case redis.Message:
				switch n.Channel {
				case "bad_order":
					is_bad_order[string(n.Data)] = true
				}
			}
		}
	}()

	/*go func() {
		for {
			time.Sleep(1300 * time.Millisecond)
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
	}()*/

	mux := &MyMux{}
	http.ListenAndServe(addr, mux)

}
