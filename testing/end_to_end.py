import subprocess
import sys
import json
import random
import io
import time

LOG_BASE='/var/loghub/'
LOGS_COUNT = 10
BASE_PORT = 10000
STAT_PORT = 9999

#Starting logs.
log_procs = []
log_lim = 1

for i in range(1, LOGS_COUNT+1):
	log_proc = subprocess.Popen([
		'loghub', 'log', 
		'-debug',
		'-listen', 'localhost:' + str(BASE_PORT + i), 
		'-home', LOG_BASE + 'log' + str(i),
		'-cert', 'cert.pem',
		'-key', 'testkey.pem',
		'-tls', '-trust',
		'-hub', ':' + str(STAT_PORT),
		'-lim', str(log_lim)])

	log_procs.append(log_proc)
	log_lim *= 2

time.sleep(1)

#Starting hub.
hub_proc = subprocess.Popen([
	'loghub', 'hub', 
	'-debug',
	'-home', LOG_BASE + 'hub',
	'-cert', 'cert.pem',
	'-key', 'testkey.pem',
	'-tls', '-trust',
	'-listen', ':' + str(BASE_PORT), 
	'-stat', ':' + str(STAT_PORT)])

time.sleep(1)

#Starting ui.
ui_proc = subprocess.Popen([
	'loghub', 'ui', 
	'-debug',
	'-addr', 'localhost:' + str(BASE_PORT),
	'-cert', 'cert.pem',
	'-key', 'testkey.pem',
	'-tls', '-trust'])

def write_some_log(log_proc_num, min_cnt, max_cnt):
	writer_proc = subprocess.Popen([
		'loghub', 'put', 
		'-tls', '-trust',
		'-addr', ':' + str(BASE_PORT + log_proc_num)], stdin=subprocess.PIPE)

	for i in range(0, random.randint(min_cnt, max_cnt)):
		msg = io.StringIO()

		for j in range(0, random.randint(1, 100)):
			msg.write(u'Word' + unicode(j) + u' ')

		ent = {
			'Sev': random.randint(0, 10),
			'Src': 'LogSource' + str(random.randint(1, 100)),
			'Msg': msg.getvalue()
		}

		ent_json = json.dumps(ent)

		writer_proc.stdin.write(unicode(ent_json))
		writer_proc.stdin.write(u'\n')

	writer_proc.stdin.close()
	writer_proc.wait()

try:
	#Waiting for input.
	if sys.stdin.readline() != 'q\n':
		while True:
			#Write some logs.
			for i in range(1, len(log_procs) + 1):
				write_some_log(i, 1, 10000)

			#Getting stats.
			subprocess.call([
				'loghub', 'stat', 
				'-tls', '-trust',
				'-u', 'admin',
				'-addr', ':' + str(BASE_PORT)])

			#Waiting for input.
			if sys.stdin.readline()  == 'q\n':
				break
finally:
	#Terminating hub.
	hub_proc.terminate()

	#Terminating logs.
	for p in log_procs:
		p.terminate()

	#Terminating ui.
	ui_proc.terminate()
