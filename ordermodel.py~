#!/usr/bin/env python
# -*- coding: utf-8 -*-
from flask import current_app

from datamodel import DataModel
from cartmodel import CartModel

import json


class OrderModel(DataModel):
	colmap = ['id', 'cartid']

	@classmethod
	def init_data_structure(cls):
		current_app.datapool[cls.__name__] = dict()
		current_app.userid2orders = dict()

	def __init__(self, cartid):
		self.cart = CartModel.fetch(cartid)
		if self.cart.is_bad_order == True:
			raise Exception, "outofstock"
		user_orders = current_app.userid2orders.get(self.cart.userid, None)
		if user_orders is None:
			self.cart.is_locked = True
			DataModel.__init__(self)
			self.cartid = cartid
			current_app.userid2orders[self.cart.userid] = self.id
		else:
			raise Exception, "outoflimit"

	@classmethod
	def fetch_orderid_by_userid(cls, userid):
		return current_app.userid2orders.get(userid, None)

	@classmethod
	def fetch_all_orderid(cls):
		return current_app.userid2orders.values()

	def __str__(self):
		items = [{"food_id": int(food_id), "count": count} for food_id, count in self.cart.food_ids.items()]
		ret = {
			"id": self.id,
			"items": items,
			"total": self.cart.total
		}
		return json.dumps(ret)
