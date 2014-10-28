TARGET=appletpreview

%: bindata


bindata:
	rm -f $(TARGET).xml.go
	go-bindata -pkg=$(TARGET) -func=$(TARGET)XML $(TARGET).xml 
