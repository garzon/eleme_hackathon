from redismodel import RedisModel


class RedisHash(RedisModel):
	def __init__(self, key):
		RedisModel.__init__(self)
		self.key = key

	def hmset(self, data_dict):
		return self.redis.hmset(self.key, data_dict)

	def hmget(self, data_cols):
		return self.redis.hmget(self.key, data_cols)

	def hgetall(self):
		return self.redis.hgetall(self.key)