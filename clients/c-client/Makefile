PROG = c-client
BB = $(GOPATH)/src/github.com/zimmski/tirion/clients/c-client
LIBS = -lm -lpthread
CC = clang
CFLAGS = -W -g
CWFLAGS = -Wall -Weverything

OBJECTS = $(patsubst %.c, %.o, $(wildcard *.c))
HEADERS = $(wildcard *.h)

.PHONY: all doc gcc clean
.PRECIOUS: $(OBJECTS)

all: $(PROG)

%.o: %.c $(HEADERS)
	$(CC) $(CWFLAGS) $(CFLAGS) -c $< -o $@

$(PROG): $(OBJECTS)
	mkdir -p $(GOPATH)/bin
	$(CC) $(CWFLAGS) $(CFLAGS) $(OBJECTS) $(LIBS) -o $(GOPATH)/bin/$@

doc:
	cd $(BB)/doc ;\
        doxygen doxygen.conf

clean:
	echo $(BB)
	-rm -f $(BB)/*.o
	-rm -f $(GOPATH)/bin/$(PROG)

gcc: CC = gcc
gcc: CWFLAGS = -Wall -Wextra
gcc: all

lib: tirion.o
	ar rcs libtirion.a tirion.o

