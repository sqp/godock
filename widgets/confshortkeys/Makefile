TARGET=confshortkeys

%: bindata


bindata:
	rm -f $(TARGET).xml.go
	go-bindata -func=$(TARGET)XML -pkg=$(TARGET) $(TARGET).xml 

