.PHONY: install clean test save

PROGRAMNAME:=goserver
CURDIR:=$(shell pwd)
OLDGOPATH=$(GOPATH)

all: save install

save:
	@export GOPATH=$(CURDIR); go get -v ./...; cd $(CURDIR)/src/$(PROGRAMNAME); godep save

restore:
	@export GOPATH=$(CURDIR); cd $(CURDIR)/src/$(PROGRAMNAME); godep restore

install:
	@export GOPATH=$(CURDIR); cd $(CURDIR)/src/$(PROGRAMNAME); godep go install

test:
	@export GOPATH=$(CURDIR); cd $(CURDIR)/src/$(PROGRAMNAME); godep go test 

clean:
	@rm -rf bin pkg

