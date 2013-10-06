#ifndef tirion_h_INCLUDED
#define tirion_h_INCLUDED

#include <pthread.h>
#include <stdbool.h>

#define TIRION_VERSION "0.1"

enum {
	TIRION_OK,
	TIRION_ERROR_METRIC_COUNT,
	TIRION_ERROR_SHM_ATTACH,
	TIRION_ERROR_SHM_DETACH,
	TIRION_ERROR_SHM_KEY,
	TIRION_ERROR_SHM_INIT,
	TIRION_ERROR_SHM_PATH,
	TIRION_ERROR_SHM_NO_INIT,
	TIRION_ERROR_SOCKET_CONNECT,
	TIRION_ERROR_SOCKET_CREATE,
	TIRION_ERROR_SOCKET_RECEIVE,
	TIRION_ERROR_SOCKET_SHUTDOWN,
	TIRION_ERROR_SOCKET_SEND,
	TIRION_ERROR_THREAD_HANDLE_COMMANDS,
	TIRION_ERROR_THREAD_JOIN,
};

typedef struct TirionShmStruct {
	long id;
	bool create;
	float *addr;
	long count;
} TirionShm;

typedef struct TirionPrivateStruct TirionPrivate;
typedef struct TirionStruct {
	bool running;
	bool verbose;
	char *logPrefix;
	TirionPrivate *p;
} Tirion;


Tirion *tirionNew(const char *socket, bool verbose);
long tirionInit(Tirion *tirion);
long tirionClose(Tirion *tirion);
void tirionDestroy(Tirion *tirion);

float tirionAdd(Tirion *tirion, long i, float v);
float tirionDec(Tirion *tirion, long i);
float tirionInc(Tirion *tirion, long i);
float tirionSub(Tirion *tirion, long i, float v);

long tirionTag(Tirion *tirion, const char *format, ...);

void tirionD(const Tirion *tirion, const char *format, ...);
void tirionE(const Tirion *tirion, const char *format, ...);
void tirionV(const Tirion *tirion, const char *format, ...);

#endif
