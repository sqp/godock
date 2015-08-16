TARGET=cdc
SOURCE=github.com/sqp/godock
VERSION=0.0.3.2
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





# Could be useful for some distro packagers.
# FLAGSHARETHEME=$(SOURCE)/libs/gldi/maindock.CairoDockShareThemesDir '/usr/share/cairo-dock/themes'
# FLAGLOCALE=$(SOURCE)/libs/gldi/maindock.CairoDockLocaleDir '/usr/share/locale'

FLAGAPPVERSION=$(SOURCE)/libs/cdglobal.AppVersion '$(VERSION)'
FLAGGITHASH=$(SOURCE)/libs/cdglobal.GitHash '$(shell git rev-parse HEAD)'
# git describe --tags
FLAGBUILDDATE=$(SOURCE)/libs/cdglobal.BuildDate '$(shell date --rfc-3339=seconds)'


FLAGS=-ldflags "-X $(FLAGAPPVERSION) -X $(FLAGBUILDDATE) -X $(FLAGGITHASH) "


%: build

archive-%:
	go build -tags '$(APPLETS)'  -o applets/$(TARGET) $(SOURCE)/cmd/$(TARGET)
	@echo "Make archive $(TARGET)-$(VERSION)-$*.tar.xz"
	tar cJfv $(TARGET)-$(VERSION)-$*.tar.xz applets  --exclude-vcs
	rm applets/$(TARGET)

build:
	go install -tags '$(APPLETS)'  $(FLAGS) $(SOURCE)/cmd/$(TARGET)

unstable:
	go install -tags '$(APPLETS) $(UNSTABLE) $(UNSTABLE_TAGS)' $(FLAGS) $(SOURCE)/cmd/$(TARGET)

dock:
	go install -tags '$(DOCK)' $(FLAGS) $(SOURCE)/cmd/$(TARGET)

patch:
	# Patch GTK - some patches required to build a dock.
	cd "$(GOPATH)/src/github.com/gotk3/gotk3" && git pull --commit --no-edit origin few_methods deprecated
	
	# Patch Dbus (for Notifications)
	cd "$(GOPATH)/src/github.com/godbus/dbus" && git pull --commit --no-edit https://github.com/sqp/dbus fixeavesdrop

patch-dock: patch

	# Patch gettext (for dock)
	cd "$(GOPATH)/src/github.com/gosexy/gettext" && git pull --commit --no-edit https://github.com/sqp/gettext nil_string


install:
	mkdir -p "$(PKGDIR)/usr/bin"
	install -p -m755 "$(GOPATH)/bin/cdc" "$(PKGDIR)/usr/bin"

	mkdir -p "$(PKGDIR)/$(APPDIRDBUS)"
	for f in $(APPLETS); do	\
		cp -Rv --preserve=timestamps "applets/$$f" "$(PKGDIR)/$(APPDIRDBUS)" ;\
		rm $(PKGDIR)/$(APPDIRDBUS)/$$f/applet.go ;\
		rm $(PKGDIR)/$(APPDIRDBUS)/$$f/last-modif ;\
		rm $(PKGDIR)/$(APPDIRDBUS)/$$f/Makefile ;\
	done

install-dock:
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




help:
	@# update command documentation.

	cdc help documentation > cmd/$(TARGET)/doc.go
	gofmt -w cmd/$(TARGET)/doc.go


stop:
	dbus-send --session --dest=org.cairodock.CairoDock /org/cdc/Cdc org.cairodock.CairoDock.Quit
	
	@## ActivateModule string:$(TARGET) boolean:false


cover:
	@# tests coverage with overalls: go get github.com/bluesuncorp/overalls

	overalls -covermode=count -debug  -project=$(SOURCE)
	go tool cover -html=overalls.coverprofile