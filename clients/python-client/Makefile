PROG = python-client
BB = $(GOPATH)/src/github.com/zimmski/tirion/clients/python-client

.PHONY: doc clean
.PRECIOUS: $(OBJECTS)


doc:
	cd $(BB)/doc ;\
	doxygen doxygen.conf

lib:
	sudo python setup.py install --verbose
