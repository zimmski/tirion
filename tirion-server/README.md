# tirion-server

The tirion-server is a HTTP server using the [revel framework](http://robfig.github.io/revel/) who receives data from agents and serves clients for displaying and analysing this data. If you use the precompiled Tirion binaries you can use the server nearly right away otherwise you have to [build Tirion](/#how-to-build-tirion) first or [start the server without deploying](#run-the-tirion-server) it. This README focuses on how to configure and use the tirion-server and not how it works, have a look at ["How does Tirion work?"](/#how-does-tirion-work) if you want to know more.

## Configure tirion-server

The tirion-server requires a working backend to save run data. Currently only a PostgreSQL backend is implemented but others can be easily added by implementing the [Backend](/backend/backend.go) interface. For example adding a MySQL backend would only require copying the PostgreSQL, changing the Go driver and adapting the SQL statements.The PostgreSQL backend requires a running PostgreSQL server, a user and a database.

To initialize the database, run the following command and make sure that all statements executed without errors.

```bash
psql <database> <user> < <tirion-server path>/scripts/postgresql_ddl.sql
```

After initializing the backend you have to create the server configuration. With the following command you can add a template for your configuration.

```bash
cp <tirion-server path>/conf/app.conf.sample <tirion-server path>/conf/app.conf
```

Now you can edit <tirion-server path>/conf/app.conf as you need. There are some important parameters you should change:

* app.secret - Is the key for cryptographic functions and signing so make sure that no one gets hold of this key!
* db.driver - Is the key of the backend you want to use e.g. "postgresql".
* db.maxIdleConns - Maximum idle db connections.
* db.maxOpenConns - Maximum open db connections.
* db.spec - Is the [connection string](#connection-string-of-backends) of the backend.

Other parameters like the HTTP port and timeouts are revel specific and are specified in the [revel documentation](http://robfig.github.io/revel/manual/appconf.html).

# Connection string of backends

Connection strings of all backends are lists of name-value parameter pairs separated by one space. Name and value are separated by one equal sign "=". Values can be quoted with either a double " or single ' quotation mark.

For example this is a valid connection string for the PostgreSQL backend: <code>user=zimmski dbname=tirion sslmode=disable</code>

The following subsections will describe their distinct parameters.

## PostgreSQL

The PostgreSQL backend uses the Go package [<code>github.com/lib/pq</code>](https://github.com/lib/pq) therefore all valid parameters can also be found on the project’s page.

- <code>user</code> The database user.
- <code>password</code> The user’s password.
- <code>dbname</code> The name of the database.
- <code>host</code> The database server host to connect to. Values that start with "/" are for unix domain sockets. (default "localhost")
- <code>port</code> The database server port. (default "5432")
- <code>sslmode</code> Can be one of the following values.
    - <code>disable</code> No SSL is used.
    - <code>require</code> Use SSL and skip the verification.
    - <code>verify-full</code> Use SSL and require verification. (default)

## Run the tirion-server

As the tirion-server is a revel application the [revel documentation](http://robfig.github.io/revel/manual/deployment.html) shows all configurations of deploying it. If you do not use the precompiled binaries you can also just run the server with the source code in development mode

```bash
revel run github.com/zimmski/tirion/tirion-server dev
```

or production mode.

```bash
revel run github.com/zimmski/tirion/tirion-server prod
```

## Routes of the tirion-server (server API)

The tirion-server provides the following HTTP routes.

- GET <code>/</code>

	Shows all available programs of all runs.

	- URI parameters

		<code>none</code>

	- Request parameters

		<code>none</code>

	- Output <code>HTML</code>

- GET <code>/program/:programName</code>

	Shows all finished and still ongoing runs of a program.

	- URI parameters

		<code>:programName</code> URI-cleaned program name

	- Request parameters

		<code>none</code>

	- Output <code>HTML</code>

	- Errors

		- <code>404</code> if there is no program with the given program name

- POST <code>/program/:programName/run/start</code>

	Start a new run.

	- URI parameters

		- <code>:programName</code> URI-cleaned program name

	- Request parameters

		- <code>name</code> original program name (string)
		- <code>sub_name</code> (optional) subname of the program or run (string)
		- <code>interval</code> interval of this run for metric fetching (int64)
		- <code>metrics</code> metrics of this run ([metric file](/#metric-file))
		- <code>prog</code> program command (string)
		- <code>prog_arguments</code> (optional) program command arguments (string)

	- Output <code>JSON</code>

		```json
		{
			"Run": "int32 # the ID of the started run",
			"Error": "string # the error string if an error occured"
		}
		```

	- Errors

		- <code>404</code> if there is no program with the given program name
		- <code>non-empty Error field</code> on various errors concerning the validation of the run parameters

- GET <code>/program/:programName/run/:runID</code>

	Shows all information and all metrics (as graphs) of a run.

	- URI parameters

		- <code>:programName</code> URI-cleaned program name
		- <code>:runID</code> ID of the run

	- Request parameters

		<code>none</code>

	- Output <code>HTML</code>

		- <code>404</code> if there is no program with the given program name
		- <code>404</code> if there is no run with the given ID

- GET <code>/program/:programName/run/:runID/metric/:metricName</code>

	Returns all data of single metric of a given run.

	- URI parameters

		- <code>:programName</code> URI-cleaned program name
		- <code>:runID</code> ID of the run

	- Request parameters

		<code>none</code>

	- Output <code>JSON</code>

		```json
		[
			[ <timestamp>, <value> ]
			...
		]
		```

	- Errors

		- <code>404</code> if there is no program with the given program name
		- <code>404</code> if there is no run with the given ID

- POST <code>/program/:programName/run/:runID/insert</code>

	Inserts rows of metric data for a given ongoing run.

	- URI parameters

		- <code>:programName</code> URI-cleaned program name
		- <code>:runID</code> ID of the run

	- Request parameters

		- <code>metrics</code> rows of metric data

			```json
			[
				{
					"Time": "timestamp # time of the metric row",
					"Data": [
						<value>
						...
					]
				}
				...
			]
			```

	- Output <code>JSON</code>

		```json
		{
			"Error": "string # the error string if an error occured"
		}
		```

	- Errors

		- <code>404</code> if there is no program with the given program name
		- <code>404</code> if there is no run with the given ID
		- <code>non-empty Error field</code> on various errors concerning the validation of the metric values

- GET <code>/program/:programName/run/:runID/stop</code>

	Stops an ongoing run.

	- URI parameters

		- <code>:programName</code> URI-cleaned program name
		- <code>:runID</code> ID of the run

	- Request parameters

		<code>none</code>

	- Output <code>JSON</code>

		```json
		{
			"Error": "string # the error string if an error occured"
		}
		```

	- Errors

		- <code>404</code> if there is no program with the given program name
		- <code>404</code> if there is no run with the given ID
		- <code>non-empty Error field</code> on various errors concerning stopping the program

- POST <code>/program/:programName/run/:runID/tag</code>

	Inserts a tag for a given ongoing run.

	- URI parameters

		- <code>:programName</code> URI-cleaned program name
		- <code>:runID</code> ID of the run

	- Request parameters

		- <code>tag</code> the tag string (string)
		- <code>time</code> the time of the tag (timestamp)

	- Output <code>JSON</code>

		```json
		{
			"Error": "string # the error string if an error occured"
		}
		```

	- Errors

		- <code>404</code> if there is no program with the given program name
		- <code>404</code> if there is no run with the given ID
		- <code>non-empty Error field</code> on various errors concerning the validation of the tag  values

- GET <code>/program/:programName/run/:runID/tags</code>

	Returns all tags of a given run.

	- URI parameters

		- <code>:programName</code> URI-cleaned program name
		- <code>:runID</code> ID of the run

	- Request parameters

		<code>none</code>

	- Output <code>JSON</code>

			```json
			[
				{
					"x": "timestamp # time of the tag",
					"title": "string # tag string"
				}
				...
			]
			```

	- Errors

		- <code>404</code> if there is no program with the given program name
		- <code>404</code> if there is no run with the given ID

