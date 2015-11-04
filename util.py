import json, random, string
from flask import request, abort

from usermodel import UserModel
from ordermodel import OrderModel


def parse_req_body():
	if request.data == '':
		abort(400)
	try:
		return json.loads(request.data)
	except:
		abort(753)
		return dict()


def gen_random_string():
	return ''.join([string.printable[random.randint(0, 61)] for x in xrange(32)])


def auth():
	token = request.args.get('access_token', None)
	if token is None:
		token = request.headers.get('access_token', '')
	userid = UserModel.find_userid_by_token(token)
	if userid is False:
		abort(401)
	return userid


def admin_auth():
	userid = auth()
	if userid != 1:
		abort(401)
	return userid


def dump_orders(order_ids):
	return '[' + ','.join(map(lambda order_id: str(OrderModel.fetch(order_id)), order_ids)) + ']'
