# Build and install Tirion on Ubuntu 13.04 (64bit)

This short guide will guide you step-by-step through the build and install process to run a complete Tirion infrastructure on Ubuntu 13.04 64bit. This includes a Tirion client, agent and server as well as the required backend for the server. This guide can work with older (and newer) versions of Ubuntu but it is not guaranteed to do so.

#### 1. Install and configure backend

The Tirion server requires a working backend to save data of its clients. We will use PostgreSQL. If you already have a working PostgreSQL server make sure that you have a working user and database for the server configuration.

```bash
sudo apt-get install postgresql
sudo -u postgres createuser --superuser -W $USER
createdb --owner $USER $USER
```

This will install the PostgreSQL server and add a user with a password as well as a database for the user. To test if everything went well issue the following command.

```bash
psql -c "select now()"
```

This should print out the current time of your system.

#### 2. Install  and configure Go

All major components of Tirion are written in Go therefore an up to date version of Go is needed. As described in the [main README](/#how-to-build-tirion) you can follow the [official Go documentation](http://golang.org/doc/install) to install Go or use the distribution's packages. Unfortunately the official Go packages for Ubuntu 13.03 are more than a year old.

[This](http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go) blog provides a script for easily installing Go with Ubuntu.

```bash
wget https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz
tar xvfz godeb-amd64.tar.gz
./godeb install 1.1.2
```

For stability make sure that you use a stable version. 1.1.2 for example is known to work with Tirion. After installing Go a basic configuration is needed.

```bash
echo "export GOPATH=~/gocode" >> ~/.bashrc
echo "export PATH=~/gocode/bin:$PATH" >> ~/.bashrc
```

This will add Go’s needed environment variables to your bash configuration. Make sure that you load this new configuration or just reopen your terminal.

#### 3. Installing Tirion requirements

Some packages are needed to successfully build Tirion and its requirements.

```bash
sudo apt-get install clang gcc git make mercurial
```

Tirion needs the Revel web framework as well as the backend driver. You can install both by issuing the following commands.

```bash
go get github.com/lib/pq
go get github.com/robfig/revel
go install github.com/robfig/revel/revel
```

#### 4. Build Tirion

To fetch, install and compile Tirion just issue the following commands.

```bash
go get github.com/zimmski/tirion
cd $GOPATH/src/github.com/zimmski/tirion
make tirion-agent
make go-client
```

This will compile the Tirion agent and the Go example client.

#### 5. Configure Tirion

The Tirion server needs a configuration file and a initialized backend. Both can be done by issuing the following commands. Note: You should adapt the password of the second command to your own password for the PostgreSQL user!

```bash
cp $GOPATH/src/github.com/zimmski/tirion/tirion-server/conf/app.conf.sample $GOPATH/src/github.com/zimmski/tirion/tirion-server/conf/app.conf
sed -i "s/user=zimmski dbname=tirion/user=$USER dbname=$USER password='YOUR PASSWORD'/" $GOPATH/src/github.com/zimmski/tirion/tirion-server/conf/app.conf
psql < $GOPATH/src/github.com/zimmski/tirion/tirion-server/scripts/postgresql_ddl.sql
```

#### 6. Start Tirion

Please have a look at the [main README](/) for a more complete look at how you can run Tirion. As an example the following commands start the Tirion server in development mode and run the Go example client through the Tirion agent.

```bash
revel run github.com/zimmski/tirion/tirion-server dev
```

This start the HTTP server on port 9000. Now open up a new terminal for the following command.

```bash
tirion-agent -verbose -interval 100 -metrics-filename $GOPATH/src/github.com/zimmski/tirion/clients/example-metrics.json -exec go-client -exec-arguments "-verbose -runtime 2" -socket /tmp/tirion.sock -server "localhost:9000"
```

After the agent has finished you can look at the results via the HTTP server [http://localhost:9000](http://localhost:9000).
