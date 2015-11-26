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
          MaxIdle: 80,
          MaxActive: 12000, // max number of connections
          IdleTimeout: 240 * time.Second,
          Dial: func () (redis.Conn, error) {
              c, err := redis.Dial("tcp", server)
              if err != nil {
                  return nil, err
              }
              return c, err
          },
          TestOnBorrow: func(c redis.Conn, t time.Time) error {
              _, err := c.Do("PING")
              return err
          },
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
	redisConn.Close()

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
		foodModel.create(id, stock, price)
	}

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
				w.Write([]byte(fmt.Sprintf("{\"user_id\":%d,\"username\":\"%s\",\"access_token\":\"%s\"}", user.realid, user.username, user.token)))
			}
		}
	}))

	mux.Get("/foods", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		redisConn := redisPool.Get()
		ret, _ := redis.String(redisConn.Do("GET", "foods_cache"))
		if ret == "" {
			ret = foodModel.dumpAll(redisConn)
			redisConn.Do("PSETEX", "foods_cache", 300, ret)
		}
		redisConn.Close()
		w.Write([]byte(ret))
	}))

	mux.Post("/carts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		cartid := createCart(userid)
		w.Write([]byte(fmt.Sprintf("{\"cart_id\":\"%s\"}", cartid)))
	}))

	mux.Add("PATCH", "/carts/:cartId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
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
		cart := cartModel.fetch(cartid)
		if cart == nil { 
			cartError(w)
			return
		}
		if cart.Userid != userid {
			cartNotOwned(w)
			return
		}
		food := foodModel.fetch(fmt.Sprintf("Food_%d", req.FoodId))
		if food == nil {
			foodError(w)
			return
		}
		err := cart.addFood(food, req.Count)
		if err == "" {
			noContent(w)
			return
		}
		customError(w, err, 403)
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
		cart := cartModel.fetch(req.CartId)
		if cart.Userid != userid {
			cartNotOwned(w)
			return
		}
		ret := cart.makeOrder(userid)
		if ret != "" {
			customError(w, ret, 403)
			return
		}
		w.Write([]byte(fmt.Sprintf("{\"id\":\"%s\"}", cart.Id)))
	}))

	mux.Get("/orders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		cart := cartModel.fetch(userid2orderid(userid))
		if cart == nil {
			w.Write([]byte("[]"))
			return
		}
		w.Write([]byte("[" + cart.dump() + "]"))
	}))

	mux.Get("/admin/orders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userid := auth(w, r)
		if userid == "" { return }
		if userid != "User_1" {
			invalidToken(w)			
			return
		}
		w.Write([]byte(cartModel.dumpAll()))
	}))

	http.Handle("/", mux)

	http.ListenAndServe(addr, nil)

}
