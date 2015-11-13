'''from app import app
import os
from gevent.baseserver import _tcp_listener
from gevent import pywsgi
from gevent.monkey import patch_all; patch_all()
from multiprocessing import Process, current_process, cpu_count

host = os.getenv("APP_HOST", "localhost")
port = int(os.getenv("APP_PORT", "8080"))
#http_server = WSGIServer((host, port), )
#http_server.serve_forever()

listener = _tcp_listener((host, port))

def serve_forever(listener, app):
	pywsgi.WSGIServer(listener, app).serve_forever()

number_of_processes = 4
print 'Starting %s processes' % number_of_processes
for i in xrange(number_of_processes):
	Process(target=serve_forever, args=(listener, app)).start()

serve_forever(listener, app)'''

import sys
import os
from app import app
import gevent
import gevent.monkey
import gevent.wsgi
import gevent.server
gevent.monkey.patch_all()

import multiprocessing
host = os.getenv("APP_HOST", "localhost")
port = int(os.getenv("APP_PORT", "8080"))

if __name__ == '__main__':
	listener = gevent.server._tcp_listener((host, port), backlog=500, reuse_addr=True)
	for i in xrange(multiprocessing.cpu_count()*2):
		server = gevent.wsgi.WSGIServer(listener, app, log=None)
		process = multiprocessing.Process(target=server.serve_forever)
		process.start()
