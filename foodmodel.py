from mysqlmodel import MysqlModel


class FoodModel(MysqlModel):
	colmap = ['id', 'stock', 'price']

	def __init__(self):
		self.token = None

	def __str__(self):
		return '{"id": %d, "price": %d, "stock": %d}' % (self.id, self.price, self.stock)

	def reserve(self, count):
		if self.stock >= count:
			self.stock -= count
			return True
		else:
			return False
