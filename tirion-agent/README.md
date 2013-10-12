# tirion-agent

The tirion-agent is the connective link between the client (your application) and the tirion-server which saves all run data like metrics and tags. If you use the precompiled Tirion binaries you can use the agent right away otherwise you have to [build Tirion](/#how-to-build-tirion) first. This README focuses on how you can use the tirion-agent and not how it works, have a look at [“How does Tirion work?”](/#how-does-tirion-work) if you want to know more.

## CLI arguments

```
  -exec="": Execute this command
  -exec-arguments="": Arguments for the command
  -help=false: Show this help
  -interval=250: How often metrics are fetched (in milliseconds)
  -limit-memory=0: Limit the memory of the program and its children (in MB)
  -limit-memory-interval=5: Interval for checking the memory limit (in milliseconds)
  -limit-time=0: Limit the runtime of the program (in seconds)
  -metrics-filename="": Definition of needed program metrics
  -name="": The name of this run (defaults to exec)
  -pid=-1: PID of program which should be monitored
  -send-interval=5: How often data is pushed to the server (in seconds)
  -server="": Server address for agent<-->server communication
  -socket="": Unix socket path for client<-->agent communication
  -sub-name="": The subname of this run
  -verbose=false: Verbose output of what is going on
```

The <code>-metrics-filename</code> argument is required as well as <code>-pid</code> which monitors an existing process or <code>-exec</code> which starts a new one. To allow communication between client and agent, and therefore the exchange of internal metrics, the <code>-socket</code> argument is needed.

Usage:

* tirion-agent -pid <pid> -metrics-filename <json file> [other options]
* tirion-agent -exec <program> -metrics-filename <json file> [other options]

If no <code>-server</code> argument is used, the agent will write all data to STDOUT formatted as CSV.

The arguments <code>-limit-memory</code> and <code>-limit-time</code> can only be used with <code>-exec</code> as the agent must have control over the monitored program.

## Example arguments

* Monitor the process with the PID 2342 using the metrics file in folder/metrics.json
	<pre><code>tirion-agent -pid 2342 -metrics-filename folder/metrics.json</code></pre>

* Execute and monitor the program [go-mandelbrot](/examples/go-mandelbrot) with verbose output while sending data to a server
	<pre><code>tirion-agent -exec go-mandelbrot -metrics-filename folder/metrics.json -verbose -server "localhost:9000"</code></pre>

* The same arguments as before but with communication between client and agent through a socket
	<pre><code>tirion-agent -exec go-mandelbrot -metrics-filename folder/metrics.json -verbose -server "localhost:9000" -socket /tmp/tirion.sock</code></pre>

* Execute the program [c-client](/clients/c-client) with the arguments “-v” and “-r 5” while fetching metrics every 10 milliseconds and sending metrics every other second to a server
	<pre><code>tirion-agent -exec c-client -exec-arguments “-v -r 5” -metrics-filename folder/metrics.json -server "localhost:9000" -send-interval</code></pre>

## Limits

All limit arguments restrict the running process of the monitored program by some metric. If a limit is reached, the process is instantaneously killed by sending a <code>SIGKILL</code> signal to the process’s group id which also kills all child processes of the parent process.

* <code>-limit-memory</code>

    Limits the monitored process by the accumulated <code>Resident Set Size</code> (RSS) of the parent and child processes. The interval time of the check can be adapted by the argument <code>-limit-memory-interval</code>. This implies that the monitored processes can in fact disobey the given limit until the check reoccurs.

* <code>-limit-time</code>

    Limits the monitored process by the runtime measured in real seconds. This means that the agent does not oversee the CPU time of the process nor any other CPU related metric.

