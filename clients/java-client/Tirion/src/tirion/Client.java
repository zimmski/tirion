package tirion;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.RandomAccessFile;
import java.nio.ByteOrder;
import java.nio.FloatBuffer;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.util.LinkedList;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

import com.etsy.net.*;

public class Client {
	public static String TirionVersion = "0.2";

	private static int FloatSize = 4;
	private static String LogPrefix = "[client]";
	private static int TirionBufferSize = 4096;
	private static int TirionTagSize = 513;

	private int count;
	private Thread handleCommands;
	private FloatBuffer metrics;
	private Lock metricLock;
	private UnixDomainSocketClient net;
	private InputStream netIn;
	private LinkedList<String> netInQueue;
	private OutputStream netOut;
	private boolean running;
	private String socket;
	private boolean verbose;

	public Client(String socket, boolean verbose) {
		this.count = 0;
		this.running = false;
		this.socket = socket;
		this.verbose = verbose;
	}

	// Init initializes the client
	public void init() throws Exception {
		/*
		 * TODO find out how to do this in java... if r, err :=
		 * syscall.Setsid(); r == -1 {
		 * this.e("Cannot set new session and group id of process: %v");
		 *
		 * return err }
		 */

		this.v("Open unix socket to %s", this.socket);
		this.net = new UnixDomainSocketClient(this.socket, JUDS.SOCK_STREAM);
		this.netIn = this.net.getInputStream();
		this.netOut = this.net.getOutputStream();

		this.netInQueue = new LinkedList<String>();

		this.v("Request tirion protocol version v%s", tirion.Client.TirionVersion);
		this.send("tirion v" + tirion.Client.TirionVersion + "\tmmap");

		String[] t = this.receive().split("\t");

		if (t.length < 2 || t[1].length() == 0) {
			throw new Exception("Did not receive correct metric count and mmap filename");
		}

		try {
			this.count = Integer.parseInt(t[0]);
		} catch (NumberFormatException e) {
			this.e("Did not receive correct metric count");

			throw e;
		}

		this.metricLock = new ReentrantLock();

		if (!t[1].startsWith("mmap://")) {
			throw new Exception("Did not receive correct mmap filename");
		}

		String mmapFilename = t[1].substring(7);

		this.v("Received metric count %d and mmap filename %s", this.count, mmapFilename);

		this.mmapOpen(mmapFilename);

		this.v("Initialized metric collector mmap");

		this.running = true;

		// we want to handle commands not in the main thread
		this.handleCommands = new Thread(new HandleCommands());
		this.handleCommands.start();
	}

	// Close uninitializes the client by closing all connections of the client.
	public void close() throws IOException {
		this.running = false;

		this.mmapClose();

		if (this.net != null) {
			this.netIn.close();
			this.netOut.close();
			this.net.close();

			netInQueue.clear();

			this.netIn = null;
			this.netOut = null;
			this.net = null;
		}

		if (this.handleCommands != null) {
			try {
				this.handleCommands.join();
			} catch (InterruptedException e) {
			}
			this.handleCommands = null;
		}
	}

	// Destroy deallocates all data of the client.
	public void destroy() {
		this.metricLock = null;
		this.netInQueue = null;
	}

	// Add adds a value to a metric
	public float add(int i, float v) {
		if (i < 0 || i >= this.count) {
			return 0.0f;
		}

		float ret = 0.0f;

		this.metricLock.lock();

		try {
			if (this.metrics != null) {
				ret = this.metrics.get(i) + v;

				this.metrics.put(i, ret);
			}
		} finally {
			this.metricLock.unlock();
		}

		return ret;
	}

	// Dec decrements a metric by 1.0
	public float dec(int i) {
		return this.add(i, -1.0f);
	}

	// Inc increments a metric by 1.0
	public float inc(int i) {
		return this.add(i, 1.0f);
	}

	// Sub subtracts a value of a metric
	public float sub(int i, float v) {
		return this.add(i, -v);
	}

	public boolean running() {
		return this.running;
	}

	// Tag sends a tag to the agent
	public void tag(String format, Object... args) throws IOException {
		this.send(this.prepareTag(String.format("t" + format, args)));
	}

	private void m(String type, String format, Object... args) {
		if (!this.verbose) {
			return;
		}

		System.err.print(Client.LogPrefix + "[" + type + "] " + String.format(format, args) + "\n");
	}

	// D outputs a Tirion debug message.
	public void d(String format, Object... args) {
		this.m("debug", format, args);
	}

	// E outputs a Tirion error message.
	public void e(String format, Object... args) {
		this.m("error", format, args);
	}

	// V outputs a Tirion verbose message.
	public void v(String format, Object... args) {
		this.m("verbose", format, args);
	}

	private void mmapOpen(String filename) throws IOException {
		RandomAccessFile file = new RandomAccessFile(filename, "rw");
		MappedByteBuffer buffer = file.getChannel().map(FileChannel.MapMode.READ_WRITE, 0, FloatSize * this.count);

		buffer.limit(FloatSize * this.count);
		buffer.order(ByteOrder.LITTLE_ENDIAN);

		// buffer.force();
		buffer.load();

		this.metrics = buffer.asFloatBuffer();

		file.close();
	}

	private void mmapClose() {
		this.metricLock.lock();

		try {
			this.metrics = null;
		} finally {
			this.metricLock.unlock();
		}
	}

	private String prepareTag(String tag) {
		if (tag.length() > TirionTagSize) {
			tag = tag.substring(0, TirionTagSize);
		}

		return tag.replace("\n", " ");
	}

	private synchronized String receive() throws Exception {
		if (this.netInQueue.size() != 0) {
			return this.netInQueue.pop();
		} else {
			while (this.netInQueue.size() == 0) {
				byte[] buf = new byte[Client.TirionBufferSize];

				int ret = this.netIn.read(buf, 0, Client.TirionBufferSize - 1);

				if (ret == -1) {
					throw new Exception("End of the stream");
				}

				String s = new String(buf, 0, ret);

				for (String i : s.split("\n")) {
					if (i.compareTo("") != 0) {
						this.netInQueue.push(i);
					}
				}
			}

			return this.netInQueue.pop();
		}
	}

	private void send(String msg) throws IOException {
		this.netOut.write((msg + "\n").getBytes());
	}

	private class HandleCommands implements Runnable {
		@Override
		public void run() {
			Client.this.v("Start listening to commands");

			while (Client.this.running) {
				Exception err = null;
				String rec = null;

				try {
					rec = Client.this.receive();
				} catch (Exception t) {
					err = t;
				}

				if (err == null) {
					char com = rec.charAt(0);

					switch (com) {
					default:
						Client.this.e("Unknown command '%c'", com);
					}
				} else {
					Client.this.e("Unix socket error: " + err);

					Client.this.running = false;
				}
			}

			Client.this.v("Stop listening to commands");
		}
	}
}
