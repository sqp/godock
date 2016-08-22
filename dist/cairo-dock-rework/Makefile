
# This Makefile is designed to be moved at the root of the GOPATH.
#
# getsrc packsrc packdeb   Download/update/patch and package the sources.
# build install clean      Default build actions for the Debian autobuild.
# 

SOURCE=github.com/sqp/godock

TAGS=dock gtk_3_10 # limit gtk version for trusty.
APPLETS=Audio Clouds Cpu DiskActivity DiskFree GoGmail Mem NetActivity Notifications TVPlay Update


# export GOPATH=$(CURDIR)

APPNAME=cairo-dock-rework

APPVERSION=0.0.3.5

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
	# This should be enabled, but seem too dangerous.
	# rm -rf bin pkg


# Get sources and apply patches.
# Result: subdir src with go sources. Current dir usable as GOPATH.
getsrc:
	# Download and patch.
	GOPATH=$(CURDIR) go get -u -d -tags 'dock $(APPLETS)' $(SOURCE)/cmd/cdc
	
	GOPATH=$(CURDIR) make --directory="$(CURDIR)/src/$(SOURCE)" patch-dock


# Create a sources package to use as Debian or OpenSuse Build Service source.
# Result: archive name_version.orig.tar.gz (removes sources dir).
packsrc:
	 # Ensure the dest dir doesn't exist.
	mkdir $(SUBDIR)

	# Move Debian install files and go sources to root path.
	mv src/$(SOURCE)/dist/$(APPNAME)/* src $(SUBDIR)

	# Compress.
	tar --exclude-vcs -zcf $(SRCFILE) $(SUBDIR)

	# Clean sources.
	rm -rf $(SUBDIR)


# Create Debian package.
# Result: package name_version_sysarch.deb (and others) without removing the archive.
packdeb:
	 # Ensure the dest dir doesn't exist.
	mkdir $(SUBDIR)

	tar xvf $(SRCFILE)

	cd $(SUBDIR) && dpkg-buildpackage -us -uc -tc
	rm -rf $(SUBDIR)
