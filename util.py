import json, random, string
from flask import request, abort, current_app

from collections import deque

from usermodel import UserModel


def parse_req_body():
	if request.data == '':
		abort(400)
	try:
		return json.loads(request.data)
	except:
		abort(753)
		return dict()


def gen_random_string_pool():
	randint = random.randint
	printable = string.printable
	current_app.random_string_pool = deque([''.join([printable[randint(0, 61)] for _ in xrange(32)]) for _ in xrange(50000)])


def gen_random_string():
	randint = random.randint
	printable = string.printable
	#try:
	#	return current_app.random_string_pool.pop()
	#except IndexError:
	#	gen_random_string_pool()
	#	return current_app.random_string_pool.pop()
	return ''.join([printable[randint(0, 61)] for _ in xrange(32)])


def auth():
	token = request.args.get('access_token', None) or request.headers.get('access_token', '')
	userid = UserModel.find_userid_by_token(token)
	if userid is False:
		abort(401)
	return userid


def admin_auth():
	userid = auth()
	if str(userid) != "1":
		abort(401)
	return userid
