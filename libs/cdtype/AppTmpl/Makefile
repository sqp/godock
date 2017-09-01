TARGET=AppTmpl

# set to your code directory in GOPATH (the applet parent dir).
DIRBASE=github.com/sqp/godock/libs/cdtype

# Default is standard build for current arch.
build:
	go build -o $(TARGET)

# make a link to the user external directory for easy install.
link:
	ln -s $(GOPATH)/src/$(DIRBASE)/$(TARGET) $(HOME)/.config/cairo-dock/third-party/$(TARGET)

# remove the link in the user external directory.
rmlink:
	rm $(HOME)/.config/cairo-dock/third-party/$(TARGET)
