from redisstring import RedisString
from mysqlmodel import MysqlModel
from flask import current_app


class FoodModel(MysqlModel):
	colmap = ['id', 'stock', 'price']

	def __init__(self):
		self.token = None

	def __str__(self):
		return '{"id": %d, "price": %d, "stock": %d}' % (self.id, self.price, self.stock)