TARGET=index
PACKAGE=videodl

%: bindata


bindata:
	rm -f $(TARGET).html.go
	go-bindata -func=$(TARGET)HTML -pkg=$(PACKAGE) $(TARGET).html 
