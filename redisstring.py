from redismodel import RedisModel


class RedisString(RedisModel):
	def __init__(self, key):
		RedisModel.__init__(self)
		self.key = key
	
	def get(self):
		return self.redis.get(self.key)

	def set(self, val):
		return self.redis.set(self.key, val)

	def setex(self, val, expire):
		return self.redis.setex(self.key, val, expire)

	def psetex(self, val, expire):
		return self.redis.psetex(self.key, expire, val)

	def delete(self, val):
		return self.redis.delete(self.key)

	def incrBy(self, val):
		return self.redis.incr(self.key, val)

	def decrBy(self, val):
		return self.redis.decr(self.key, val)
