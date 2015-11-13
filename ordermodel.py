#!/usr/bin/env python
# -*- coding: utf-8 -*-
from flask import current_app

from datamodel import DataModel
from cartmodel import CartModel

from redisstring import RedisString
from redisset import RedisSet

import json


class OrderModel(DataModel):
	OUT_OF_STOCK = 0
	OUT_OF_LIMIT = 1

	def __init__(self, id = None, cartid = None):
		if cartid is None:
			self.prefix = 'ordermodel_obj_string_'
			DataModel.__init__(self, id)
			return
		is_bad_order, userid = CartModel.fetchCols(cartid, ['is_bad_order', 'userid'])
		if is_bad_order == "1":
			raise RuntimeError, OrderModel.OUT_OF_STOCK
		redis_user_order_obj = RedisString('userid2orders_' + userid)
		user_orders = redis_user_order_obj.get()
		if user_orders is None:
			#self.cart.is_locked = True
			self.prefix = 'ordermodel_obj_string_'
			DataModel.__init__(self)
			self.cartid = cartid
			RedisSet('set_order_ids').sadd(self.id)
			redis_user_order_obj.set(self.id)
		else:
			raise RuntimeError, OrderModel.OUT_OF_LIMIT

	def load(self):
		if DataModel.load(self) is False: return False
		self.cartid = self.data_dict['cartid']

	def save(self):
		self.data_dict['cartid'] = self.cartid
		DataModel.save(self)

	@classmethod
	def fetch_orderid_by_userid(cls, userid):
		return RedisString('userid2orders_' + userid).get()

	@classmethod
	def fetch_all_orderid(cls):
		return RedisSet('set_order_ids').smembers()

	@classmethod
	def dump(cls, id):
		redis_obj = RedisString('order_dump_string_' + id)
		dumped = redis_obj.get()
		if dumped is None:
			cart_total, food_ids, food_nums = CartModel.fetchCols(cls.fetchCols(id, 'cartid')[0], ['total', 'food_ids', 'food_nums'])
			items = [{"food_id": int(food_id), "count": int(count)} for food_id, count in zip(food_ids.split(','), food_nums.split(','))]
			ret = {
				"id": id,
				"items": items,
				"total": int(cart_total)
			}
			dumped = json.dumps(ret)
			redis_obj.set(dumped)
		return dumped

