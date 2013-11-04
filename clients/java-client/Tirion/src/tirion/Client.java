package tirion;

import java.io.IOException;
import java.io.RandomAccessFile;
import java.nio.ByteOrder;
import java.nio.FloatBuffer;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.util.concurrent.locks.*;

public class Client {
	private static int floatSize = 4;
	private static String logPrefix = "[client]";
	
	private int count;
	private FloatBuffer metrics;
	private Lock metricLock;
	private String socket;
	private boolean verbose;
	
	public Client(String socket, boolean verbose) {
		this.count = 0;
		this.metrics = null;
		this.socket = socket;
		this.verbose = verbose;
	}
	
	// Init initializes the client
	public void init() {
		var err error

		if r, err := syscall.Setsid(); r == -1 {
			c.E("Cannot set new session and group id of process: %v")

			return err
		}

		c.V("Open unix socket to %s", c.socket)
		c.fd, err = net.Dial("unix", c.socket)

		if err != nil {
			if strings.HasSuffix(err.Error(), "use of closed network connection") || strings.HasSuffix(err.Error(), "no such file or directory") {
				c.E("Cannot open unix socket %s", c.socket)
			}

			return err
		} else {
			c.V("Request tirion protocol version v%s", Version)
			if err := c.send("tirion v" + Version + "\t" + c.PreferredMetricProtocoll); err != nil {
				c.E(err.Error())

				return err
			}

			m, err := c.receive()

			switch err {
			case nil:
				var t = strings.SplitN(m, "\t", 2)

				if len(t) == 1 || t[1] == "" {
					err := errors.New("Did not receive correct metric count and protocol URL")

					c.E(err.Error())

					return err
				}

				var metricCount, err = strconv.Atoi(t[0])

				if err != nil {
					c.E("Did not receive correct metric count")

					return err
				}

				u, err := url.Parse(t[1])

				if err != nil {
					c.E("Did not receive correct protocol URL")

					return err
				}

				c.V("Received metric count %d and protocol URL %v", metricCount, u)

				c.metricsCollector, err = collector.NewCollector(u.Scheme)

				if err != nil {
					c.E("Cannot create metric collector")

					return err
				}
				
				if _, err := os.Stat(u.Path); os.IsNotExist(err) {
					return errors.New(fmt.Sprintf("Cannot open mmap file: %v", err))
				}

				err = c.metricsCollector.InitClient(u, int32(metricCount))

				if err != nil {
					c.E("Cannot initialize metrics collector")

					return err
				}

				c.V("Initialized metric collector %s", u.Scheme)

				c.Running = true

				// we want to handle commands not in the main thread
				go c.handleCommands()
			case io.EOF:
				c.V("Unix socket got closed with EOF")

				return err
			default:
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					c.V("Unix socket suddenly got closed")
				}

				return err
			}
		}

		return nil
	}

	// Close uninitializes the client by closing all connections of the client.
	public void close() {
		c.Running = false

		if c.metricsCollector != nil {
			f := C.CString(c.filename)
					defer C.free(unsafe.Pointer(f))

					var cr C.char

					if c.create {
						cr = C.char(1)
					} else {
						cr = C.char(0)
					}

					if C.mmapClose(c.addr, f, cr, C.long(c.count)) != 0 {
						return errors.New("Mmap close error")
					}

					return nil

			c.metricsCollector = nil
		}

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
		
		float f;
		
		this.metricLock.lock();
		
		try {
			f = this.metrics.get(i) + v;
			
			this.metrics.put(i, f);
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

		float f;
		
		this.metricLock.lock();
		
		try {
			f = this.metrics.get(i) - v;
			
			this.metrics.put(i, f);
		} finally {
			this.metricLock.unlock();
		}

		return f;
	}

	// Tag sends a tag to the agent
	public tag(String format, Object... args) {
		c.send(PrepareTag(String.format("t" + format, args)))
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
		c.V("Start listening to commands")
	
		for c.Running {
			var data, err = c.receive()
	
			switch err {
			case nil:
				com := data[0]
	
				switch com {
				default:
					c.E("Unknown command '%c'", com)
				}
			case io.EOF:
				c.V("Unix socket got closed with EOF")
	
				c.Running = false
			default:
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					if c.Running {
						c.V("Unix socket suddenly got closed")
					}
				} else {
					c.E("%v", err)
				}
	
				c.Running = false
	
				break
			}
		}
	
		c.V("Stop listening to commands")
	}
	
	private void mmapOpen(String filename, int metricCount) throws IOException {
		RandomAccessFile file = new RandomAccessFile(filename, "rw");
        MappedByteBuffer buffer = file.getChannel().map(FileChannel.MapMode.READ_WRITE, 0, floatSize * metricCount);
        
        buffer.limit(floatSize * metricCount);
        buffer.order(ByteOrder.LITTLE_ENDIAN);
        
        //buffer.force();
        buffer.load();
        
        this.metrics = buffer.asFloatBuffer();
	}

}
