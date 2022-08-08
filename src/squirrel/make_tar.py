#!python

import os
import re

# ----------------------------------
# Definitions:
# ----------------------------------

# Game Script name
gs_name = "AdminCmd"
gs_pack_name = gs_name.replace(" ", "-")

# ----------------------------------


# Script:
version = -1
for line in open("version.nut"):

	r = re.search('SELF_VERSION\s+<-\s+(.+);', line)
	if(r != None):
		version = r.group(1)

if(version == -1):
	print("Couldn't find " + gs_name + " version in version.nut!")
	exit(-1)

dir_name = gs_pack_name + "-v" + version
tar_name = dir_name + ".tar"
os.system("mkdir " + dir_name);
os.system("cp -Ra *.nut lang " + dir_name);
os.system("tar -cf " + tar_name + " " + dir_name);
os.system("rm -r " + dir_name);
os.system("mv " + tar_name + " ../../openttd/game");
