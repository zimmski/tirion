#!/usr/bin/python
# -*- coding: utf-8 -*-

from distutils.core import setup

import tirion

setup(
	name='Tirion',
	version=tirion.__version__,
	description='Tirion Python client',
	author='Markus Zimmermann',
	author_email='mz@nethead.at',
	url='https://github.com/zimmski/tirion',
	packages=['tirion'],
	requires=['numpy'],
)
