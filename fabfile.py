# -*- coding: utf-8 -*-

import os

from fabric.api import env, local, run, cd, lcd, sudo, warn_only, prompt

os.environ["__GEN_DOCS__"] = "1"



HERE_PATH =  os.path.abspath( os.path.dirname( __file__ ))



def build_dev():
	"""uses go-bindata to packace assets in -debug mode, ie live to file"""
	local("go-bindata -debug -pkg server -o server/bindata_templates.go templates/")