# Tirion Go client

## How do I use Tirion in my Go application?

To use the Tirion Go client library copy the package folder of the precompiled binary archive into your pkg folder or just fetch Tirion into your Go path.

```bash
go get github.com/zimmski/tirion
```

and include the <code>github.com/zimmski/tirion</code> package.

```go
import "github.com/zimmski/tirion"
```

After that, you have to instantiate a client object with the function <code>NewTirionClient(socket string, verbose bool)</code>. The socket is needed for the client <-> agent communication. The verbose parameter states whether the library should print verbose output or not.

To initialize the client object <code>Init()</code> must be called with the object itself. If the function returns no error, the initialization was successful and the object can be used to set and modify internal metrics and send tags.

Internal metric indices are defined via a [metric file](/#metric-file) which is fed to the agent.

The following functions can be used on the object to interact with metrics and tags. Have a look at the [API](#api) section for a more complete documentation.

* <code>Add(i int, v float32)</code>
* <code>Dec(i int)</code>
* <code>Inc(i int)</code>
* <code>Sub(i int, v float32)</code>
* <code>Tag(format string, a ...interface{})</code>

As the agent lives as long as the client program lives, there is no need to prematurely close the connection to the agent. If you still want (or need) to close the connection between client and agent the function <code>Close()</code> must be called and <code>Destroy()</code> to free allocated objects.

Have a look at the [example program](#example-usage) for a more complete example otherwise here is a small one:

```go
import "github.com/zimmski/tirion"
...
t := tirion.NewTirionClient("/tmp/tirion.socket", true)

if t.Init() == nil {
	t.Tag("start loop")

	for i := 0; i < 10; i++ {
		t.Add(2, 0.5);
		t.Inc(1);
	}

	t.Tag("end loop")
}

t.Close()
t.Destroy()
```

## Multi-process applications

Due to the [architecture of Tirion's agent](/#how-does-tirion-work) it is very important that the initialization of the Tirion object must occur before forking new child processes. Otherwise, they would not inherit the group id of the parent process which is needed for [restricting](/tirion-agent#limits) and completely killing the monitored process.

## API

Please have a look at [client.go](/client.go) for a complete API overview of Tirion's Go client library.

## Example usage

There is a complete example in [main.go](/clients/go-client/main.go) on how to use the library and its functions.
