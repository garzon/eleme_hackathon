#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import MySQLdb

from flask import Flask, abort, current_app
import werkzeug.exceptions as ex

from util import *

from usermodel import UserModel
from foodmodel import FoodModel


app = Flask(__name__)


@app.before_first_request
def initialize_my_app():
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
		cursor.execute("SELECT * from %s" % table_name)
		data = cursor.fetchall()
		for datum in data:
			model_class.parse(datum)

	# connection is useless now
	mysql.close()


@app.route('/')
def hello_world():
	return 'Hello World!'


class SyntaxErrorException(ex.HTTPException):
	code = 753
	description = '{"message": "malformed json"}'


abort.mapping[753] = SyntaxErrorException


@app.errorhandler(753)
def syntax_error(error):
	return SyntaxErrorException.description, 753


@app.errorhandler(401)
def syntax_error(error):
	return '{"code": "INVALID_ACCESS_TOKEN", "message": "无效的令牌"}', 401


@app.route('/login', methods=['POST'])
def login_handler():
	fail = (json.dumps({
		"code": "USER_AUTH_FAIL",
		"message": 'username or password incorrect'
	}), 403)

	data = parse_req_body()
	try:
		username = data['username']
		password = data['password']
	except: return fail

	user = UserModel.login(username, password)
	if user is False: return fail
	return (json.dumps({
		"user_id": user.id,
		"username": user.username,
		"access_token": user.token
	}), 200)


@app.route('/foods', methods=['GET'])
def foods_handler():
	auth()
	ret = '[' + ','.join(map(lambda foodid: str(FoodModel.fetch(foodid)), current_app.datapool['FoodModel'])) + ']'
	return (ret, 200)


@app.route('/carts/<cart_id>', methods=['POST', 'PATCH'])
def carts_handler(cart_id):
	auth()


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
