#!/usr/bin/python
# -*- coding: utf-8 -*-

'''
Description: Tirion python example client
'''

import argparse
import threading
import sys
import time

import tirion.client

def main():
	'''The main function'''

	runtime = 5
	socket = "/tmp/tirion.sock"

	parser = argparse.ArgumentParser(description="Tirion Java example client v%s" % (tirion.__version__))

	parser.add_argument("-r", "--runtime", nargs=1, type=int, default=runtime, help="Runtime of the example client in seconds. Default is %d." % (runtime))
	parser.add_argument("-s", "--socket", nargs=1, default=socket, help="Unix socket path for client<-->agent communication. Default is %s" % (socket))
	parser.add_argument("-v", "--verbose", help="Enable verbose output.")

	args = parser.parse_args()

	tirion_client = tirion.client.Client(args.socket, True if args.verbose else False)

	try:
		tirion_client.init()
	except (RuntimeError, TypeError) as err:
		sys.stderr.write("[ERROR] Cannot initialize Tirion: " + str(err) + "\n")

		sys.exit(1)

	threading.Thread(target=__after, args=(tirion_client, args.runtime,)).start()

	while tirion_client.running():
		ret = tirion_client.inc(0)
		tirion_client.dec(1)
		tirion_client.add(2, 0.3)
		tirion_client.sub(3, 0.3)

		time.sleep(0.01)

		if ret % 20.0 == 0.0:
			tirion_client.tag("index 0 is {}", ret)

	tirion_client.close()

	tirion_client.verbose("Stopped")

	tirion_client.destroy()

def __after(tirion_client, runtime):
	"""this function stops the Tirion client after x seconds"""

	time.sleep(runtime)

	tirion_client.debug("Program ran for {} seconds, this is enough data.", runtime)

	tirion_client.close()

if __name__ == '__main__':
	main()
