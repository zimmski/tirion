package tirion;

import java.io.IOException;

import org.apache.commons.cli.*;

public class Main {
	private static tirion.Client t;
	private static int runtime;

	private static Option getCommandlineOption(String opt, String longOpt, String description, boolean hasArg, String argName) {
		Option option = new Option(opt, hasArg, description);

		if (longOpt != null) {
			option.setLongOpt(longOpt);
		}

		if (hasArg && argName != null) {
			option.setArgName(argName);
		}

		return option;
	}

	public static void main(final String[] args) throws InterruptedException, IOException {
		boolean help = false;
		runtime = 5;
		String socket = "/tmp/tirion.sock";
		boolean verbose = false;

		CommandLineParser argParser = new PosixParser();

		Options options = new Options();

		options.addOption(getCommandlineOption("h", "help", "Print help.", false, null));
		options.addOption(getCommandlineOption("r", "runtime", String.format("Runtime of the example client in seconds. Default is %d.", runtime), true, "INTEGER"));
		options.addOption(getCommandlineOption("s", "socket", String.format("Unix socket path for client<-->agent communication. Default is %s.", socket), true, "FILE"));
		options.addOption(getCommandlineOption("v", "verbose", "Enable verbose output.", false, null));

		try {
			CommandLine commandLine = argParser.parse(options, args);

			if (commandLine.hasOption("h")) {
				help = true;
			}
			if (commandLine.hasOption("r")) {
				try {
					runtime = Integer.parseInt(commandLine.getOptionValue("r"));
				}
				catch (NumberFormatException e) {
					throw new RuntimeException("runtime is not a number");
				}
			}
			if (commandLine.hasOption("s")) {
				socket = commandLine.getOptionValue("o");
			}
			if (commandLine.hasOption("v")) {
				verbose = true;
			}
		} catch (ParseException e) {
			System.out.println("Unexpected parse exception " + e.getMessage());

			System.exit(1);
		}

		if (help) {
			HelpFormatter helpFormatter = new HelpFormatter();

			helpFormatter.printHelp("Tirion Java example client v" + tirion.Client.TirionVersion, options);

			System.exit(1);
		}

		t = new tirion.Client(socket, verbose);

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
					t.d("Program ran for %d seconds, this is enough data.", runtime);

					try {
						t.close();
					} catch (IOException e) {
					}
				}
			},
			1000 * runtime
		);

		while (t.running()) {
			final float r = t.inc(0);
			t.dec(1);
			t.add(2, 0.3f);
			t.sub(3, 0.3f);
			t.set(4, t.get(4) + 4);

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
