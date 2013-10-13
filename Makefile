.PHONY: all c-client c-doc c-lib clients docs examples fmt go-doc go-client g-lib tirion-agent vet
all: tirion-agent
clients: c-client go-client
c-client:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client
go-client:
	go install github.com/zimmski/tirion/clients/go-client
docs: c-doc
c-doc:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client doc
libs: c-lib go-lib
c-lib:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client lib
go-lib:
	go install github.com/zimmski/tirion
tirion-agent:
	go install github.com/zimmski/tirion/tirion-agent
examples:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/examples/c-multiprocess
	go install github.com/zimmski/tirion/examples/go-mandelbrot
# Go coding conventions
fmt:
	gofmt -l -w -tabs=true $(GOPATH)/src/github.com/zimmski/tirion
universe: libs clients docs all examples
# Go static analysis
vet:
	go tool vet -all=true -v=true $(GOPATH)/src/github.com/zimmski/tirion
