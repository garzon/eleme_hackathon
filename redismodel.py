import os
from flask import current_app
import redis


class RedisModel:
	@classmethod
	def flushall(cls):
		redis_addr = os.getenv('REDIS_HOST', 'localhost')
		redis_port = os.getenv('REDIS_PORT', '6379')
		handle = redis.Redis(host=redis_addr, port=redis_port, db=0)
		return handle.flushall()

	@classmethod
	def get_redis(cls, force_reconnect = False):
		handle = current_app.redis
		if handle is None:
			redis_addr = os.getenv('REDIS_HOST', 'localhost')
			redis_port = os.getenv('REDIS_PORT', '6379')
			handle = current_app.redis = redis.Redis(host=redis_addr, port=redis_port, db=0)
		return handle
	
	def __init__(self):
		self.redis = self.get_redis()


