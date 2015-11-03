from flask import current_app


class MysqlModel:
	# ORM settings
	colmap = []

	# ===============================================================================
	# here are the methods could be overrided
	# ===============================================================================
	def after_parse(self):
		pass

	@classmethod
	def init_data_structure(cls):
		pass

	# ===============================================================================
	# here are the methods shouldn't be overrided, related to the global datapool
	# ===============================================================================
	@classmethod
	def parse(cls, raw_data):
		'''
			@raw_data raw_data return by fetchone(), raw_data[0] must be primary key(id)
		'''
		obj = cls()
		for i, v in enumerate(raw_data):
			setattr(obj, cls.colmap[i], v)
		current_app.datapool[cls.__name__][str(raw_data[0])] = obj
		obj.after_parse()
		return obj

	@classmethod
	def fetch(cls, id):
		# TODO: load if the corresponding obj is not in the datapool
		return current_app.datapool[cls.__name__][str(id)]
