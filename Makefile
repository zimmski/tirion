.PHONY: all c-client examples fmt go-client tirion-agent vet
all: tirion-agent
clients: c-client go-client
c-client:
	make -C $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client
go-client:
	go install github.com/zimmski/tirion/clients/go-client
tirion-agent:
	go install github.com/zimmski/tirion/tirion-agent
examples:
	go install github.com/zimmski/tirion/examples/go-mandelbrot
# Go coding conventions
fmt:
	gofmt -l -w -tabs=true $(GOPATH)/src/github.com/zimmski/tirion
universe: all clients examples
# Go static analysis
vet:
	go tool vet -all=true -v=true $(GOPATH)/src/github.com/zimmski/tirion
