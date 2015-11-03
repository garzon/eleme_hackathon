from mysqlmodel import MysqlModel
from flask import current_app

import util


class UserModel(MysqlModel):
	colmap = ['id', 'username', 'password']

	def __init__(self):
		self.token = None

	def update_token(self):
		if self.token is not None: del current_app.token2userid[self.token]
		token = util.gen_random_string()
		self.token = token
		current_app.token2userid[token] = self.id
		return token

	@classmethod
	def find_userid_by_token(cls, token):
		# find who is it.
		try:
			userid = current_app.token2userid[token]
		except:
			return False
		return userid

	@classmethod
	def init_data_structure(cls):
		current_app.token2userid = dict()
		current_app.username2userid = dict()

	def after_parse(self):
		current_app.username2userid[self.username] = self.id

	@classmethod
	def login(cls, username, password):
		try:
			userid = current_app.username2userid[username]
			user = cls.fetch(userid)
			if user.password != password: return False
			user.update_token()
			return user
		except:
			return False