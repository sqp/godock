PACKAGE=confsettings
CONFIG=gui.conf


%: bindata


bindata:
	rm -f $(CONFIG).go
	go-bindata -func=guiConfDefault -pkg=$(PACKAGE) $(CONFIG) 
