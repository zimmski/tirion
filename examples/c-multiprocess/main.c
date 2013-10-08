#include <errno.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

#include "../../clients/c-client/tirion.h"

enum Metrics {
	MetricChilds,
	MetricAllocated
};

long parseLongArgument() {
	char *endptr;

	long r = strtol(optarg, &endptr, 10);

	if ((errno == ERANGE && (r == INT_MAX || r == INT_MIN)) || (errno != 0 && r == 0)) {
		return -1;
	}

	return r;
}

int main (int argc, char **argv) {
	int flagErrors = 0;

	long flag_childs = 10;
	bool flagHelp = false;
	long flag_mb_per_child = 1;

	char c;

	while ((c = getopt(argc, argv, ":hc:m:")) != -1) {
		switch(c) {
		case 'c':
			flag_childs = parseLongArgument();

			if (flag_childs == -1) {
				printf("ERROR: child count is not a number\n");

				flagErrors++;
			} else if (flag_childs < 1) {
				printf("ERROR: child count must be greater than 0");

				flagErrors++;
			}
			break;
		case 'h':
			flagHelp = true;
			break;
		case 'm':
			flag_mb_per_child = parseLongArgument();

			if (flag_mb_per_child == -1) {
				printf("ERROR: MB per child is not a number\n");

				flagErrors++;
			} else if (flag_mb_per_child < 1) {
				printf("ERROR: MB per child must be greater than 0\n");

				flagErrors++;
			}
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
	if (flagHelp || flagErrors > 0) {
		printf("Tirion multi process example client v%s\n", TIRION_VERSION);
		printf("usage: %s [options]\n", argv[0]);
		printf("  -c 10: Children count\n");
		printf("  -h false: Show this help\n");
		printf("  -m 1: Count of one MB allocations per child\n");

		if (! flagHelp) {
			printf("Wrong arguments\n");
		}

		return 1;
	}

	Tirion *tirion = tirionNew("/tmp/tirion.sock", true);

	if (tirionInit(tirion) == TIRION_OK) {
		pid_t* ids = malloc(sizeof(pid_t) * flag_childs);

		int i = 0;
		for (; i < flag_childs; i++) {
			ids[i] = fork();

			if (ids[i] == 0) { // child
				tirionV(tirion, "start child %d", i);
				tirionInc(tirion, MetricChilds);

				char **a = malloc(sizeof(char*) * flag_mb_per_child);

				int j = 0;
				for (; j < flag_mb_per_child; j++) {
					a[j] = malloc(sizeof(char) * 1024 * 1024);

					tirionV(tirion, "Child %d allocated another MB", i);
					tirionInc(tirion, MetricAllocated);

					usleep(100000); // sleep for 100ms
				}

				for (; j < flag_mb_per_child; j++) {
					free(a[j]);
					tirionDec(tirion, MetricAllocated);
				}

				free(a);

				tirionV(tirion, "exit child %d", i);
				tirionDec(tirion, MetricChilds);

				exit(0);
			} else if (ids[i] < 0) { // fork failed
				tirionE(tirion, "Failed to fork child #%d", i);

				exit(1);
			}
		}

		for (i = 0; i < flag_childs; i++) {
			int status;
			int w = waitpid(ids[i], &status, 0);

			if (w == -1) {
				tirionE(tirion, "Wait failed for child #%d", i);
            }
		}

		free(ids);
	} else {
		printf("ERROR: Cannot initialize Tirion\n");
	}

	tirionClose(tirion);

	tirionV(tirion, "Stopped");

	tirionDestroy(tirion);

	return 0;
}
