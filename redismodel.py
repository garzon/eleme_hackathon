import os
from flask import g
import redis


class RedisModel:
	@classmethod
	def get_redis(cls):
		handle = getattr(g, 'redis', None)
		if handle is None:
			redis_addr = os.getenv('REDIS_HOST', 'localhost')
			redis_port = os.getenv('REDIS_PORT', '6379')
			handle = g.redis = redis.Redis(host=redis_addr, port=redis_port, db=0)
		return handle
	
	def __init__(self):
		self.redis = self.get_redis()
		

