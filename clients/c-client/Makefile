PROG = c-client
BB = $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client
LIBS = -lm -lpthread
CC = gcc
CFLAGS = -Wall -Wextra -W -g

OBJECTS = $(patsubst %.c, %.o, $(wildcard *.c))
HEADERS = $(wildcard *.h)

.PHONY: all clean
.PRECIOUS: $(OBJECTS)

all: $(PROG)

%.o: %.c $(HEADERS)
	$(CC) $(CFLAGS) -c $< -o $@

$(PROG): $(OBJECTS)
	$(CC) $(CFLAGS) $(OBJECTS) $(LIBS) -o $(GOPATH)/bin/$@

clean:
	echo $(BB)
	-rm -f $(BB)/*.o
	-rm -f $(GOPATH)/bin/$(PROG)
