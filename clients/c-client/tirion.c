#include <errno.h>
#include <limits.h>
#include <pthread.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdio.h>
#include <sys/shm.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/un.h>
#include <unistd.h>

#include "tirion.h"

#define TIRION_BUFFER_SIZE 4096
#define TIRION_TAG_SIZE 513

char defaultLogPrefix[] = "[client]";

void *tirionThreadHandleCommands(void* arg);

long tirionShmInit(Tirion *tirion, const char *filename, long count);
long tirionShmClose(Tirion *tirion);
float tirionShmGet(Tirion *tirion, long i);
long tirionShmRead(Tirion *tirion);
void tirionShmSet(Tirion *tirion, long i, float v);

long tirionSocketReceive(Tirion *tirion, char *buf, long size);
long tirionSocketSend(Tirion *tirion, const char *msg);

typedef struct TirionShmStruct {
	float *addr;
	bool create;
	long count;
	long id;
} TirionShm;

struct TirionPrivateStruct {
	long fd;
	char *logPrefix;
	long metricCount;
	TirionShm shm;
	char *socket;
	pthread_t *tHandleCommands;
};

Tirion *tirionNew(const char *socket, bool verbose) {
	Tirion *tirion = (Tirion*)malloc(sizeof(Tirion));
	tirion->p = (TirionPrivate*)malloc(sizeof(TirionPrivate));
	tirion->p->shm.id = -1;
	tirion->p->tHandleCommands = NULL;

	tirion->p->socket = strdup(socket);
	tirion->verbose = verbose;

	tirion->p->logPrefix = defaultLogPrefix;

	return tirion;
}

long tirionInit(Tirion *tirion) {
	long err;

	if (setsid() == -1) {
		tirionE(tirion, "Cannot set new session and group id of process");

		if (errno == EPERM) {
			tirionE(tirion, "No rights");
		}

		return TIRION_ERROR_SET_SID;
	}

	tirionV(tirion, "Open unix socket to %s", tirion->p->socket);

	struct sockaddr_un addr;

	if ((tirion->p->fd = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
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
	sprintf(buf, "tirion v%s\tshm", TIRION_VERSION);
	if ((err = tirionSocketSend(tirion, buf)) != TIRION_OK) {
		return err;
	}

	if ((err = tirionSocketReceive(tirion, buf, sizeof(buf) - 1)) != TIRION_OK) {
		return err;
	}

	char *tMetricCount = strtok(buf, "\t");
	char *tMetricUrl = strtok(NULL, "\t");

	if (strncmp(tMetricUrl, "shm://", 6) != 0) {
		tirionE(tirion, "Did not receive correct metric protocol URL");

		return TIRION_ERROR_METRIC_URL;
	}

	char *tShmPath = tMetricUrl + 6;

	struct stat statBuffer;

	char *endptr;
	long metricCount = strtol(tMetricCount, &endptr, 10);

	if (metricCount <= 0 || (errno == ERANGE && (metricCount == LONG_MAX || metricCount == LONG_MIN)) || (errno != 0 && metricCount == 0)) {
		tirionE(tirion, "Did not receive correct metric count");

		return TIRION_ERROR_METRIC_COUNT;
	} else if (tShmPath == NULL || strlen(tShmPath) == 0 || stat(tShmPath, &statBuffer) != 0) {
		tirionE(tirion, "Did not receive correct shm path");

		return TIRION_ERROR_SHM_PATH;
	}

	tirion->p->metricCount = metricCount;
	tirionV(tirion, "Received metric count %d and shm path %s", metricCount, tShmPath);

	tirionV(tirion, "Open shared memory");
	if ((err = tirionShmInit(tirion, tShmPath, metricCount)) != TIRION_OK) {
		return err;
	}

	tirionV(tirion, "Read shared memory");
	if ((err = tirionShmRead(tirion)) != TIRION_OK) {
		return err;
	}

	tirion->running = true;

	tirion->p->tHandleCommands = (pthread_t*)malloc(sizeof(pthread_t));

	long rHandleCommands = pthread_create(tirion->p->tHandleCommands, NULL, tirionThreadHandleCommands, (void*) tirion);
	if (rHandleCommands != 0) {
		tirionE(tirion, "Failed creating HandleCommands thread");

		return TIRION_ERROR_THREAD_HANDLE_COMMANDS;
	}

	return TIRION_OK;
}

long tirionClose(Tirion *tirion) {
	long err;

	tirion->running = false;

	if ((err = tirionShmClose(tirion)) != TIRION_OK) {
		return err;
	}
	if (shutdown(tirion->p->fd, SHUT_RDWR) == -1) {
		tirionE(tirion, "Cannot shutdown socket");

		return TIRION_ERROR_SOCKET_SHUTDOWN;
	}
	if (tirion->p->tHandleCommands != NULL) {
		if (pthread_join(*tirion->p->tHandleCommands, NULL) != 0) {
			tirionE(tirion, "Cannot join HandleCommands thread");

			return TIRION_ERROR_THREAD_JOIN;
		}

		free(tirion->p->tHandleCommands);

		tirion->p->tHandleCommands = NULL;
	}

	return TIRION_OK;
}

long tirionDestroy(Tirion *tirion) {
	free(tirion->p->socket);
	free(tirion->p);
	free(tirion);

	return TIRION_OK;
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

long tirionShmInit(Tirion *tirion, const char *filename, long count) {
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

long tirionShmClose(Tirion *tirion) {
	if (tirion->p->shm.id > 0 && shmdt(tirion->p->shm.addr) == -1) {
		tirionE(tirion, "Cannot detach shm");

		return TIRION_ERROR_SHM_DETACH;
	}

	tirion->p->shm.id = 0;

	return TIRION_OK;
}

float tirionShmGet(Tirion *tirion, long i) {
	if (i < 0 || i >= tirion->p->metricCount) {
		return 0.0f;
	}

	return tirion->p->shm.addr[i];
}

long tirionShmRead(Tirion *tirion) {
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

void tirionShmSet(Tirion *tirion, long i, float v) {
	if (i < 0 || i >= tirion->p->metricCount) {
		return;
	}

	tirion->p->shm.addr[i] = v;
}

float tirionAdd(Tirion *tirion, long i, float v) {
	if (i < 0 || i >= tirion->p->metricCount) {
		return 0.0f;
	}

	return tirion->p->shm.addr[i] = (tirion->p->shm.addr[i] + v);
}

float tirionDec(Tirion *tirion, long i) {
	return tirionSub(tirion, i, 1.0);
}

float tirionInc(Tirion *tirion, long i) {
	return tirionAdd(tirion, i, 1.0);
}

float tirionSub(Tirion *tirion, long i, float v) {
	if (i < 0 || i >= tirion->p->metricCount) {
		return 0.0f;
	}

	return tirion->p->shm.addr[i] = (tirion->p->shm.addr[i] - v);
}

long tirionTag(Tirion *tirion, const char *format, ...) {
	va_list args;

	char buf[TIRION_TAG_SIZE];

	buf[0] = 't';

	va_start(args, format);
	vsnprintf(&buf[1], TIRION_TAG_SIZE, format, args);
	va_end(args);

	char *c = &buf[1];

	for (; *c; c++) {
		if (*c == '\n') {
			*c = ' ';
		}
	}

	return tirionSocketSend(tirion, buf);
}

long tirionSocketReceive(Tirion *tirion, char *buf, long size) {
	long rc = recv(tirion->p->fd, buf, size, 0);

	if (rc <= 0) {
		if (rc == 0) {
			tirionV(tirion, "Unix socket got closed with EOF");
		} else {
			tirionE(tirion, "Unix socket receive error");
		}

		return TIRION_ERROR_SOCKET_RECEIVE;
	}

	buf[rc] = '\0';

	long sc = strlen(buf);

	if (buf[sc - 1] == '\n') {
		buf[sc - 1] = '\0';
	}

	return TIRION_OK;
}

long tirionSocketSend(Tirion *tirion, const char *msg) {
	long wc = send(tirion->p->fd, msg, strlen(msg), 0);

	if (wc <= 0) {
		tirionE(tirion, "Unix socket send error");

		return TIRION_ERROR_SOCKET_SEND;
	}

	return TIRION_OK;
}

void tirionD(const Tirion *tirion, const char *format, ...) {
	if (! tirion->verbose) {
		return;
	}

	va_list args;

	char buf[TIRION_BUFFER_SIZE];

	va_start(args, format);
	vsprintf(buf, format, args);
	va_end(args);

	fprintf(stderr, "%s[debug] %s\n", tirion->p->logPrefix, buf);
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

	fprintf(stderr, "%s[error] %s\n", tirion->p->logPrefix, buf);
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

	fprintf(stderr, "%s[verbose] %s\n", tirion->p->logPrefix, buf);
}
