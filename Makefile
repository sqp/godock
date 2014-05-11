TARGET=cdc
SOURCE=github.com/sqp/godock/cmd
VERSION=0.0.1-2
APPLETS=Audio DiskActivity DiskFree GoGmail NetActivity Update

# unstable applets requires uncommited patches to build.
UNSTABLE=Notifications TVPlay config log gtk

%: build

archive-%:
	go build -tags '$(APPLETS)'  -o applets/$(TARGET) $(SOURCE)/$(TARGET)
	@echo "Make archive $(TARGET)-$(VERSION)-$*.tar.xz"
	tar cJfv $(TARGET)-$(VERSION)-$*.tar.xz applets  --exclude-vcs
	rm applets/$(TARGET)

build:
	go install -tags '$(APPLETS)' $(SOURCE)/$(TARGET)

unstable:
	go install -tags '$(APPLETS) $(UNSTABLE)' $(SOURCE)/$(TARGET)

