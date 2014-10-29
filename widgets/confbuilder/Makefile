TARGET=handbook
PACKAGE=confbuilder

%: bindata


bindata:
	rm -f $(TARGET).xml.go
	go-bindata -func=$(TARGET)XML -pkg=$(PACKAGE) $(TARGET).xml 

