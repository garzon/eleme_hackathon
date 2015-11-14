#!/usr/bin/env python
# -*- coding: utf-8 -*-
from datamodel import DataModel
from foodmodel import FoodModel


class CartModel(DataModel):
	def __init__(self, id = None):
		self.prefix = 'cartmodel_obj_string_'
		DataModel.__init__(self, id)
		self.food_ids = dict()
		self.food_count = 0
		self.total = 0
		self.is_locked = False
		self.is_bad_order = False

	def load(self):
		if DataModel.load(self) is False: return False
		data_dict = self.data_dict
		tmp = zip(data_dict['food_ids'].split(','), data_dict['food_nums'].split(','))
		self.food_ids = {x: int(y) for x, y in tmp if y != ''}
		self.food_count = int(data_dict['food_count'])
		self.total = int(data_dict['total'])
		self.is_locked = (data_dict['is_locked'] == "1")
		self.is_bad_order = (data_dict['is_bad_order'] == "1")
		self.userid = data_dict['userid']

	def save(self):
		data_dict = self.data_dict
		data_dict['food_count'] = self.food_count
		data_dict['total'] = self.total
		data_dict['is_locked'] = "1" if self.is_locked else "0"
		data_dict['is_bad_order'] = "1" if self.is_bad_order else "0"
		data_dict['userid'] = self.userid
		data_dict['food_ids'] = ','.join(map(lambda i: str(i), self.food_ids.keys()))
		data_dict['food_nums'] = ','.join(map(lambda i: str(i), self.food_ids.values()))
		DataModel.save(self)

	def add_food(self, id, count):
		if self.is_locked:
			return '''{
				"code": "ORDER_LOCKED",
				"message": "订单已经提交不能修改"
			}'''
		if self.food_count + count > 3:
			return '''{
				"code": "FOOD_OUT_OF_LIMIT",
				"message": "篮子中食物数量超过了三个"
			}'''
		# TODO: lock
		food = FoodModel.fetch(id)
		if food.reserve(count):
			self.food_count += count
			self.total += count * food.price
			try:
				self.food_ids[id] += count
			except KeyError:
				self.food_ids[id] = count
		else:
			self.is_bad_order = True
		self.save()
		return True

