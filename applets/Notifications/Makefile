TARGET=Notifications
SOURCE=github.com/sqp/godock/applets

# Default is standard build for current arch.

%: build

build:
	go build -o $(TARGET) $(SOURCE)/$(TARGET)

link:
	ln -s $(GOPATH)/src/$(SOURCE)/$(TARGET) $(HOME)/.config/cairo-dock/third-party/$(TARGET)
