from mysqlmodel import MysqlModel
from redisstring import RedisString

from flask import current_app


class FoodModel(MysqlModel):
	colmap = ['id', 'stock', 'price']

	def __init__(self):
		self.token = None

	@classmethod
	def init_data_structure(cls):
		current_app.food_ids_arr = []
		current_app.food_template_str = []

	def after_parse(self):
		RedisString("food_stock_of_" + str(self.id)).set(self.stock)
		current_app.food_ids_arr += [str(self.id)]
		current_app.food_template_str += ['{"id": ' + str(self.id) + ', "price": ' + str(self.price) + ', "stock": %s}']

	def reserve(self, count):
		redisObj = RedisString("food_stock_of_" + str(self.id))
		if int(redisObj.get()) >= count:
			redisObj.decrBy(count)
			return True
		else:
			return False
