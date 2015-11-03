#!/usr/bin/env python
# -*- coding: utf-8 -*-
from datamodel import DataModel
from foodmodel import FoodModel


class CartModel(DataModel):
	colmap = ['id', 'userid', 'food_ids', 'food_count']

	def __init__(self, userid):
		DataModel.__init__(self)
		self.userid = userid
		self.food_ids = dict()
		self.food_count = 0

	def add_food(self, id, count):
		if self.food_count + count > 3:
			return {
				"code": "FOOD_OUT_OF_LIMIT",
				"message": "篮子中食物数量超过了三个"
			}
		# TODO: lock
		if FoodModel.fetch(id).reserve(count):
			self.food_count += count
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