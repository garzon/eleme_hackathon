#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import MySQLdb

from flask import Flask, abort, current_app
from redismodel import RedisModel
from redisstring import RedisString
import werkzeug.exceptions as ex

from util import *

from usermodel import UserModel
from foodmodel import FoodModel
from cartmodel import CartModel


# clean the cache last time
RedisModel.flushall()
app = Flask(__name__)


# Error handlers =====================================================================


class SyntaxErrorException(ex.HTTPException):
	code = 753
	description = '{"code": "MALFORMED_JSON", "message": "格式错误"}'


abort.mapping[753] = SyntaxErrorException


@app.errorhandler(753)
def syntax_error(error):
	return SyntaxErrorException.description, 400


@app.errorhandler(400)
def empty_req_error(error):
	return '{"code": "EMPTY_REQUEST", "message": "请求体为空"}', 400


@app.errorhandler(401)
def token_error(error):
	return '{"code": "INVALID_ACCESS_TOKEN", "message": "无效的令牌"}', 401


@app.errorhandler(403)
def auth_error(error):
	return '{"code": "USER_AUTH_FAIL", "message": "用户名或密码错误"}', 403


@app.errorhandler(404)
def cart_error(error):
	return '{"code": "CART_NOT_FOUND", "message": "篮子不存在"}', 404


# ====================================================================================


@app.before_first_request
def initialize_my_app():
	current_app.redis = None

	# model classes and ORM settings
	mysql_model_list = {'user': UserModel, 'food': FoodModel}

	# initialize the indexes and data structure of model classes
	current_app.datapool = {cls.__name__: dict() for cls in mysql_model_list.values()}
	for modelclass in mysql_model_list.values():
		modelclass.init_data_structure()

	# connect to mysql
	mysql_addr = os.getenv('DB_HOST', 'localhost')
	mysql_port = os.getenv('DB_PORT', '3306')
	mysql_db = os.getenv('DB_NAME', 'eleme')
	mysql_user = os.getenv('DB_USER', 'root')
	mysql_pass = os.getenv('DB_PASS', 'toor')
	mysql = MySQLdb.connect(host=mysql_addr, port=int(mysql_port), user=mysql_user, passwd=mysql_pass, db=mysql_db)

	# load all the data
	cursor = mysql.cursor()
	for table_name, model_class in mysql_model_list.items():
		cursor.execute("SELECT * from " + table_name)
		data = cursor.fetchall()
		for datum in data:
			model_class.parse(datum)

	# connection is useless now
	mysql.close()

	current_app.food_template_str = '[' + ','.join(current_app.food_template_str) + ']'


# Controllers ========================================================================

@app.route('/')
def hello_world():
	return 'Hello World!'


@app.route('/login', methods=['POST'])
def login_handler():
	data = parse_req_body()
	try:
		username = data['username']
		password = data['password']
	except:
		abort(403)

	user = UserModel.login(username, password)
	if user is False:
		abort(403)
	return '{"user_id":%s,"username":"%s","access_token":"%s"}' % (str(user.id), user.username, user.token), 200


@app.route('/foods', methods=['GET'])
def foods_handler():
	auth()
	stocks = tuple(RedisModel.get_redis().mget(current_app.food_ids_arr))
	return current_app.food_template_str % stocks, 200


@app.route('/carts', methods=['POST'])
def new_carts_handler():
	userid = auth()
	cart = CartModel()
	cart.userid = userid
	cart.save()
	return '{"cart_id": "' + cart.id + '"}', 200


@app.route('/carts/<cart_id>', methods=['PATCH'])
def carts_handler(cart_id):
	userid = auth()
	data = parse_req_body()
	food_id = data.get('food_id', '-1')
	count = int(data.get('count', '0'))
	cart = CartModel.fetch(cart_id)
	if cart is None:
		abort(404)
	if cart.userid != userid:
		return '{"code": "NOT_AUTHORIZED_TO_ACCESS_CART", "message": "无权限访问指定的篮子"}', 401
	if FoodModel.fetch(food_id) is None:
		return '{"code": "FOOD_NOT_FOUND", "message": "食物不存在"}', 404
	ret = cart.add_food(food_id, count)
	if ret is True:
		return '', 204
	return ret, 403


@app.route('/orders', methods=['POST', 'GET'])
def orders_handler():
	userid = auth()
	if request.method == 'POST':
		# POST
		data = parse_req_body()
		cart_id = data.get('cart_id', '')
		cart_userid, cart_id = CartModel.fetchCols(cart_id, ['userid', 'id'])
		if (cart_id is None) or (cart_userid != str(userid)):
			return '{"code": "NOT_AUTHORIZED_TO_ACCESS_CART", "message": "cart not owned by user"}', 401
		try:
			order = OrderModel(cartid = cart_id)
			order.save()
		except Exception as inst:
			if inst.args[0] == OrderModel.OUT_OF_LIMIT:
				return '{"code": "ORDER_OUT_OF_LIMIT", "message": "每个用户只能下一单"}', 403
			elif inst.args[0] == OrderModel.OUT_OF_STOCK:
				return '{"code": "FOOD_OUT_OF_STOCK", "message": "食物库存不足"}', 403
			else:
				return str(inst.args), 555
		return '{"id": "' + order.id + '"}', 200
	else:
		# GET
		orderid = OrderModel.fetch_orderid_by_userid(userid)
		if orderid is None:
			return '[]', 200
		else:
			return '[' + OrderModel.dump(orderid) + ']', 200


@app.route('/admin/orders', methods=['GET'])
def admin_orders_handler():
	admin_auth()
	orderids = OrderModel.fetch_all_orderid()
	return '[' + ','.join(map(lambda order_id: OrderModel.dump(order_id), orderids)) + ']', 200


if __name__ == '__main__':
	host = os.getenv("APP_HOST", "localhost")
	port = int(os.getenv("APP_PORT", "8080"))
	app.run(host=host, port=port)

