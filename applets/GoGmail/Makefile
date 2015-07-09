TARGET=GoGmail
SOURCE=github.com/sqp/godock/applets

%: build

build:
	go build -o $(TARGET) $(SOURCE)/$(TARGET)

link:
	ln -s $(GOPATH)/src/$(SOURCE)/$(TARGET) $(HOME)/.config/cairo-dock/third-party/$(TARGET)

start:
	@dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$(TARGET) boolean:true

stop:
	@dbus-send --session --dest=org.cairodock.CairoDock /org/cairodock/CairoDock org.cairodock.CairoDock.ActivateModule string:$(TARGET) boolean:false


restart: start stop


# must drop the libnotify file for cross compile as it require cgo which isn't available in that mode.
build-386:
	@echo "Build source: 386"
	GOARCH=386 go build -o $(TARGET) src/config.go src/GoGmail.go src/mail.go

build-amd64:
	@echo "Build source: amd64"
	GOARCH=amd64 go build -o $(TARGET) src/*.go

