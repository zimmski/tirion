package tirion;

import java.io.IOException;
import java.io.RandomAccessFile;
import java.nio.ByteOrder;
import java.nio.FloatBuffer;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.util.concurrent.locks.*;

public class Client {
	public static String TIRION_VERSION = "0.2";
	
	private static int floatSize = 4;
	private static String logPrefix = "[client]";
	private static int tirionTagSize = 513;
	
	private int count;
	private FloatBuffer metrics;
	private Lock metricLock;
	private boolean running;
	private String socket;
	private boolean verbose;
	
	public Client(String socket, boolean verbose) {
		this.count = 0;
		this.metrics = null;
		this.running = false;
		this.socket = socket;
		this.verbose = verbose;
	}
	
	// Init initializes the client
	public void init() {
		var err error

		if r, err := syscall.Setsid(); r == -1 {
			this.e("Cannot set new session and group id of process: %v");

			return err
		}

		this.v("Open unix socket to %s", c.socket)
		c.fd, err = net.Dial("unix", c.socket)

		if err != nil {
			if strings.HasSuffix(err.Error(), "use of closed network connection") || strings.HasSuffix(err.Error(), "no such file or directory") {
				this.e("Cannot open unix socket %s", c.socket);
			}

			return err
		} else {
			this.v("Request tirion protocol version v%s", Version)
			if err := c.send("tirion v" + Version + "\t" + c.PreferredMetricProtocoll); err != nil {
				this.e(err.Error())

				return err
			}

			m, err := c.receive()

			switch err {
			case nil:
				var t = strings.SplitN(m, "\t", 2)

				if len(t) == 1 || t[1] == "" {
					err := errors.New("Did not receive correct metric count and protocol URL");

					this.e(err.Error())

					return err
				}

				var metricCount, err = strconv.Atoi(t[0])

				if err != nil {
					this.e("Did not receive correct metric count");

					return err
				}

				u, err := url.Parse(t[1])

				if err != nil {
					this.e("Did not receive correct protocol URL");

					return err
				}

				this.v("Received metric count %d and protocol URL %v", metricCount, u)

				c.metricsCollector, err = collector.NewCollector(u.Scheme)

				if err != nil {
					this.e("Cannot create metric collector");

					return err
				}
				
				if _, err := os.Stat(u.Path); os.IsNotExist(err) {
					return errors.New(fmt.Sprintf("Cannot open mmap file: %v", err))
				}

				err = c.metricsCollector.InitClient(u, int32(metricCount))

				if err != nil {
					this.e("Cannot initialize metrics collector");

					return err
				}

				this.v("Initialized metric collector %s", u.Scheme);

				this.running = true;

				// we want to handle commands not in the main thread
				go c.handleCommands()
			case io.EOF:
				this.v("Unix socket got closed with EOF");

				return err
			default:
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					this.v("Unix socket suddenly got closed");
				}

				return err
			}
		}

		return nil
	}

	// Close uninitializes the client by closing all connections of the client.
	public void close() {
		this.running = false;

		this.mmapClose();

		if c.fd != nil {
			if err := c.fd.Close(); err != nil {
				return err
			}

			c.fd = nil
		}
	}

	// Destroy deallocates all data of the client.
	public void destroy() {
	}

	// Add adds a value to a metric
	public float add(int i, float v) {
		if (i < 0 || i >= this.count) {
			return 0.0f;
		}
		
		float f = 0.0f;
		
		this.metricLock.lock();
		
		try {
			if (this.metrics != null) {
				f = this.metrics.get(i) + v;
				
				this.metrics.put(i, f);
			}
		} finally {
			this.metricLock.unlock();
		}

		return f;
	}

	// Dec decrements a metric by 1.0
	public float dec(int i) {
		return this.sub(i, 1.0f);
	}

	// Inc increments a metric by 1.0
	public float inc(int i) {
		return this.add(i, 1.0f);
	}

	// Sub subtracts a value of a metric
	public float sub(int i, float v) {
		if (i < 0 || i >= this.count) {
			return 0.0f;
		}

		float f = 0.0f;
		
		this.metricLock.lock();
		
		try {
			if (this.metrics != null) {
				f = this.metrics.get(i) - v;
				
				this.metrics.put(i, f);
			}
		} finally {
			this.metricLock.unlock();
		}

		return f;
	}
	
	public boolean running() {
		return this.running;
	}

	// Tag sends a tag to the agent
	public void tag(String format, Object... args) {
		this.send(this.prepareTag(String.format("t" + format, args)));
	}
	
	// D outputs a Tirion debug message.
	public void d(String format, Object... args) {
		if (!this.verbose) {
			return;
		}
		
		System.err.print(this.logPrefix + "[debug] " + String.format(format, args) + "\n");
	}

	// E outputs a Tirion error message.
	public void e(String format, Object... args) {
		if (!this.verbose) {
			return;
		}
		
		System.err.print(this.logPrefix + "[error] " + String.format(format, args) + "\n");
	}

	// V outputs a Tirion verbose message.
	public void v(String format, Object... args) {
		if (!this.verbose) {
			return;
		}
		
		System.err.print(this.logPrefix + "[verbose] " + String.format(format, args) + "\n");
	}

	private void handleCommands() {
		this.v("Start listening to commands");
	
		for c.Running {
			var data, err = c.receive()
	
			switch err {
			case nil:
				com := data[0]
	
				switch com {
				default:
					this.e("Unknown command '%c'", com)
				}
			case io.EOF:
				this.v("Unix socket got closed with EOF")
	
				c.Running = false
			default:
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					if c.Running {
						this.v("Unix socket suddenly got closed")
					}
				} else {
					this.e("%v", err)
				}
	
				c.Running = false
	
				break
			}
		}
	
		this.v("Stop listening to commands")
	}
	
	private void mmapOpen(String filename, int metricCount) throws IOException {
		RandomAccessFile file = new RandomAccessFile(filename, "rw");
        MappedByteBuffer buffer = file.getChannel().map(FileChannel.MapMode.READ_WRITE, 0, floatSize * metricCount);
        
        buffer.limit(floatSize * metricCount);
        buffer.order(ByteOrder.LITTLE_ENDIAN);
        
        //buffer.force();
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
		if (tag.length() > tirionTagSize) {
			tag = tag.substring(0, tirionTagSize);
		}

		return tag.replace("\n", " ");
	}
	
	private String receive() {
		var buf = make([]byte, 4096)

		nr, err := t.fd.Read(buf)

		if err != nil {
			return "", err
		}

		return strings.Trim(string(buf[0:nr]), "\n"), nil
	}

	private void send(String msg) {
		_, err := t.fd.Write([]byte(msg + "\n"))

		return err
	}
}
