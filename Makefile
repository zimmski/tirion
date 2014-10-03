.PHONY: all c-client c-doc c-lib clients docs examples fmt go-client go-doc go-lib java-client java-doc java-lib lint python-client python-doc python-lib tirion-agent
all: tirion-agent
clean:
	rm -fr $(GOPATH)/pkg/*/github.com/zimmski/tirion*

	go clean github.com/zimmski/tirion
	go clean github.com/zimmski/tirion/backend
	go clean github.com/zimmski/tirion/tirion-agent
	go clean github.com/zimmski/tirion/tirion-server

	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client clean
	go clean github.com/zimmski/tirion/clients/go-client
	ant -buildfile $(GOPATH)/src/github.com/zimmski/tirion/clients/java-client/Tirion/build.xml clean

	make -C $(GOPATH)/src/github.com/zimmski/tirion/examples/c-multiprocess clean
	go clean github.com/zimmski/tirion/examples/go-mandelbrot
clients: c-client java-client go-client python-client
c-client:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client
go-client:
	go install github.com/zimmski/tirion/clients/go-client
java-client:
	ant -buildfile $(GOPATH)/src/github.com/zimmski/tirion/clients/java-client/Tirion/build.xml client
	cp $(GOPATH)/src/github.com/zimmski/tirion/clients/java-client/Tirion/bin/java-client.jar $(GOPATH)/bin/java-client.jar
python-client: python-lib
	cp $(GOPATH)/src/github.com/zimmski/tirion/clients/python-client/python_client.py $(GOPATH)/bin/python_client.py
	chmod +x $(GOPATH)/bin/python_client.py
dependencies:
	sh $(GOPATH)/src/github.com/zimmski/tirion/scripts/dependencies.sh
docs: c-doc java-doc python-doc
c-doc:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client doc
java-doc:
	ant -buildfile $(GOPATH)/src/github.com/zimmski/tirion/clients/java-client/Tirion/build.xml doc
python-doc:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/python-client doc
libs: c-lib java-lib go-lib python-lib
c-lib:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client lib
go-lib:
	go install github.com/zimmski/tirion
java-lib:
	ant -buildfile $(GOPATH)/src/github.com/zimmski/tirion/clients/java-client/Tirion/build.xml lib
python-lib:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/python-client lib
lint:
	go install github.com/zimmski/tirion/backend

	go tool vet -all=true -v=true $(GOPATH)/src/github.com/zimmski/tirion
	golint $(GOPATH)/src/github.com/zimmski/tirion/...
tirion-agent:
	go install github.com/zimmski/tirion/tirion-agent
examples:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/examples/c-multiprocess
	go install github.com/zimmski/tirion/examples/go-mandelbrot
fmt:
	gofmt -l -w -tabs=true $(GOPATH)/src/github.com/zimmski/tirion
package: clean
package:
	# Currently Go does not allow crosscompiling of programs using cgo.
	# We use cgo in /tirion/shm. So right now, we have to manually compile on
	# different hosts :-(
	#GOOS=linux GOARCH=amd64 sh $(GOPATH)/src/github.com/zimmski/tirion/scripts/package.sh
	#GOOS=linux GOARCH=386 sh $(GOPATH)/src/github.com/zimmski/tirion/scripts/package.sh

	GOOS=linux sh $(GOPATH)/src/github.com/zimmski/tirion/scripts/package.sh
universe: libs clients docs all examples
