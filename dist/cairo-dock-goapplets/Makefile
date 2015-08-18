
# This Makefile is designed to be moved at the root of the GOPATH.
#
# getsrc, packsrc:         Download/update/patch and package the sources.
# build ,install, clean:   Default build actions for the Debian autobuild.
# 

SOURCE=github.com/sqp/godock

TAGS=gtk gtk_3_10 # limit gtk version for trusty.
APPLETS=Audio Cpu DiskActivity DiskFree GoGmail Mem NetActivity Notifications TVPlay Update


export GOPATH=$(CURDIR)

APPNAME=cairo-dock-goapplets

APPVERSION=$(shell cat src/$(SOURCE)/version)

SUBDIR=$(APPNAME)-$(APPVERSION)
SRCFILE=$(APPNAME)_$(APPVERSION).orig.tar.gz


# Forward default make actions to the real makefile lost in the GOPATH.
# This enables the Debian autobuild system.

%: build

build:
	GOPATH=$(CURDIR) make --directory=src/$(SOURCE) APPLETS="$(TAGS) $(APPLETS)" build

install:
	GOPATH=$(CURDIR) make --directory=src/$(SOURCE) PKGDIR="$(DESTDIR)" APPLETS="$(APPLETS)" install
	# PKGDIR="$(GOPATH)/$(PKGDIR)"

clean:
	rm -rf bin pkg


# Get sources and apply patches.
getsrc:
	go get -u -d -tags '$(APPLETS)' $(SOURCE)/cmd/cdc
	
	make --directory="$(CURDIR)/src/$(SOURCE)" patch
	

# Create a sources package to use as Debian or OpenSuse Build Service source.
packsrc:
	mkdir $(SUBDIR)

	# Move Debian install files and go sources to root path.

	mv src/$(SOURCE)/dist/$(APPNAME)/* src $(SUBDIR)

	# Compress.
	
	tar -zcf $(SRCFILE) $(SUBDIR) --exclude-vcs
	
	@# Display results

	@echo
	@echo ------------
	@echo
	@ls -l $(SRCFILE)
	
	@md5sum $(SRCFILE)
	@sha256sum $(SRCFILE)