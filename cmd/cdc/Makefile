TARGET=cdc
VERSION=0.0.1-1
SOURCE=github.com/sqp/godock/cmd
APPLETS=Audio Cpu DiskActivity DiskFree GoGmail Mem NetActivity Update

# unstable applets requires uncommited patches to build.
UNSTABLE=Notifications TVPlay config log gtk


%: build

build:
	go install -tags '$(APPLETS)'  $(SOURCE)/$(TARGET)

link:
	ln -s $(GOPATH)/src/$(SOURCE)/$(TARGET) $(HOME)/.config/cairo-dock/third-party/$(TARGET)

list:
	@dbus-send --session --dest=org.cairodock.GoDock /org/cairodock/GoDock org.cairodock.GoDock.ListServices

stop:
	@dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$(TARGET) boolean:false


doc:
	cdc help documentation > doc.go
	gofmt -w doc.go




# Argument for archive is a golang linux arch: 386, amd64, arm.
archives: archive-386 archive-amd64


archive-%: build-%
	@echo "Make archive ../$(TARGET)-$(VERSION)-$*.tar.xz"
	@cd .. && tar cJfv $(TARGET)-$(VERSION)-$*.tar.xz $(TARGET)/  --exclude-vcs

# must drop the libnotify file for cross compile as it require cgo which isn't available in that mode.
build-386:
	@echo "Build source: 386"
	GOARCH=386 go build -o $(TARGET) src/config.go src/GoGmail.go src/mail.go

build-amd64:
	@echo "Build source: amd64"
	GOARCH=amd64 go build -o $(TARGET) src/*.go

