from mysqlmodel import MysqlModel
from redisstring import RedisString


class FoodModel(MysqlModel):
	colmap = ['id', 'stock', 'price']

	def __init__(self):
		self.token = None

	def __str__(self):
		return '{"id": %d, "price": %d, "stock": %s}' % (self.id, self.price, RedisString("food_stock_of_" + str(self.id)).get())

	def after_parse(self):
		RedisString("food_stock_of_" + str(self.id)).set(self.stock)

	def reserve(self, count):
		redisObj = RedisString("food_stock_of_" + str(self.id))
		if int(redisObj.get()) >= count:
			redisObj.decrBy(count)
			return True
		else:
			return False
