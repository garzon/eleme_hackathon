from redishash import RedisHash

import util

class DataModel:
	def __init__(self, id = None):
		if id is None: id = util.gen_random_string()
		self.id = id
		self.data_dict = {'id': id}
		self.redis_hash = RedisHash(self.prefix + id)

	def save(self):
		self.data_dict['id'] = self.id
		self.redis_hash.hmset(self.data_dict)

	def load(self):
		self.data_dict = self.redis_hash.hgetall()
		self.id = self.data_dict.get('id', None)
		return not (self.id is None)

	@classmethod
	def fetch(cls, id):
		ret = cls(id)
		if ret.load() is False: return None
		return ret

	@classmethod
	def fetchCols(cls, id, cols):
		return cls(id).redis_hash.hmget(cols)