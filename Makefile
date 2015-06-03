TARGET=cdc
SOURCE=github.com/sqp/godock
VERSION=0.0.3-1
APPLETS=Audio Cpu DiskActivity DiskFree GoGmail Mem NetActivity Update

# unstable applets requires uncommited patches to build.
UNSTABLE=Notifications TVPlay
UNSTABLE_TAGS=gtk

# and dock even more, plus the rewritten dock.
DOCK=dock all

# Install prefix if any.
PKGDIR=

APPDIRGLDI=usr/share/cairo-dock/plug-ins/goapplets/
APPDIRDBUS=usr/share/cairo-dock/plug-ins/Dbus/third-party/


%: build

archive-%:
	go build -tags '$(APPLETS)'  -o applets/$(TARGET) $(SOURCE)/cmd/$(TARGET)
	@echo "Make archive $(TARGET)-$(VERSION)-$*.tar.xz"
	tar cJfv $(TARGET)-$(VERSION)-$*.tar.xz applets  --exclude-vcs
	rm applets/$(TARGET)

build:
	go install -tags '$(APPLETS)' $(SOURCE)/cmd/$(TARGET)

unstable:
	go install -tags '$(APPLETS) $(UNSTABLE) $(UNSTABLE_TAGS)' $(SOURCE)/cmd/$(TARGET)

dock:
	go install -tags '$(DOCK)' $(SOURCE)/cmd/$(TARGET)

patch:
	# Patch GTK - unstable branch is required to build a dock.
	cd "$(GOPATH)/src/github.com/conformal/gotk3" && git pull --commit --no-edit https://github.com/sqp/gotk3 unstable
	
	# git pull --commit --no-edit https://github.com/sqp/gotk3 current # current branch is enough for applets.

	# branches grouped in the tree:
	# https://github.com/sqp/gotk3 nil_case file-chooser scale treestore icontheme pixbuf_at_scale gliblist others widget_set liststore_hack
	# iconview link_font_button expander cellrendererpixbuf accelerator
	# https://github.com/shish/gotk3 file-chooser
	# https://github.com/shish/gotk3 paned

	# Patch Dbus (for Notifications)
	cd "$(GOPATH)/src/github.com/godbus/dbus" && git pull --commit --no-edit https://github.com/sqp/dbus fixeavesdrop

install:
	mkdir -p "$(PKGDIR)/usr/bin"
	install -p -m755 "$(GOPATH)/bin/cdc" "$(PKGDIR)/usr/bin"

	mkdir -p "$(PKGDIR)/$(APPDIRGLDI)"
	for f in $(APPLETS); do	\
		cp -Rv --preserve=timestamps "applets/$$f" "$(PKGDIR)/$(APPDIRGLDI)" ;\
		rm $(PKGDIR)/$(APPDIRGLDI)/$$f/$$f ;\
		rm $(PKGDIR)/$(APPDIRGLDI)/$$f/applet.go ;\
		rm $(PKGDIR)/$(APPDIRGLDI)/$$f/last-modif ;\
		rm $(PKGDIR)/$(APPDIRGLDI)/$$f/Makefile ;\
		rm $(PKGDIR)/$(APPDIRGLDI)/$$f/tocdc ;\
	done


	# Package license (if available)
	# for f in LICENSE COPYING LICENSE.* COPYING.*; do
	# 	if [ -e "$(GOPATH)/src/$(SOURCE)/$f" ]; then
	# 		install -Dm644 "$(GOPATH)/src/$(SOURCE)/$f" "$(PKGDIR)/usr/share/licenses/$(TARGET)/$f"
	# 	fi
	# done


install-dbus:
	mkdir -p "$(PKGDIR)/usr/bin"
	install -p -m755 "$(GOPATH)/bin/cdc" "$(PKGDIR)/usr/bin"

	mkdir -p "$(PKGDIR)/$(APPDIRDBUS)"
	for f in $(APPLETS); do	\
		cp -Rv --preserve=timestamps "applets/$$f" "$(PKGDIR)/$(APPDIRDBUS)" ;\
		rm $(PKGDIR)/$(APPDIRDBUS)/$$f/applet.go ;\
		rm $(PKGDIR)/$(APPDIRDBUS)/$$f/last-modif ;\
		rm $(PKGDIR)/$(APPDIRDBUS)/$$f/Makefile ;\
	done
