# Tirion

## What is Tirion?

Tirion ([Quenya](https://en.wikipedia.org/wiki/Quenya) for "watchtower") is a complete infrastructure for monitoring and overseeing applications and their metrics. In Tirion’s case this means that the execution of an application is overseen and restricted regarding resources and monitored by fetching OS process metrics like CPU, memory and IO usage as well as internal metrics which are produced by the application itself. Unlike other monitoring solutions Tirion is neither a profiler nor a statistical profiler (sampler) and its purpose is not monitoring for the intention of sending notifications if something goes wrong nor does Tirion do continuous monitoring for statistical purposes. Instead, single runs (note: a run is an execution of an application from the startup to exiting) are monitored for the purpose of comparing different configurations and versions of the application while overseeing resource limits of the running process.

## How does Tirion work?

Tirion consists of four components:

* A client library which is included and used by an application.
* An agent which receives and then aggregates data from exactly one client.
* A server which receives and then saves data from many agents.
* Clients who fetch data from a server to compare and analyse it.

![alt text](/doc/Architecture.png "Tirion's architecture")

The application, which should be monitored, must include the language specific client library. After the client object has been successfully initialized, it can be used to set and modify internal metrics of the application. These metrics are arbitrary definable by the programmers of the application.

An agent lives only for a single application run of the client and is therefore dependent on the lifetime of the application itself. There are two different modes to monitor an execution of an application which affects the control of the agent over the execution. Either the application is already running, which means that the agent has no control over the resource limits of the run, or the application is started by the agent which naturally grants it control over the underlying OS process. The data exchange of a client and its agent (note: a run of a client can have only one agent) occurs via two different media. The first media is a unix socket connection which is used to exchange metadata and commands. Metadata is for example the version of the socket, [tags](#tags) of the run and especially information on how metrics should be exchanged. The second media is used by the client to store current metrics and by the agent to fetch this data. This can be a posix SharedMemory segment ([shm](http://pubs.opengroup.org/onlinepubs/007908799/xsh/shm_open.html)), a memory mapped file ([mmap](http://man7.org/linux/man-pages/man2/mmap.2.html)) or (currently not implemented) for example another socket connection or even the same unix socket for issuing commands. Shm and mmap have the big advantage that they are tremendously fast for writing and reading but impose the constraint on the agent that it has to occasionally read and copy that data. Therefore metric data can be lost. For instance, a short spike in a metric can be overseen. The agent aggregates bunches of metric and other meta data like tags and prints them to STDOUT or periodically sends them to a server.

The Tirion server has two big tasks. One task is receiving and saving data of runs from many agents. The other is sending this data to clients who want to analyse and display it. For portability reasons and easier integration the server uses HTTP as its protocol with JSON for marshaling complex data structures. The configurable backend of the server is used to save run data permanently for instance into a database.

## How to build Tirion?

Tirion provides precompiled 32 and 64 bit Linux binaries. Other platforms are currently not supported, but might work. The client and the server are not OS specific. The agent on the other hand uses the [proc filesystem](https://en.wikipedia.org/wiki/Procfs) which is only available on unix-like systems.

If you do not want to use the precompiled binaries, it depends on what part of Tirion you want to use. If you just want to include the client library into your application take a look at the [clients section](#client-libraries). If you want to run the agent and the server you have to install and configure Go first, as Tirion is mostly written in Go. Your distribution will most definitely have some packages or you can be brave and just install it yourself. Have a look at [the official documentation](http://golang.org/doc/install). Good luck!

After installing Go you can download Tirion’s dependencies by issuing the following commands in a fresh terminal:

```bash
go get github.com/robfig/revel
```

After that you can fetch and install Tirion with the following commands:

```bash
go get github.com/zimmski/tirion
cd $GOPATH/src/github.com/zimmski/tirion
make
```

This will fetch the whole code of the Tirion infrastructure but will only compile the tirion-agent to <code>$GOBIN/tirion-agent</code>. As for the tirion-server you can deploy it by following the [revel documentation](http://robfig.github.io/revel/manual/deployment.html), which is the web framework that is used by the tirion-server, or you can have a look at the [README of the tirion-server](/tirion-server/README.md) for starting the server without deploying.

## How do I set up a Tirion infrastructure?

If you do not use the precompiled binaries you have to [compile Tirion](#how-to-build-tirion) before you can set up an infrastructure. There are two components that need configuration. The tirion-server needs a server configuration and a working backend. Please have a look at the [README of the tirion-server](/tirion-server/README.md) on how to accomplish that. The client (your application) must have a fitting [metric-file](#metric-file) which is fed to the agent. That is all you need to set up a complete Tirion infrastructure!

## How do I use Tirion?

### Client libraries

To send metrics from your application to the tirion-agent, the corresponding client library for your programming language must be included and used in your application. In addition, the tirion-agent needs to know which metrics you want to send to the server via a [metric file](#metric-file).

The following programming languages currently have a working client library. Please have a look at the respective README on how to include and use a library.

* [C-client](/clients/c-client/README.md)
* [Go-client](/clients/go-client/README.md)

If you do not see your favourite language here and are eager to try out Tirion with your application, just submit an [issue via project the tracker](https://github.com/zimmski/tirion/issues/new) and I will see what I can do.

### External metrics

External metrics of the client are recorded by the tirion-agent and consist mostly of data fetched via the proc filesystem. The following groups define each metric with a name and type which can be used in a metric file or other metric structure definition of Tirion. Please note that you do not need to use all metrics of a group, any at all or even any external metric for a correct metric file.

#### Currently supported external metrics

* proc.io (see the [proc man page](http://man7.org/linux/man-pages/man5/proc.5.html) header <code>/proc/[pid]/io</code> for a description of each metric)
	** proc.io.cancelled_write_bytes int
	** proc.io.rchar int
	** proc.io.read_bytes int
	** proc.io.syscr int
	** proc.io.syscw int
	** proc.io.wchar int
	** proc.io.write_bytes int
* proc.stat (see the [proc man page](http://man7.org/linux/man-pages/man5/proc.5.html) header <code>/proc/[pid]/stat</code> for a description of each metric)
	** proc.stat.blocked int
	** proc.stat.cguest_time int
	** proc.stat.cmajflt int
	** proc.stat.cminflt int
	** proc.stat.cnswap int
	** proc.stat.cstime int
	** proc.stat.cutime int
	** proc.stat.delayacct_blkio_ticks int
	** proc.stat.endcode int
	** proc.stat.exit_signal int
	** proc.stat.flags int
	** proc.stat.guest_time int
	** proc.stat.itrealvalue int
	** proc.stat.kstkeip int
	** proc.stat.kstkesp int
	** proc.stat.majflt int
	** proc.stat.minflt int
	** proc.stat.nice int
	** proc.stat.nswap int
	** proc.stat.num_threads int
	** proc.stat.pgrp int
	** proc.stat.pid int
	** proc.stat.policy int
	** proc.stat.ppid int
	** proc.stat.priority int
	** proc.stat.processor int
	** proc.stat.rss int
	** proc.stat.rsslim int
	** proc.stat.rt_priority int
	** proc.stat.session int
	** proc.stat.sigcatch int
	** proc.stat.sigignore int
	** proc.stat.signal int
	** proc.stat.startcode int
	** proc.stat.startstack int
	** proc.stat.starttime int
	** proc.stat.state int
	** proc.stat.stime int
	** proc.stat.tpgid int
	** proc.stat.tty_nr int
	** proc.stat.utime int
	** proc.stat.vsize int
	** proc.stat.wchan int
* proc.statm (see the [proc man page](http://man7.org/linux/man-pages/man5/proc.5.html) header <code>/proc/[pid]/statm</code> for a description of each metric)
	** proc.statm.data int
	** proc.statm.dt int
	** proc.statm.lib int
	** proc.statm.resident int
	** proc.statm.share int
	** proc.statm.size int
	** proc.statm.text int

#### Important external metrics

* proc.stat.num_threads - how many threads are currently used
* proc.stat.utime - the user space time of the process
* proc.statm.data - the amount of data pages of the process
* proc.statm.resident - the amount of resident pages of the process

### Internal metrics

Data of internal metrics are provided by the client itself. The client libraries provide different functions to change data of the metrics which then can be read by the corresponding agent of the application. Note that there is no guarantee that the agent fetches all changes as there is no message queue. Instead the agent retrieves all metrics periodically. An internal metric consists of a name and a type which are required attributes of a Tirion metric in general.

A internal metric name has the following restrictions:

* It must be unique
* It must not be empty
* It must only consist of alphanumeric characters, “.”, “-” and “_”
* It can have at most 256 characters

The following internal metric types are currently supported:

* float
* int

### Metric file

A metric file is just a simple text file with a JSON structure which is fed to the tirion-agent that monitors the given application. The JSON structure consists of an array of [external](#external-metrics) and [internal](#internal-metrics) metrics. Only internal metrics have to follow a specific order which must suit the given client. External metrics can be defined in any order. There is a limit of 2^32 metrics per metrics file. Each metric must have a unique name and a type. This also means that an external metric can only be used once in a metric file. Please have a look at currently available [external metrics](#external-metrics) and the definition of [internal metrics](#internal-metrics).

For example:

```json
[
	{
		"name" : "proc.stat.utime",
		"type" : "int"
	},
	{
		"name" : "entry.count",
		"type" : "int"
	},
	{
		"name" : "data.size",
		"type" : "int"
	},
	{
		"name" : "entries.avg",
		"type" : "float"
	},
	{
		"name" : "proc.statm.data",
		"type" : "int"
	},
	{
		"name" : "proc.statm.resident",
		"type" : "int"
	}
]
```

Defines the three external metrics <code>proc.stat.utime</code>, <code>proc.statm.data</code> and <code>proc.statm.resident</code> and three internal metrics <code>entry.count</code>, <code>data.size</code> and <code>entries.avg</code>. As you can see, each metric has its own name and type definition. The internal metrics order has a special meaning as it also stands for the index which can be used from the client. In this example <code>entry.count</code> has the index 0, <code>data.size</code> the index 1 and <code>entries.avg</code> the index 2. Because of this meaning, it does make sense to add new metrics at the bottom of the JSON array in order not to mix up existing indices.

### Tags

Tags are markers in the timeline of client execution and can be issued by the client itself. In comparison to internal metrics, tags can never get lost. A tag’s only attribute is the message, which has the restrictions of at most 512 characters and it can not consist of newlines. Clients will replace newlines with spaces to make the handling of tags more user-friendly.

### tirion-agent

Please have a look at the [README of the tirion-agent](/tirion-agent/README.md).

### tirion-server

Please have a look at the [README of the tirion-server](/tirion-server/README.md).

### UI

Heavy WIP (TODO document this)

## Can I make feature requests, report bugs and problems?

Sure, just submit an [issue via the project tracker](https://github.com/zimmski/tirion/issues/new) and I will see what I can do. Please note that I do not guarantee to implement anything as Tirion is purely a leisure project. Also bugs and problems are more important to me than new features.
