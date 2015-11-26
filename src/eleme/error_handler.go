package eleme

import "net/http"

func malformedJson(w http.ResponseWriter) {
	http.Error(w, "{\"code\": \"MALFORMED_JSON\", \"message\": \"格式错误\"}", 400)
}

func emptyRequest(w http.ResponseWriter) {
	http.Error(w, "{\"code\": \"EMPTY_REQUEST\", \"message\": \"请求体为空\"}", 400)
}

func invalidToken(w http.ResponseWriter) {
	http.Error(w, "{\"code\": \"INVALID_ACCESS_TOKEN\", \"message\": \"无效的令牌\"}", 401)
}

func authError(w http.ResponseWriter) {
	http.Error(w, "{\"code\": \"USER_AUTH_FAIL\", \"message\": \"用户名或密码错误\"}", 403)
}

func cartError(w http.ResponseWriter) {
	http.Error(w, "{\"code\": \"CART_NOT_FOUND\", \"message\": \"篮子不存在\"}", 404)
}

func cartNotOwned(w http.ResponseWriter) {
	http.Error(w, "{\"code\": \"NOT_AUTHORIZED_TO_ACCESS_CART\", \"message\": \"无权限访问指定的篮子\"}", 401)
}

func foodError(w http.ResponseWriter) {
	http.Error(w, "{\"code\": \"FOOD_NOT_FOUND\", \"message\": \"食物不存在\"}", 404)
}

func noContent(w http.ResponseWriter) {
	http.Error(w, "", 204)
}

func customError(w http.ResponseWriter, body string, code int) {
	http.Error(w, body, code)
}

