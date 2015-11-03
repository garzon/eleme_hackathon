from redisstring import RedisString


class CartModel:
	cart2content_prefix = "cart_%s_content"
	cart2user_prefix = "cart_%s_user"

	@classmethod
	def setToken(cls, uid, token):
		RedisString(cls.user2token_prefix % uid).set(token)
		RedisString(cls.token2user_prefix % token).set(uid)

	# create a cart
	def __init__(self):


	@classmethod
	def fetchCart(cls):
		content = RedisString(cls.user2token_prefix % uid).get(token)

	@classmethod
	def verifyToken(cls, token):
		# find who is it.
		reqUser = RedisString(cls.token2user_prefix % token).get()
		# prevent double tokens
		if reqUser and RedisString(cls.user2token_prefix % reqUser).get() == token:
			return reqUser
		else:
			return False

