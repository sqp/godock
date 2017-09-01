TARGET=cdc
VERSION=0.0.3.5

SOURCE=github.com/sqp/godock

APPLETS=Audio Clouds Cpu DiskActivity DiskFree GoGmail Mem NetActivity Update


# unstable applets requires unmerged patches to build.
UNSTABLE=Notifications TVPlay
UNSTABLE_TAGS=gtk

# and dock even more, plus the rewritten dock.
DOCK=dock Audio Clouds Cpu DiskActivity DiskFree GoGmail Mem NetActivity Update Notifications
#all

# Install prefix if any.
PKGDIR=

APPDIRGLDI=usr/share/cairo-dock/appletsgo/
APPDIRDBUS=usr/share/cairo-dock/plug-ins/Dbus/third-party/


# old version had:
# Could be useful for some distro packagers.
# FLAGSHARETHEME=$(SOURCE)/libs/gldi/maindock.CairoDockShareThemesDir '/usr/share/cairo-dock/themes'
# FLAGLOCALE=$(SOURCE)/libs/gldi/maindock.CairoDockLocaleDir '/usr/share/locale'



BUILDDATE=$(shell date --rfc-3339=seconds)
NBEDITED=$(shell git diff --numstat | wc -l) + $(shell git diff --cached --numstat | wc -l) + $(shell git ls-files --others | wc -l)
# git describe --tags

# trim GOPATH: https://github.com/golang/go/issues/13809

FLAGS=-gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH \
	-ldflags " \
	-X '$(SOURCE)/libs/cdglobal.AppVersion=$(VERSION)' \
	-X '$(SOURCE)/libs/cdglobal.BuildMode=makefile' \
	-X '$(SOURCE)/libs/cdglobal.BuildDate=$(BUILDDATE)' \
	-X '$(SOURCE)/libs/cdglobal.BuildNbEdited=$(NBEDITED)' \
	-X '$(SOURCE)/libs/cdglobal.GitHash=$(shell git rev-parse HEAD)' "
	


%: build


build:
	go install -tags '$(APPLETS)'  $(FLAGS) $(SOURCE)/cmd/$(TARGET)

unstable:
	go install -tags '$(APPLETS) $(UNSTABLE) $(UNSTABLE_TAGS)' $(FLAGS) $(SOURCE)/cmd/$(TARGET)

dock:
	go install \
		-tags '$(DOCK)' \
		$(FLAGS) \
		$(SOURCE)/cmd/$(TARGET)

patch:
	# Patch Dbus (for Notifications)
	cd "$(GOPATH)/src/github.com/godbus/dbus" && git pull --commit --no-edit https://github.com/sqp/dbus fixeavesdrop

	# Patch ini parser for config.
	cd "$(GOPATH)/src/github.com/go-ini/ini"  && git pull --commit --no-edit https://github.com/sqp/ini gtk_keyfile_compat

	# Patch webserver restarter to use default mux.
	cd "$(GOPATH)/src/github.com/braintree/manners"  && git pull --commit --no-edit https://github.com/grazzini/manners master

patch-dock: patch


install: install-common

	install -Dm755 "$(GOPATH)/src/$(SOURCE)/cmd/$(TARGET)/data/tocdc"  "$(PKGDIR)/usr/bin/tocdc"

	install -d "$(PKGDIR)/$(APPDIRDBUS)"
	for f in $(APPLETS); do	\
		cp -Rv --preserve=timestamps "applets/$$f"  "$(PKGDIR)/$(APPDIRDBUS)" ;\
		rm "$(PKGDIR)/$(APPDIRDBUS)/$$f/applet.go" ;\
		rm "$(PKGDIR)/$(APPDIRDBUS)/$$f/Makefile" ;\
		ln -s "/usr/bin/tocdc"  "$(PKGDIR)/$(APPDIRDBUS)/$$f/$$f" ;\
	done

	install -Dm644 "$(GOPATH)/src/$(SOURCE)/LICENSE"  "$(PKGDIR)/usr/share/licenses/cairo-dock-goapplets/LICENSE"


install-dock: install-common

	install -d "$(PKGDIR)/$(APPDIRGLDI)"
	for f in $(APPLETS); do	\
		cp -Rv --preserve=timestamps "applets/$$f"  "$(PKGDIR)/$(APPDIRGLDI)" ;\
		rm $(PKGDIR)/$(APPDIRGLDI)/$$f/applet.go ;\
		rm $(PKGDIR)/$(APPDIRGLDI)/$$f/Makefile ;\
	done

	install -Dm644 "$(GOPATH)/src/$(SOURCE)/LICENSE"  "$(PKGDIR)/usr/share/licenses/cairo-dock-rework/LICENSE"

	install -d "$(GOPATH)/src/$(SOURCE)/data/defaults"   "$(PKGDIR)/usr/share/cairo-dock/defaults"
	install -d "$(GOPATH)/src/$(SOURCE)/data/templates"  "$(PKGDIR)/usr/share/cairo-dock/templates"

	install -D "$(GOPATH)/src/$(SOURCE)/cmd/$(TARGET)/data/cmd.desktop"         "$(PKGDIR)/usr/share/applications/cairo-dock-rework.desktop"
	install -D "$(GOPATH)/src/$(SOURCE)/cmd/$(TARGET)/data/rework.conf"         "$(PKGDIR)/usr/share/cairo-dock/defaults/rework.conf"


install-common:
	install -p -Dm755 "$(GOPATH)/bin/$(TARGET)"  "$(PKGDIR)/usr/bin/$(TARGET)"

	install -D "$(GOPATH)/src/$(SOURCE)/cmd/$(TARGET)/data/upload.nemo_action"  "$(PKGDIR)/usr/share/nemo/actions/cairo-dock-rework_upload.nemo_action"

	gzip -9 < "$(GOPATH)/src/$(SOURCE)/cmd/$(TARGET)/data/man.1" > "$(GOPATH)/src/$(SOURCE)/cmd/$(TARGET)/data/man.1.gz"
	install -pD "$(GOPATH)/src/$(SOURCE)/cmd/$(TARGET)/data/man.1.gz" "$(PKGDIR)/usr/share/man/man1/cdc.1.gz"


help:
	@# update command documentation.

	$(TARGET) help documentation > cmd/$(TARGET)/doc.go
	gofmt -w cmd/$(TARGET)/doc.go


stop:
	dbus-send --session --dest=org.cairodock.CairoDock /org/cdc/Cdc org.cairodock.CairoDock.Quit
	
	@## ActivateModule string:$(TARGET) boolean:false


cover:
	@# tests coverage with overalls: go get github.com/bluesuncorp/overalls

	overalls -covermode=count -debug  -project=$(SOURCE)
	go tool cover -html=overalls.coverprofile


cover-all:
	@# merged test coverage: go get -u github.com/ory/go-acc
	@# https://www.ory.am/golang-go-code-coverage-accurate.html

	go-acc $(SOURCE)/...
	go tool cover -html=coverage.txt

dep-list:
	@go list -tags '$(DOCK)' -f '{{join .Deps "\n"}}' $(SOURCE)/cmd/$(TARGET) | \
		xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | \
		grep -v "github.com/sqp/" | \
		sort

dep-graph:
	godepgraph -s -tags '$(DOCK)' $(SOURCE)/cmd/$(TARGET) | dot -Tsvg -o deps.svg
	xdg-open deps.svg &


# archive-%:
# 	go build -tags '$(APPLETS)'  -o applets/$(TARGET) $(SOURCE)/cmd/$(TARGET)
# 	@echo "Make archive $(TARGET)-$(VERSION)-$*.tar.xz"
# 	tar cJfv $(TARGET)-$(VERSION)-$*.tar.xz applets  --exclude-vcs
# 	rm applets/$(TARGET)
