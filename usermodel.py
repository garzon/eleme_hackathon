from mysqlmodel import MysqlModel
from flask import current_app, abort
from redisstring import RedisString

import util, time


class UserModel(MysqlModel):
	colmap = ['id', 'username', 'password']

	def __init__(self):
		self.token = None

	def update_token(self):
		token = util.gen_random_string()
		self.token = token
		RedisString("token2userid_" + token).set(self.id)
		return token

	@classmethod
	def find_userid_by_token(cls, token):
		userid = RedisString("token2userid_" + token).get()
		if userid is None:
			return False
		return userid

	@classmethod
	def init_data_structure(cls):
		current_app.username2userid = dict()

	def after_parse(self):
		current_app.username2userid[self.username] = self.id

	@classmethod
	def login(cls, username, password):
		userid = current_app.username2userid.get(username, None)
		if userid is None: return False
		user = cls.fetch(userid)
		if user.password != password: return False
		user.update_token()
		return user

