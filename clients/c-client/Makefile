PROG = c-client
BB = $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client
LIBS = -lm -lpthread
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
	$(CC) $(CFLAGS) $(OBJECTS) $(LIBS) -o $(GOPATH)/bin/$@

clean:
	echo $(BB)
	-rm -f $(BB)/*.o
	-rm -f $(GOPATH)/bin/$(PROG)

gcc: CC = gcc
gcc: all

lib: tirion.o
	ar rcs libtirion.a tirion.o

