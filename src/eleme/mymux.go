package eleme

import "net/http"
import "encoding/json"

type MyMux struct {
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
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
}

func foodsHandler(w http.ResponseWriter, r *http.Request) {
	userid := auth(w, r)
	if userid == "" { return }
	w.Write(food_cache)
}

func createCartHandler(w http.ResponseWriter, r *http.Request) {
	userid := auth(w, r)
	if userid == "" { return }
	w.Write([]byte("{\"cart_id\":\""  + userid + genRandomString()[:(16-len(userid))] + "\"}"))
}

func addFoodHandler(w http.ResponseWriter, r *http.Request, cartid string) {
	userid := auth(w, r)
	if userid == "" { return }
	if len(cartid) < 16 {
		cartError(w)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var req struct {
		FoodId int `json:"food_id"`
		Count int `json:"count"`
	}
	if decoder.Decode(&req) != nil {
		malformedJson(w)
		return
	}
	if cartid[:len(userid)] != userid {
		cartNotOwned(w)
		return
	}
	food, ok := foodrealidmap[req.FoodId]
	if !ok {
		foodError(w)
		return
	}
	redisConn := redisPool.Get()
	defer redisConn.Close()
	if addFood(&redisConn, cartid, userid, food, req.Count) == "1" {
		customError(w, "{\"code\":\"FOOD_OUT_OF_LIMIT\",\"message\":\"篮子中食物数量超过了三个\"}", 403)
		return 
	}
	noContent(w)
}

func makeOrderHandler(w http.ResponseWriter, r *http.Request) {
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
}

// not important
func viewOrderHandler(w http.ResponseWriter, r *http.Request) {
	userid := auth(w, r)
	if userid == "" { return }
	redisConn := redisPool.Get()
	cart := cartModel.fetch(&redisConn, userid2orderid(&redisConn, userid))
	redisConn.Close()
	if cart == nil {
		w.Write([]byte("[]"))
		return
	}
	w.Write([]byte("[" + cart.dump() + "]"))
}

// not important
func adminOrderHandler(w http.ResponseWriter, r *http.Request) {
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
}

// a home-made mux to improve performance
func (this *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path6 := path[:6]
	if path6 == "/login" {
		loginHandler(w, r)
		return
	}
	if path6 == "/foods" {
		foodsHandler(w, r)
		return
	}
	if path6 == "/carts" {
		if r.Method == "POST" {
			createCartHandler(w, r)
		} else {
			var cartid string
			pathLen := len(path)
			if pathLen < 23 {
				cartid = path[7:pathLen]
			} else {
				cartid = path[7:23]
			}
			addFoodHandler(w, r, cartid)
		}
		return
	}
	if path6 == "/order" {
		if r.Method == "POST" {
			makeOrderHandler(w, r)
		} else {
			viewOrderHandler(w, r)
		}
		return
	}
	if path6 == "/admin" {
		adminOrderHandler(w, r)
	}
}