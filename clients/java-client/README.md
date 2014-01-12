# Tirion Java client library

## How do I use Tirion in my Java application?

To use the Tirion Java client library include the tirion.jar file in your project and import the namespace "tirion.*" in your source code. The jar file can be found in the precompiled binary archive.

After that, you have to instantiate a client object with the constructor <code>tirion.Client(String socket, boolean verbose)</code>. The socket is needed for the client <-> agent communication. The verbose parameter states whether the library should print verbose output or not.

To initialize the client object <code>init()</code> must be called with the object itself. After the initialization the object can be used to set and modify internal metrics and send tags.

Internal metric indices are defined via a [metric file](/#metric-file) which is fed to the agent.

The following functions can be used to interact with metrics and tags. Have a look at the [API](#api) section for a more complete documentation.

* <code>get(int i)</code>
* <code>set(int i, float v)</code>
* <code>add(int i, float v)</code>
* <code>dec(int i)</code>
* <code>inc(int i)</code>
* <code>sub(int i, float v)</code>
* <code>tag(String format, Object... args)</code>

As the agent lives as long as the client program lives, there is no need to prematurely close the connection to the agent. If you still want (or need) to close the connection between client and agent the function <code>close()</code> must be called and <code>destroy()</code> to cleanup unneeded objects.

Have a look at the [example program](#example-usage) for a more complete example otherwise here is a small one:

```java
import tirion.*;
...

tirion.Client t = null;

try {
	tirion.Client t = new tirion.Client("/tmp/tirion.sock", true);

	t.init();

	t.tag("start loop");

	int i = 0;
	for (; i < 10; i++) {
		t.add(2, 0.5);
		t.inc(1);
	}

	t.tag("end loop");
} catch(Exception e) {
	// ...
} finally {
	if (t != null) {
		t.close();
		t.destroy();
	}
}
```

## Client <-> agent communication

The Tirion Java library uses a Unix Domain Socket via the excellent [JUDS library](https://github.com/mcfunley/juds) to communicate with the corresponding agent. Internal metrics of the client application are exchanged through a memory mapped file (mmap) which is attached to a float array. Changing metric values is therefore very fast but still a synchronized operation.

## Multi-process applications

Due to the [architecture of Tirion's agent](/#how-does-tirion-work) it is very important that the initialization of the Tirion object must occur before forking new child processes. Otherwise, they would not inherit the group id of the parent process which is needed for [restricting](/tirion-agent#limits) and completely killing the monitored process.

## API

Please have a look at the [Java API documentation](https://rawgithub.com/zimmski/tirion/master/clients/java-client/Tirion/doc/tirion/Client.html) or [Client.java](/clients/java-client/Tirion/src/tirion/Client.java) for a complete API overview of Tirion's Java client library.

## Example usage

There is a complete example in [Main.java](/clients/java-client/Tirion/src/tirion/Main.java) on how to use the library and its functions.
