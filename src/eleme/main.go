package eleme

import (
	"time"
	"fmt"
	"net/http"
	"os"
	"encoding/json"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/garyburd/redigo/redis"
	"github.com/bmizerany/pat"
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

	redisConn.Close()

	addr := fmt.Sprintf("%s:%s", host, port)

	mux := pat.New()

	mux.Post("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			emptyRequest(w)
			return
		}
		decoder := json.NewDecoder(r.Body)
		var req struct {
			Username string
			Password string
		}
		if decoder.Decode(&req) != nil {
			malformedJson(w)
			return
		} else {
			user := userModel.login(req.Username, req.Password)
			if user == nil {
				authError(w)
				return
			} else {
				w.Write([]byte("{\"user_id\":" + user.realidString + ",\"username\":\"" + user.username + "\",\"access_token\":\"" + user.token + "\"}"))
			}
		}
	}))

	mux.Get("/foods", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		redisConn := redisPool.Get()
		ret, _ := redis.String(redisConn.Do("GET", "foods_cache"))
		if ret == "" {
			lock, _ := redis.Int(redisConn.Do("SETNX", "foods_cache_lock", "1"))
			if lock == 1 {
				ret = foodModel.dumpAll(&redisConn)
				redisConn.Do("PSETEX", "foods_cache", 1300, ret)
				redisConn.Do("SET", "foods_cache_forever", ret)
				redisConn.Do("DEL", "foods_cache_lock")
			} else {
				ret, _ = redis.String(redisConn.Do("GET", "foods_cache_forever"))
			}
		}
		redisConn.Close()
		w.Write([]byte(ret))
	}))

	mux.Post("/carts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		//cartid := createCart(userid)
		w.Write([]byte("{\"cart_id\":\""  + userid + genRandomString() + "\"}"))
	}))

	mux.Add("PATCH", "/carts/:cartId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		redisConn := redisPool.Get()
		defer redisConn.Close()
		if userid2orderid(&redisConn, userid) != "" {
			noContent(w)
			return
		}
		cartid := r.URL.Query().Get(":cartId")
		decoder := json.NewDecoder(r.Body)
		var req struct {
			FoodId int `json:"food_id"`
			Count int `json:"count"`
		}
		if decoder.Decode(&req) != nil {
			malformedJson(w)
			return
		}
		if len(cartid) <= 16 {
			cartError(w)
			return
		}
		if cartid[:len(userid)] != userid {
			cartNotOwned(w)
			return
		}
		cart := cartModel.fetch(&redisConn, cartid)
		if cart == nil {
			cart = createCart(userid, cartid)
		}
		food, ok := foodrealidmap[req.FoodId]
		if !ok {
			foodError(w)
			return
		}
		if cart.FoodCount + req.Count > 3 {
			customError(w, "{\"code\":\"FOOD_OUT_OF_LIMIT\",\"message\":\"篮子中食物数量超过了三个\"}", 403)
			return 
		}
		cart.addFood(&redisConn, food, req.Count)
		noContent(w)
	}))

	mux.Post("/orders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		var req struct {
			CartId string `json:"cart_id"`
		}
		decoder := json.NewDecoder(r.Body)
		if decoder.Decode(&req) != nil {
			malformedJson(w)
			return
		}
		if req.CartId[:len(userid)] != userid {
			cartNotOwned(w)
			return
		}
		redisConn := redisPool.Get()
		ret := makeOrder(&redisConn, req.CartId, userid)
		redisConn.Close()
		if ret != "" {
			customError(w, ret, 403)
			return
		}
		w.Write([]byte("{\"id\":\"" + req.CartId +  "\"}"))
	}))

	// not important
	mux.Get("/orders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		//time.Sleep(1000)
		redisConn := redisPool.Get()
		cart := cartModel.fetch(&redisConn, userid2orderid(&redisConn, userid))
		redisConn.Close()
		if cart == nil {
			w.Write([]byte("[]"))
			return
		}
		w.Write([]byte("[" + cart.dump() + "]"))
	}))

	// not important
	mux.Get("/admin/orders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		if userid != "User_1" {
			invalidToken(w)			
			return
		}
		redisConn := redisPool.Get()
		ret := cartModel.dumpAll(&redisConn)
		redisConn.Close()
		w.Write([]byte(ret))
	}))

	http.Handle("/", mux)

	http.ListenAndServe(addr, nil)

}
