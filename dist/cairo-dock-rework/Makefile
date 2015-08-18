
# This Makefile is designed to be moved at the root of the GOPATH.
#
# getsrc, packsrc:         Download/update/patch and package the sources.
# build ,install, clean:   Default build actions for the Debian autobuild.
# 

SOURCE=github.com/sqp/godock

TAGS=dock gtk_3_10 # limit gtk version for trusty.
APPLETS=Audio Cpu DiskActivity DiskFree GoGmail Mem NetActivity Notifications TVPlay Update


export GOPATH=$(CURDIR)

APPNAME=cairo-dock-rework

APPVERSION=$(shell cat src/$(SOURCE)/version)

SUBDIR=$(APPNAME)-$(APPVERSION)
SRCFILE=$(APPNAME)_$(APPVERSION).orig.tar.gz


# Forward default make actions to the real makefile lost in the GOPATH.
# This enables the Debian autobuild system.

%: build

build:
	GOPATH=$(CURDIR) make --directory=src/$(SOURCE) DOCK="$(TAGS) $(APPLETS)" dock

install:
	GOPATH=$(CURDIR) make --directory=src/$(SOURCE) PKGDIR="$(DESTDIR)" APPLETS="$(APPLETS)" install-dock

clean:
	rm -rf bin pkg


# Get sources and apply patches.
getsrc:
	go get -u -d -tags 'dock all' $(SOURCE)/cmd/cdc
	
	make --directory="$(CURDIR)/src/$(SOURCE)" patch-dock
	

# Create a sources package to use as Debian or OpenSuse Build Service source.
packsrc:
	# Move Debian install files to root path.

	mv src/$(SOURCE)/dist/$(APPNAME)/* src

	# Rename src dir and compress.
	
	mv src $(SUBDIR)
	
	tar -zcf $(SRCFILE) $(SUBDIR) --exclude-vcs
	
	@# Display results

	@echo
	@echo ------------
	@echo
	@ls -l ../$(SRCFILE)
	
	@md5sum ../$(SRCFILE)
	@sha256sum ../$(SRCFILE)