#include <stdlib.h>
#include <string.h>
#include <sys/shm.h>
#include <sys/types.h>

#include "shm_linux.h"

float* shmAttach(long shm_id) {
	float* addr = (float*)shmat(shm_id, NULL, 0);

	if (addr == (float*)-1) {
		return NULL;
	}

	return addr;
}

long shmClose(long shm_id) {
	return shmctl(shm_id, IPC_RMID, NULL);
}

void shmCopy(float* from, float* to, long count) {
	long i = 0;

	for (; i < count; i++) {
		to[i] = from[i];
	}
}

long shmDetach(float *addr) {
	return shmdt(addr);
}

float shmGet(float* addr, long i) {
	return addr[i];
}

long shmOpen(char* filename, char create, long count) {
	key_t key = ftok(filename, 0x03);

	if (key == -1) {
		return -1;
	}

	if (create) {
		return shmget(key, sizeof(float) * count, IPC_CREAT|IPC_EXCL|0600);
	} else {
		return shmget(key, 0, 0);
	}
}

float shmSet(float *addr, long i, float v) {
	return addr[i] = v;
}

float shmAdd(float* addr, long i, float v) {
	return addr[i] = (addr[i] + v);
}

float shmDec(float* addr, long i) {
	return shmSub(addr, i, 1.0);
}

float shmInc(float* addr, long i) {
	return shmAdd(addr, i, 1.0);
}

float shmSub(float* addr, long i, float v) {
	return addr[i] = (addr[i] - v);
}

