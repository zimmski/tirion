#include <stdlib.h>
#include <string.h>
#include <sys/shm.h>
#include <sys/types.h>

#include "shm_linux.h"

float* shmAttach(int shm_id) {
	float* addr = (float*)shmat(shm_id, NULL, 0);

	if (addr == (float*)-1) {
		return NULL;
	}

	return addr;
}

int shmClose(int shm_id) {
	return shmctl(shm_id, IPC_RMID, NULL);
}

void shmCopy(float* from, float* to, int count) {
	int i = 0;

	for (; i < count; i++) {
		to[i] = from[i];
	}
}

int shmDetach(float *addr) {
	return shmdt(addr);
}

float shmGet(float* addr, int i) {
	return addr[i];
}

int shmOpen(char* filename, int create, int count) {
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

float shmSet(float *addr, int i, float v) {
	return addr[i] = v;
}

float shmAdd(float* addr, int i, float v) {
	return addr[i] = (addr[i] + v);
}

float shmDec(float* addr, int i) {
	return shmSub(addr, i, 1.0);
}

float shmInc(float* addr, int i) {
	return shmAdd(addr, i, 1.0);
}

float shmSub(float* addr, int i, float v) {
	return addr[i] = (addr[i] - v);
}

