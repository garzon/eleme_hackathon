from redismodel import RedisModel


class RedisSet(RedisModel):
	def __init__(self, key):
		RedisModel.__init__(self)
		self.key = key

	def sadd(self, val):
		return self.redis.sadd(self.key, val)

	def smembers(self):
		return list(self.redis.smembers(self.key))