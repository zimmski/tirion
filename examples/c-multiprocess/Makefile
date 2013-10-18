PROG = c-multiprocess
BB = $(GOPATH)/src/github.com/zimmski/tirion/examples/c-multiprocess
LD = $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client
LIBS = -lm -lpthread -ltirion
CC = clang
CFLAGS = -Wall -Wextra -W -g

OBJECTS = $(patsubst %.c, %.o, $(wildcard *.c))
HEADERS = $(wildcard *.h)

.PHONY: all gcc clean
.PRECIOUS: $(OBJECTS)

all: $(PROG)

%.o: %.c $(HEADERS)
	$(CC) $(CFLAGS) -c $< -o $@

$(PROG): $(OBJECTS)
	mkdir -p $(GOPATH)/bin
	$(CC) $(CFLAGS) -L $(LD) $(OBJECTS) $(LIBS) -o $(GOPATH)/bin/$@

clean:
	echo $(BB)
	-rm -f $(BB)/*.o
	-rm -f $(GOPATH)/bin/$(PROG)

gcc: CC = gcc
gcc: all

