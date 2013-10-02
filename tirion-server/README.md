# tirion-server

The tirion-server is a HTTP server using the [revel framework](http://robfig.github.io/revel/) who receives data from agents and serves clients for displaying and analysing this data. If you use the precompiled Tirion binaries you can use the server nearly right away otherwise you have to [build Tirion](/README.md#how-to-build-tirion) first or [start the server without deploying](#non-deploy-start) it. This README focuses on how to configure and use the tirion-server and not how it works, have a look at [“How does Tirion work?”](/README.md#how-does-tirion-work) if you want to know more.

## Configure tirion-server

The tirion-server requires a working backend to save run data. Currently only a PostgreSQL backend is implemented but others can be easily added by implementing the [Backend interface](/backend/backend.go). For example adding a MySQL backend would only require copying the PostgreSQL, changing the Go driver and adapting the SQL statements.The PostgreSQL backend requires a running PostgreSQL server, a user and a database.

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
* db.driver - Is the key of the backend you want to use e.g. “postgresql”.
* db.spec - Is the connection string for the backend.

Other parameters like the HTTP port and timeouts are revel specific and are specified in the [revel documentation](http://robfig.github.io/revel/manual/appconf.html).

## Run the tirion-server

As the tirion-server is a revel application the [revel documentation](http://robfig.github.io/revel/manual/deployment.html) shows all configurations of deploying it. If you do not use the precompiled binaries you can also just run the server with the source code in development mode

```bash
revel run github.com/zimmski/tirion/tirion-server dev
```

or production mode.

```bash
revel run github.com/zimmski/tirion/tirion-server prod
```

## Routes of the tirion-server

Heavy WIP (TODO document this)
