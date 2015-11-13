import multiprocessing, os

workers = multiprocessing.cpu_count()
host = os.getenv("APP_HOST", "localhost")
port = os.getenv("APP_PORT", "8080")
bind = "%s:%s" % (host, port)
worker_class = 'gevent'
worker_connections = 1000
