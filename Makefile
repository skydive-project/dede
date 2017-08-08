bindata:
	go-bindata ${BINDATA_FLAGS} -nometadata -o statics/bindata.go -pkg=statics -ignore=bindata.go statics/* statics/js/vendor/* statics/css/vendor/*
	gofmt -w -s statics/bindata.go

install: bindata
	go install -v ./...
