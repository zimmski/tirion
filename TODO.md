# T H E   T O D O   O F   T I R I O N,   P A R T   I

## Agent

* Use the time of the agent for metrics and tags NOT the time of the server. This makes the metric timestamps more exact because of HTTP and server delays.
* A program is monitored as long as the process of the program is alive and not as long as the socket.
* Sockets can reconnect
* A program which is not started by the agent should be able to connect to the agent via a socket.
* It is not a fatal if a connection fails or got closed. So fail gracefully.
* Add CPU limit (if the agent starts the program)
* Oversee client process with Linux containers [LXC](https://wiki.deimos.fr/LXC_:_Install_and_configure_the_Linux_Containers#Memory)

## Client libraries

* It is not a fatal if a connection fails or got closed. So fail gracefully.
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
