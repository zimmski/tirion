#include <errno.h>
#include <limits.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdio.h>
#include <sys/shm.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/un.h>
#include <unistd.h>

#include "tirion.h"

#define TIRION_BUFFER_SIZE 4096

void *tirionThreadHandleCommands(void* arg);

int tirionShmInit(Tirion *tirion, const char *filename, int count);
int tirionShmClose(Tirion *tirion);
float tirionShmGet(Tirion *tirion, int i);
int tirionShmRead(Tirion *tirion);
void tirionShmSet(Tirion *tirion, int i, float v);

int tirionSocketReceive(Tirion *tirion, char *buf, int size);
int tirionSocketSend(Tirion *tirion, const char *msg);

typedef struct TirionPrivateStruct {
	int fd;
	TirionShm shm;
	char *socket;
	pthread_t tHandleCommands;
} TirionPrivate;

Tirion *tirionNew(const char *socket, bool verbose) {
	Tirion *tirion = malloc(sizeof(Tirion));
	tirion->p = malloc(sizeof(TirionPrivate));

	tirion->p->socket = strdup(socket);
	tirion->verbose = verbose;

	tirion->logPrefix = "[client]";

	return tirion;
}

int tirionInit(Tirion *tirion) {
	int err;

	// TODO should we create a sighandler to set tirion->running = false?

	tirionV(tirion, "Open unix socket to %s", tirion->p->socket);

	struct sockaddr_un addr;

	if ( (tirion->p->fd = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
		tirionE(tirion, "socket create error");

		return TIRION_ERROR_SOCKET_CREATE;
	}

	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, tirion->p->socket, sizeof(addr.sun_path)-1);

	if (connect(tirion->p->fd, (struct sockaddr*)&addr, sizeof(addr)) == -1) {
		tirionE(tirion, "socket connect error");

		return TIRION_ERROR_SOCKET_CONNECT;
	}

	char buf[TIRION_BUFFER_SIZE];

	tirionV(tirion, "Request tirion protocol version v%s", TIRION_VERSION);
	sprintf(buf, "tirion v%s", TIRION_VERSION);
	if ((err = tirionSocketSend(tirion, buf)) != TIRION_OK) {
		return err;
	}

	if ((err = tirionSocketReceive(tirion, buf, sizeof(buf) - 1)) != TIRION_OK) {
		return err;
	}

	char *endptr;
	int metricCount = strtol(buf, &endptr, 10);

	if (metricCount <= 0 || (errno == ERANGE && (metricCount == INT_MAX || metricCount == INT_MIN)) || (errno != 0 && metricCount == 0)) {
		tirionE(tirion, "Did not receive metric count");

		return TIRION_ERROR_METRIC_COUNT;
	}

	tirionV(tirion, "Received metric count %d", metricCount);

	tirionV(tirion, "Open shared memory");
	if ((err = tirionShmInit(tirion, "/tmp", metricCount)) != TIRION_OK) {
		return err;
	}

	tirionV(tirion, "Read shared memory");
	if ((err = tirionShmRead(tirion)) != TIRION_OK) {
		return err;
	}

	tirion->running = true;

	int rHandleCommands = pthread_create(&tirion->p->tHandleCommands, NULL, tirionThreadHandleCommands, (void*) tirion);
	if (rHandleCommands != 0) {
		tirionE(tirion, "Failed creating HandleCommands thread");

		return TIRION_ERROR_THREAD_HANDLE_COMMANDS;
	}

	return TIRION_OK;
}

int tirionClose(Tirion *tirion) {
	int err;

	tirion->running = false;

	if ((err = tirionShmClose(tirion)) != TIRION_OK) {
		return err;
	}
	if (shutdown(tirion->p->fd, SHUT_RDWR) == -1) {
		tirionE(tirion, "Cannot shutdown socket");

		return TIRION_ERROR_SOCKET_SHUTDOWN;
	}
	if (pthread_join(tirion->p->tHandleCommands, NULL) != 0) {
		tirionE(tirion, "Cannot join HandleCommands thread");

		return TIRION_ERROR_THREAD_JOIN;
	}

	return TIRION_OK;
}

void tirionDestroy(Tirion *tirion) {
	free(tirion->p->socket);
	free(tirion->p);
	free(tirion);
}

void *tirionThreadHandleCommands(void *arg) {
	Tirion *tirion = (Tirion*)arg;

	char data[TIRION_BUFFER_SIZE];

	tirionV(tirion, "Start listening to commands");

	while (tirion->running) {
		if (tirionSocketReceive(tirion, data, sizeof(data) - 1) != TIRION_OK) {
			tirion->running = false;
		} else {
			char com = data[0];

			switch (com) {
			default:
				tirionE(tirion, "Unknown command '%c'", com);
			}
		}
	}

	tirionV(tirion, "Stop listening to commands");

	return NULL;
}

int tirionShmInit(Tirion *tirion, const char *filename, int count) {
	key_t key = ftok(filename, 0x03);

	if (key == -1) {
		tirionE(tirion, "Cannot generate shm key");

		return TIRION_ERROR_SHM_KEY;
	}

	tirion->p->shm.id = shmget(key, 0, 0);
	tirion->p->shm.count = count;

	if (tirion->p->shm.id == -1) {
		tirionE(tirion, "Cannot initialize shm");

		return TIRION_ERROR_SHM_INIT;
	}

	return TIRION_OK;
}

int tirionShmClose(Tirion *tirion) {
	if (shmdt(tirion->p->shm.addr) == -1) {
		tirionE(tirion, "Cannot detach shm");

		return TIRION_ERROR_SHM_DETACH;
	}

	return TIRION_OK;
}

float tirionShmGet(Tirion *tirion, int i) {
	return tirion->p->shm.addr[i];
}

int tirionShmRead(Tirion *tirion) {
	if (tirion->p->shm.id <= 0) {
		tirionE(tirion, "Shm is not initialized");

		return TIRION_ERROR_SHM_NO_INIT;
	}

	tirion->p->shm.addr = (float*)shmat(tirion->p->shm.id, NULL, 0);

	if (tirion->p->shm.addr == (float*)-1) {
		tirion->p->shm.addr = NULL;

		tirionE(tirion, "Cannot attach shm");

		return TIRION_ERROR_SHM_ATTACH;
	}

	return TIRION_OK;
}

void tirionShmSet(Tirion *tirion, int i, float v) {
	tirion->p->shm.addr[i] = v;
}

float tirionAdd(Tirion *tirion, int i, float v) {
	return tirion->p->shm.addr[i] = (tirion->p->shm.addr[i] + v);
}

float tirionDec(Tirion *tirion, int i) {
	return tirionSub(tirion, i, 1.0);
}

float tirionInc(Tirion *tirion, int i) {
	return tirionAdd(tirion, i, 1.0);
}

float tirionSub(Tirion *tirion, int i, float v) {
	return tirion->p->shm.addr[i] = (tirion->p->shm.addr[i] - v);
}

int tirionTag(Tirion *tirion, const char *format, ...) {
	va_list args;

	char buf[TIRION_BUFFER_SIZE];

	buf[0] = 't';

	va_start(args, format);
	vsprintf(&buf[1], format, args);
	va_end(args);

	return tirionSocketSend(tirion, buf);
}

int tirionSocketReceive(Tirion *tirion, char *buf, int size) {
	int rc = recv(tirion->p->fd, buf, size, 0);

	if (rc <= 0) {
		if (rc == 0) {
			tirionE(tirion, "Unix socket got closed with EOF");
		} else {
			tirionE(tirion, "Unix socket receive error");
		}

		return TIRION_ERROR_SOCKET_RECEIVE;
	}

	buf[rc] = '\0';

	int sc = strlen(buf);

	if (buf[sc - 1] == '\n') {
		buf[sc - 1] = '\0';
	}

	return TIRION_OK;
}

int tirionSocketSend(Tirion *tirion, const char *msg) {
	int wc = send(tirion->p->fd, msg, strlen(msg), 0);

	if (wc <= 0) {
		tirionE(tirion, "Unix socket send error");

		return TIRION_ERROR_SOCKET_SEND;
	}

	return TIRION_OK;
}

// TODO remove this or maybe we should put them in a logging package or use another logging package

void tirionD(const Tirion *tirion, const char *format, ...) {
	if (! tirion->verbose) {
		return;
	}

	va_list args;

	char buf[TIRION_BUFFER_SIZE];

	va_start(args, format);
	vsprintf(buf, format, args);
	va_end(args);

	fprintf(stderr, "%s[debug] %s\n", tirion->logPrefix, buf);
}

void tirionE(const Tirion *tirion, const char *format, ...) {
	if (! tirion->verbose) {
		return;
	}

	va_list args;

	char buf[TIRION_BUFFER_SIZE];

	va_start(args, format);
	vsprintf(buf, format, args);
	va_end(args);

	fprintf(stderr, "%s[error] %s\n", tirion->logPrefix, buf);
}

void tirionV(const Tirion *tirion, const char *format, ...) {
	if (! tirion->verbose) {
		return;
	}

	va_list args;

	char buf[TIRION_BUFFER_SIZE];

	va_start(args, format);
	vsprintf(buf, format, args);
	va_end(args);

	fprintf(stderr, "%s[verbose] %s\n", tirion->logPrefix, buf);
}
