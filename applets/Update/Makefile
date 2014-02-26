TARGET=Update
VERSION=0.0.3
SOURCE=github.com/sqp/godock/applets

# Default is standard build for current arch.

%: build

build:
	go build -o $(TARGET) $(SOURCE)/$(TARGET)/src

link:
	ln -s $(GOPATH)/src/$(SOURCE)/$(TARGET) $(HOME)/.config/cairo-dock/third-party/$(TARGET)

active:
	@dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$(TARGET) boolean:true

inactive:
	@dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$(TARGET) boolean:false

# reload: inactive active
reload: 
	cdc reload Update

restart-dock:
	@dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.Quit
	cairo-dock&

archives: archive-386 archive-amd64

# Argument for build and archive is a golang linux arch: 386, amd64, arm.

build-%:
	@echo "Build source: $*"
	@GOARCH=$* go build -o $(TARGET) $(SOURCE)/$(TARGET)/src

archive-%: build-%
	@echo "Make archive ../$(TARGET)-$(VERSION)-$*.tar.xz"
	@cd .. && tar cJfv $(TARGET)-$(VERSION)-$*.tar.xz $(TARGET)/  --exclude-vcs


