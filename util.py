import json, random, string
from flask import request, abort

from usermodel import UserModel


def parse_req_body():
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


def adminAuth():
	user = auth()
	if user != 0:
		abort(401)
	return user
