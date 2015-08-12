TARGET=handbook
PACKAGE=handbook

%: bindata


bindata:
	rm -f $(TARGET).xml.go
	go-bindata -func=$(TARGET)XML -pkg=$(PACKAGE) $(TARGET).xml 

