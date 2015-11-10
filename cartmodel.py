#!/usr/bin/env python
# -*- coding: utf-8 -*-
from datamodel import DataModel
from foodmodel import FoodModel


class CartModel(DataModel):
	colmap = ['id', 'userid', 'food_ids', 'food_count', 'total', 'is_locked']

	def __init__(self, userid):
		DataModel.__init__(self)
		self.userid = userid
		self.food_ids = dict()
		self.food_count = 0
		self.total = 0
		self.is_locked = False

	def add_food(self, id, count):
		if self.is_locked:
			return {
				"code": "ORDER_LOCKED",
				"message": "订单已经提交不能修改"
			}
		if self.food_count + count > 3:
			return {
				"code": "FOOD_OUT_OF_LIMIT",
				"message": "篮子中食物数量超过了三个"
			}
		# TODO: lock
		food = FoodModel.fetch(id)
		if food.reserve(count):
			self.food_count += count
			self.total += count * food.price
			if id in self.food_ids.keys():
				self.food_ids[id] += count
			else:
				self.food_ids[id] = count
			return True
		else:
			return {
				"code": "FOOD_OUT_OF_STOCK",
				"message": "食物库存不足"
			}