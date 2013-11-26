#include <fcntl.h>
//#include <stdio.h>
#include <sys/mman.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>

//#include <errno.h>
//#include <string.h>

#include "mmap_linux.h"

float *mmapOpen(const char *filename, char create, long count) {
	void *addr;
	int file;
	int size = sizeof(float) * count;

	if (
			(create && (file = open(filename, O_RDWR | O_CREAT | O_TRUNC, S_IRUSR | S_IWUSR)) < 0)
			|| (!create && (file = open(filename, O_RDWR)) < 0)
	) {
		//printf("Cannot open file: %s\n", strerror(errno));

		return NULL;
	}

	if (lseek(file, size - 1, SEEK_SET) == -1) {
		//printf("Cannot seek to last byte in file\n");

		return NULL;
	}

	if (write(file, "", 1) != 1) {
		//printf("Cannot write dummy byte into file\n");

		return NULL;
	}

	if ((addr = mmap(0, size, PROT_READ | PROT_WRITE, MAP_SHARED | MAP_LOCKED, file, 0)) == (caddr_t) -1) {
		//printf("Cannot mmap the file\n");

		return NULL;
	}

	close(file);

	return (float*)addr;
}

int mmapClose(float *addr, const char *filename, char create, long count) {
	if (create) {
		unlink(filename);
	}

	return munmap((void*)addr, sizeof(float) * count);
}

void mmapCopy(float* from, float* to, long count) {
	long i = 0;

	for (; i < count; i++) {
		to[i] = from[i];
	}
}

float mmapAdd(float* addr, long i, float v) {
	return addr[i] = (addr[i] + v);
}

float mmapDec(float* addr, long i) {
	return mmapAdd(addr, i, -1.0);
}

float mmapInc(float* addr, long i) {
	return mmapAdd(addr, i, 1.0);
}

float mmapSub(float* addr, long i, float v) {
	return mmapAdd(addr, i, -v);
}
