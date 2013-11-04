package tirion;

public class main {
	public static void main(final String[] args) {
		int flagErrors = 0;

		boolean flagHelp = false;
		int flagRuntime = 5;
		String flagSocket = "/tmp/tirion.sock";
		boolean flagVerbose = false;

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
		if (flagSocket.compareTo("") == 0 || flagHelp || flagErrors > 0) {
			printf("Tirion C example client v%s\n", tirion.Client.TIRION_VERSION);
			printf("usage: %s [options]\n", args[0]);
			printf("  -h false: Show this help\n");
			printf("  -r 5: Runtime of the example client in seconds\n");
			printf("  -s \"/tmp/tirion.sock\": Unix socket path for client<-->agent communication\n");
			printf("  -v: Verbose output of what is going on\n");

			if (! flagHelp) {
				printf("Wrong arguments\n");
			}

			return 1;
		}
		
		tirion.Client t = new tirion.Client(flagSocket, flagVerbose);
		
		try {
			t.init();
		} catch (Exception e) {
			System.out.printf("ERROR: Cannot initialize Tirion " + e + "\n");
			
			System.exit(1);
		}
		
		new java.util.Timer().schedule( 
		        new java.util.TimerTask() {
		            @Override
		            public void run() {
		                t.d("Program ran for %d seconds, this is enough data.", flagRuntime);

		            	t.close();
		            }
		        }, 
		        1000 * flagRuntime 
		);

		while (t.running()) {
			float r = t.inc(0);
			t.dec(1);
			t.add(2, 0.3f);
			t.sub(3, 0.3f);

			Thread.sleep(10);

			if (r % 20.0f == 0.0f) {
				t.tag("index 0 is %f", r);
			}
		}

		t.close();

		t.v("Stopped");

		t.destroy();
	}

}
