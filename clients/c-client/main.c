#include <errno.h>
#include <limits.h>
#include <math.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "tirion.h"

typedef struct threadAfterArgsStruct {
	int runtime;
	Tirion *tirion;
} threadAfterArgs;

void *threadAfter(void *ptr) {
	threadAfterArgs *args = (threadAfterArgs*)ptr;

	sleep(args->runtime);

	tirionD(args->tirion, "Program ran for %d seconds, this is enough data.", args->runtime);

	args->tirion->running = false;

	return NULL;
}

int main (int argc, char **argv) {
	int flagErrors = 0;

	bool flagHelp = false;
	int flagRuntime = 5;
	char *flagSocket = "/tmp/tirion.sock";
	bool flagVerbose = false;

	char c;
	char *endptr;

	while ((c = getopt(argc, argv, ":hr:s:v")) != -1) {
		switch(c) {
		case 'h':
			flagHelp = true;
			break;
		case 'r':
			flagRuntime = strtol(optarg, &endptr, 10);

			if ((errno == ERANGE && (flagRuntime == INT_MAX || flagRuntime == INT_MIN)) || (errno != 0 && flagRuntime == 0)) {
				printf("ERROR: runtime is no number");

				flagErrors++;
			}
			break;
		case 's':
			flagSocket = optarg;
			break;
		case 'v':
			flagVerbose = true;
			break;
		case ':':
			printf("Option -%c requires an operand\n", optopt);

			flagErrors++;
			break;
		case '?':
			printf("Unrecognized option: '-%c'\n", optopt);

			flagErrors++;
			break;
		}
	}
	if (strcmp(flagSocket, "") == 0 || flagHelp || flagErrors > 0) {
		printf("Tirion C example client v%s\n", TIRION_VERSION);
		printf("usage: %s [options]\n", argv[0]);
		printf("  -h false: Show this help\n");
		printf("  -r 5: Runtime of the example client in seconds\n");
		printf("  -s \"/tmp/tirion.sock\": Unix socket path for client<-->agent communication\n");
		printf("  -v: Verbose output of what is going on\n");
		printf("Wrong arguments\n");

		return 1;
	}

	Tirion *tirion = tirionNew(flagSocket, flagVerbose);

	if (tirionInit(tirion) == TIRION_OK) {
		pthread_t tAfter;
		threadAfterArgs tAfterArgs = { flagRuntime, tirion };
		int rAfter = pthread_create(&tAfter, NULL, threadAfter, (void*) &tAfterArgs);
		if (rAfter != 0) {
			tirionE(tirion, "Failed creating After thread");
		} else {
			while (tirion->running) {
				float r = tirionInc(tirion, 0);
				tirionDec(tirion, 1);
				tirionAdd(tirion, 2, 0.3);
				tirionSub(tirion, 3, 0.3);

				usleep(10 * 1000);

				if (fmod(r, 20.0) == 0.0) {
					tirionTag(tirion, "index 0 is %f", r);
				}
			}

			pthread_join(tAfter, NULL);
		}
	} else {
		printf("ERROR: Cannot initialize Tirion\n");
	}

	tirionClose(tirion);

	tirionV(tirion, "Stopped");

	tirionDestroy(tirion);

	return 0;
}
