from flask import current_app
import util


class DataModel:
	# ORM settings
	colmap = []

	# ===============================================================================
	# here are the methods could be overrided
	# ===============================================================================
	@classmethod
	def init_data_structure(cls):
		current_app.datapool[cls.__name__] = dict()

	@classmethod
	def save_to_datapool(cls, obj):
		current_app.datapool[cls.__name__][obj.id] = obj

	def __init__(self):
		self.id = util.gen_random_string()
		self.save_to_datapool(self)

	# ===============================================================================
	# here are the methods shouldn't be overrided, related to the global datapool
	# ===============================================================================
	@classmethod
	def fetch(cls, id):
		return current_app.datapool[cls.__name__].get(str(id), None)
