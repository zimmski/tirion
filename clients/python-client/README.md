# Tirion Python client library

## How do I use Tirion in my Python application?

To use the Tirion Python client library install the <code>tirion</code> package via the provided <code>setup.py</code> and include the <code>tirion.client</code> module in your source code. Both can be found in the precompiled binary archive.

```bash
sudo python lib/python/setup.py install
```

Otherwise you can just copy the [tirion](/clients/python-client/tirion) folder into your project and include the <code>tirion.client</code> module in your source code.

After that, you have to instantiate a client object with <code>tirion.client.Client(socket_filename, verbose)</code>. The socket is needed for the client <-> agent communication. The verbose parameter states whether the library should print verbose output or not.

To initialize the client object <code>init()</code> must be called with the object itself. If the function raises no exception, the initialization was successful and the object can be used to set and modify internal metrics and send tags.

Internal metric indices are defined via a [metric file](/#metric-file) which is fed to the agent.

The following functions can be used to interact with metrics and tags. Have a look at the [API](#api) section for a more complete documentation.

* <code>get(index)</code>
* <code>set(index, value)</code>
* <code>add(index, value)</code>
* <code>dec(index)</code>
* <code>inc(index)</code>
* <code>sub(index, value)</code>
* <code>tag(format_string, *args)</code>

As the agent lives as long as the client program lives, there is no need to prematurely close the connection to the agent. If you still want (or need) to close the connection between client and agent the function <code>close()</code> must be called and <code>destroy()</code> to free allocated memory.

Have a look at the [example program](#example-usage) for a more complete example otherwise here is a small one:

```python
import tirion.client
...

tirion_client = tirion.client.Client("/tmp/tirion.sock", True)

try:
	tirion_client.init()

	tirion_client.tag("start loop")

	for i in range(0, 10):
		tirion_client.add(2, 0.5)
		tirion_client.inc(1)

	tirion_client.tag("end loop")
except:
	raise
finally:
	tirion_client.close()
	tirion_client.destroy()
```

## Multi-process applications

Due to the [architecture of Tirion's agent](/#how-does-tirion-work) it is very important that the initialization of the Tirion object must occur before forking new child processes. Otherwise, they would not inherit the group id of the parent process which is needed for [restricting](/tirion-agent#limits) and completely killing the monitored process.

## API

Please have a look at the [Python API documentation](https://rawgithub.com/zimmski/tirion/master/clients/python-client/doc/html/classtirion_1_1client_1_1Client.html) or [client.py](/clients/python-client/tirion/client.py) for a complete API overview of Tirion's Python client library.

## Example usage

There is a complete example in [python_client.py](/clients/python-client/python_client.py) on how to use the library and its functions.
