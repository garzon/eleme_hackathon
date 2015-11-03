#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import MySQLdb

from flask import Flask, abort, current_app
import werkzeug.exceptions as ex

from util import *

from usermodel import UserModel
from foodmodel import FoodModel
from cartmodel import CartModel

app = Flask(__name__)


# Error handlers =====================================================================


class SyntaxErrorException(ex.HTTPException):
	code = 753
	description = '{"message": "malformed json"}'


abort.mapping[753] = SyntaxErrorException


@app.errorhandler(753)
def syntax_error(error):
	return SyntaxErrorException.description, 753


@app.errorhandler(400)
def empty_req_error(error):
	return '{"code": "EMPTY_REQUEST", "message": "empty request"}', 400


@app.errorhandler(401)
def token_error(error):
	return '{"code": "INVALID_ACCESS_TOKEN", "message": "invalid access token"}', 401


@app.errorhandler(403)
def auth_error(error):
	return '{"code": "USER_AUTH_FAIL", "message": "username or password incorrect"}', 403


@app.errorhandler(404)
def cart_error(error):
	return '{"code": "CART_NOT_FOUND", "message": "篮子不存在"}', 404


# ====================================================================================


@app.before_first_request
def initialize_my_app():
	# model classes and ORM settings
	mysql_model_list = {'user': UserModel, 'food': FoodModel}
	data_model_list = [CartModel, OrderModel]

	# initialize the indexes and data structure of model classes
	current_app.datapool = {cls.__name__: dict() for cls in mysql_model_list.values()}
	for modelclass in mysql_model_list.values():
		modelclass.init_data_structure()
	for datamodel in data_model_list:
		datamodel.init_data_structure()

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
		cursor.execute("SELECT * from %s" % table_name)
		data = cursor.fetchall()
		for datum in data:
			model_class.parse(datum)

	# connection is useless now
	mysql.close()

	current_app.pv = 0   # for debug


# Controllers ========================================================================


@app.route('/')
def hello_world():
	return 'Hello World! %d' % current_app.pv


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
	return (json.dumps({
		"user_id": user.id,
		"username": user.username,
		"access_token": user.token
	}), 200)


@app.route('/foods', methods=['GET'])
def foods_handler():
	auth()
	ret = '[' + ','.join(map(lambda foodid: str(FoodModel.fetch(foodid)), current_app.datapool['FoodModel'])) + ']'
	return ret, 200


@app.route('/carts', methods=['POST'])
def new_carts_handler():
	userid = auth()
	cart = CartModel(userid)
	return '{"cart_id": "%s"}' % cart.id, 200


@app.route('/carts/<cart_id>', methods=['PATCH'])
def carts_handler(cart_id):
	userid = auth()
	data = parse_req_body()
	food_id = int(data.get('food_id', '-1'))
	count = int(data.get('count', '0'))
	cart = CartModel.fetch(cart_id)
	if cart is None:
		abort(404)
	if cart.userid != userid:
		return '{"code": "NOT_AUTHORIZED_TO_ACCESS_CART", "message": "无权限访问指定的篮子"}', 401
	if FoodModel.fetch(food_id) is None:
		return '{"code": "FOOD_NOT_EXISTS", "message": "food not exists"}', 404
	ret = cart.add_food(food_id, count)
	if ret is True:
		return '', 204
	return json.dumps(ret), 403


@app.route('/orders', methods=['POST', 'GET'])
def orders_handler():
	userid = auth()
	if request.method == 'POST':
		# POST
		data = parse_req_body()
		cart_id = data.get('cart_id', '')
		cart = CartModel.fetch(cart_id)
		if (cart is None) or (cart.userid != userid):
			return '{"code": "NOT_AUTHORIZED_TO_ACCESS_CART", "message": "cart not owned by user"}', 401
		try:
			order = OrderModel(cart.id)
		except:
			return '{"code": "ORDER_OUT_OF_LIMIT", "message": "order count exceed maximum limit"}', 403
		return '{"id": "%s"}' % order.id, 200
	else:
		# GET
		orderid = OrderModel.fetch_orderid_by_userid(userid)
		if orderid is None:
			return '[]', 200
		else:
			order = OrderModel.fetch(orderid)
			return '[%s]' % str(order), 200


@app.route('/orders', methods=['GET'])
def admin_orders_handler():
	admin_auth()
	orderids = OrderModel.fetch_all_orderid()
	return dump_orders(orderids), 200


@app.route('/count', methods=['GET'])
def count_handler():
	context = current_app
	count = getattr(context, 'pv', 0)
	context.pv = count + 1
	return str(count), 200


if __name__ == '__main__':
	host = os.getenv("APP_HOST", "localhost")
	port = int(os.getenv("APP_PORT", "8080"))
	app.run(host=host, port=port, debug=True)
