# Tirion C client library

## How do I use Tirion in my C application?

To use the Tirion C client library just copy [tirion.c](/clients/c-client/tirion.c) and [tirion.h](/clients/c-client/tirion.h) into your project and include tirion.h in your source code.

After that, you have to instantiate a client object with the function <code>tirionNew(const char *socket, bool verbose)</code>. The socket is needed for the client <-> agent communication. The verbose parameter states whether the library should print verbose output or not.

To initialize the client object <code>tirionInit(Tirion *tirion)</code> must be called with the object itself. If the function returns the constant TIRION_OK, the initialization was successful and the object can be used to set and modify internal metrics and send tags.

Internal metric indices are defined via a [metric file](/#metric-file) which is fed to the agent.

The following functions can be used to interact with metrics and tags. Have a look at the [API](#api) section for a more complete documentation.

* <code>tirionAdd(Tirion *tirion, int i, float v)</code>
* <code>tirionDec(Tirion *tirion, int i)</code>
* <code>tirionInc(Tirion *tirion, int i)</code>
* <code>tirionSub(Tirion *tirion, int i, float v)</code>
* <code>tirionTag(Tirion *tirion, const char *format, ...)</code>

As the agent lives as long as the client program lives, there is no need to prematurely close the connection to the agent. If you still want (or need) to close the connection between client and agent the function <code>tirionClose(Tirion *tirion)</code> must be called and <code>tirionDestroy(Tirion *tirion)</code> to free allocated memory.

Have a look at the [example program](#example-usage) for a more complete example otherwise here is a small one:

```c
#include "tirion.h"
...
Tirion *t = tirionNew("/tmp/tirion.socket", true);

if (tirionInit(t) == TIRION_OK) {
	tirionTag(t, "start loop");

	int i = 0;
	for (; i < 10; i++) {
		tirionAdd(t, 2, 0.5);
		tirionInc(t, 1);
	}

	tirionTag(t, "end loop");
}

tirionClose(t);
tirionDestroy(t);
```

## API

Please have a look at [tirion.h](/clients/c-client/tirion.h) for a complete API overview of Tirion's c client library.

## Example usage

There is a complete example in [main.c](/clients/c-client/main.c) on how to use the library and its functions.
