from redismodel import RedisModel


class RedisString(RedisModel):
	def __init__(self, key):
		RedisModel.__init__(self)
		self.key = key
	
	def get(self):
		return self.redis.get(self.key)

	def set(self, val):
		return self.redis.set(self.key, val)

	def delete(self, val):
		return self.redis.delete(self.key)