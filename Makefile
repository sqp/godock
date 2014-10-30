TARGET=cdc
SOURCE=github.com/sqp/godock/cmd
VERSION=0.0.1-2
APPLETS=Audio Cpu DiskActivity DiskFree GoGmail Mem NetActivity Update

# unstable applets requires uncommited patches to build.
UNSTABLE=Notifications TVPlay log gtk
DOCK=dock

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

dock:
	go install -tags '$(APPLETS) $(UNSTABLE) $(DOCK)' $(SOURCE)/$(TARGET)

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