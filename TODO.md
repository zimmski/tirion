## General

* It is not a fatal if a connection fails or got closed. So fail gracefully. (especially agent and client libraries)
* Use the openSUSE build service to build packages for openSUSE, ubuntu and fedora. 32 and 64 bit
	* tirion-agent
	* tirion-server
	* client-libraries separately

## Agent

* Move /shm to /collector and refactor
* Use the time of the agent for metrics and tags NOT the time of the server. This makes the metric timestamps more exact because of HTTP and server delays.
* A program is monitored as long as the process of the program is alive and not as long as the socket.
* Make memory reports more accurate (especially for multi process programs)
	* Convert all memory metrics to KiloByte
	* [ps_mem](https://raw.github.com/pixelb/ps_mem) does some interesting things
	* Have a look at /proc/<pid>/smaps "Private_Dirty"
	* Consider shared memory/PSS
	* Consider swapped out memory
	* Consider the language/VM of the program for example connect to the Java VM
	* Find a linux/unix pendant to "vmmap" and have a look on how it works
* Sockets can reconnect
* A program which is not started by the agent should be able to connect to the agent via a socket.
* Add CPU limit (if the agent starts the program)
* Oversee client process with Linux containers [LXC](https://wiki.deimos.fr/LXC_:_Install_and_configure_the_Linux_Containers#Memory)

## Client libraries

* Reduce add, sub, inc and dec to add. Can be seen in the Java and Python library.
* Use a lock with all metrics related like in the Java and Python library
* Use a queue for receive like in Java library. Otherwise more than one command during one receive would be lost.
* Add "get" and "set" functions for metrics
* Init functions must handle open connections of all their members -> only start disconnected members
* Sockets can reconnect
* Shared memory and MMap can reconnect

## Server

* Program name cannot have URL forbidden characters. Put this in a function, use it in server, agent and client libraries and document it.
* JSON export of all informations about a run without metrics and tags
* Add limit parameters for metric and tag exports concerning their timestamps
	* "from"
	* "to"
* Add limit parameters for metrics concerning selection of metrics for example just m1 and m2 instead of all
* Export metrics via
	* CSV
	* JSON
* service files
	* initd
	* systemd
	* upstart
* "COPY run1 FROM STDIN ..." to insert metrics and tags for PostgreSQL backend
* SQLite backend
* MySQL backend

## UI

* Rewrite UI with AngularJS
* Show Min, Max, Average, Mean, ... in graphs
* Find an alternative for Highstock graphs maybe [nvd3](https://github.com/novus/nvd3)
* Live graphs if the run is still ongoing
	* Switches to normal view if run has finished
* Compare more than one metric of the same run (in one or more graphs)
* Compare metrics of different runs (in one or more graphs)
* Tabs to switch between metric comparisons

## Bigger things

* Competition UI and server component for running automatic solver runs with different files to compare their runtimes.
