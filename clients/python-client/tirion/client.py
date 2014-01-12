#!/usr/bin/python
# -*- coding: utf-8 -*-

'''
Description: Tirion client
'''

import numpy
import os
import socket
import sys
import threading

from __init__ import __version__

class Client:
	"""Tirion client"""

	__LOG_PREFIX = "[client]"
	__TIRION_BUFFER_SIZE = 4096
	__TIRION_TAG_SIZE = 513

	def __init__(self, socket_filename, verbose):
		"""Tirion client constructor

		@param socket_filename the socket filepath to connect to the agent
		@param verbose enable or disable verbose output of the client library
		"""

		self.__command_thread = None
		self.__count = 0
		self.__metrics = None
		self.__metric_lock = None
		self.__net = None
		self.__running = False
		self.__socket = socket_filename
		self.__verbose_output = verbose

	def init(self):
		"""Initialize a Tirion client object"""

		os.setsid()

		self.verbose("Open unix socket to {}", self.__socket)
		self.__net = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
		try:
			self.__net.connect(self.__socket)
		except socket.error as err:
			raise RuntimeError("Cannot initialize socket: " + str(err))

		self.verbose("Request tirion protocol version v{}", __version__)
		self.__send("tirion v" + __version__ + "\tmmap")

		header = self.__receive().split("\t")

		if len(header) < 2 or len(header[1]) == 0:
			raise RuntimeError("Did not receive correct metric count and mmap filename")

		try:
			self.__count = int(header[0])
		except ValueError:
			self.error("Did not receive correct metric count")

			raise RuntimeError("Metric count is not a number")

		self.__metric_lock = threading.Lock()

		if not header[1].startswith("mmap://"):
			raise RuntimeError("Did not receive correct mmap filename")

		mmap_filename = header[1][7:]

		self.verbose("Received metric count {} and mmap filename {}", self.__count, mmap_filename)

		self.__metrics = numpy.memmap(mmap_filename, dtype='float32', mode='r+', shape=(self.__count,))

		self.verbose("Initialized metric collector mmap")

		self.__running = True

		# we want to handle commands not in the main thread
		self.__command_thread = threading.Thread(target=Client.__handle_commands, args=(self,))
		self.__command_thread.start()

	def close(self):
		"""Uninitialized a Tirion client object"""
		self.__running = False

		if self.__metric_lock is not None:
			self.__metric_lock.acquire()

		if self.__metrics is not None:
			# self.__metrics.close()
			self.__metrics = None

		if self.__net is not None:
			self.__net.shutdown(socket.SHUT_RDWR)
			self.__net.close()
			self.__net = None

		if self.__command_thread is not None:
			self.__command_thread.join()
			self.__command_thread = None

		if self.__metric_lock is not None:
			self.__metric_lock.release()

	def destroy(self):
		"""Cleanup everything that was allocated by the Tirion client object"""

		self.__metric_lock = None

	def get(self, index):
		"""Return the current value of a metric

		@param index the index of the metric

		@return the value of the metric
		"""

		if index < 0 or index >= self.__count or self.__metric_lock is None or self.__metrics is None:
			return 0.0

		return self.__metrics[index]

	def set(self, index, value):
		"""Set a value for a metric

		@param index the index of the metric
		@param value the value to be set to the metric

		@return the new value of the metric
		"""

		if index < 0 or index >= self.__count or self.__metric_lock is None:
			return 0.0

		ret = 0.0

		self.__metric_lock.acquire()

		if self.__metrics is not None:
			ret = value
			self.__metrics[index] = ret

		self.__metric_lock.release()

		return ret

	def add(self, index, value):
		"""Add a value to a metric

		@param index the index of the metric
		@param value the value to be add to the metric

		@return the new value of the metric
		"""

		if index < 0 or index >= self.__count or self.__metric_lock is None:
			return 0.0

		ret = 0.0

		self.__metric_lock.acquire()

		if self.__metrics is not None:
			ret = self.__metrics[index] + value
			self.__metrics[index] = ret

		self.__metric_lock.release()

		return ret

	def dec(self, index):
		"""Decrement a metric by 1.0

		@param index the index of the metric

		@return the new value of the metric
		"""

		return self.add(index, -1.0)

	def inc(self, index):
		"""Increment a metric by 1.0

		@param index the index of the metric

		@return the new value of the metric
		"""

		return self.add(index, 1.0)

	def sub(self, index, value):
		"""Subtract a value of a metric

		@param index the index of the metric
		@param value the value to be subtracted of the metric

		@return the new value of the metric
		"""

		return self.add(index, -value)

	def running(self):
		"""States if the Tirion Client object is running

		@return running state
		"""

		return self.__running

	def tag(self, format_string, *args):
		"""Send a tag to the agent

		@param format_string the tag string that follows the same specifications as format_string in string.format
		@param args additional arguments for format_string
		"""

		self.__send(self.__prepare_tag("t" + format_string.format(*args)))

	def __message(self, message_type, format_string, *args):
		"""Output a Tirion message"""

		if not self.__verbose_output:
			return

		sys.stderr.write(Client.__LOG_PREFIX + "[" + message_type + "] " + format_string.format(*args) + "\n")

	def debug(self, format_string, *args):
		"""Output a Tirion debug message

		@param format_string the message string that follows the same specifications as format_string in string.format
		@param args additional arguments for format_string
		"""

		self.__message("debug", format_string, *args)

	def error(self, format_string, *args):
		"""Output a Tirion error message

		@param format_string the message string that follows the same specifications as format_string in string.format
		@param args additional arguments for format_string
		"""

		self.__message("error", format_string, *args)

	def verbose(self, format_string, *args):
		"""Output a Tirion verbose message

		@param format_string the message string that follows the same specifications as format_string in string.format
		@param args additional arguments for format_string
		"""

		self.__message("verbose", format_string, *args)

	def __prepare_tag(self, tag):
		"""Prepare a tag string for sending"""

		if len(tag) > self.__TIRION_TAG_SIZE:
			tag = tag[:self.__TIRION_TAG_SIZE]

		return tag.replace("\n", " ")

	def __receive(self):
		"""Receive a message over the unix socket"""

		msg = self.__net.recv(self.__TIRION_BUFFER_SIZE)

		if msg == '':
			raise RuntimeError("socket connection broken")

		return msg.strip()

	def __send(self, msg):
		"""Sent a message over the unix socket"""

		msg_len = len(msg)
		total = 0

		while total < msg_len:
			sent = self.__net.send(msg[total:])

			if sent == 0:
				raise RuntimeError("socket connection broken")

			total = total + sent

	def __handle_commands(self):
		"""Handle commands received from the agent"""

		self.verbose("Start listening to commands")

		while self.__running:
			try:
				rec = self.__receive()

				com = rec[0]

				self.error("Unknown command '{}'", com)
			except RuntimeError as err:
				self.error("Unix socket error: {}", err)

				self.__running = False

		self.verbose("Stop listening to commands")
