# Build and install Tirion on openSUSE 12.3 (64bit)

This short guide will guide you step-by-step through the build and install process to run a complete Tirion infrastructure on openSUSE 12.3 64bit. This includes a Tirion client, agent and server as well as the required backend for the server. This guide can work with older (and newer) versions of openSUSE but it is not guaranteed to do so.

#### 1. Install and configure backend

The Tirion server requires a working backend to save data of its clients. We will use PostgreSQL. If you already have a working PostgreSQL server make sure that you have a working user and database for the server configuration. The default PostgreSQL installation of openSUSE requires a ident server too.

```bash
sudo zypper install pidentd
sudo chkconfig --add identd
sudo service identd start
sudo zypper install postgresql92 postgresql92-server
sudo chkconfig --add postgresql
sudo service postgresql start
sudo sudo -u postgres createuser --superuser $USER
createdb --owner $USER $USER
```

This will install the ident and PostgreSQL server and also adds a user with no password as well as a database for the user. To test if everything went well issue the following command.

```bash
psql -c "select now()"
```

This should print out the current time of your system.

#### 2. Install  and configure Go

All major components of Tirion are written in Go therefore an up to date version of Go is needed. As described in the [main README](/#how-to-build-tirion) you can follow the [official Go documentation](http://golang.org/doc/install) to install Go or use the distribution's packages.

OpenSUSE provides a repository with up to date packages.

```bash
sudo zypper ar -f “http://download.opensuse.org/repositories/devel:/languages:/go/openSUSE_12.3/” "devel language go"
sudo zypper install go
```

For easier use  add the go binary path to the bash path.

```bash
echo "export PATH=$GOBIN:$PATH" >> ~/.bashrc
```

After installing Go reopen your terminal to load the new bash configurations.

#### 3. Installing Tirion requirements

Some packages are needed to successfully build Tirion and its requirements.
(OpenSUSE 12.3 issues a problem at this step. Use the suggested solution of removing the patterns-openSUSE-minimal_base-conflicts pattern.)

```bash
sudo zypper install clang gcc git make mercurial
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

### 5. Configure Tirion

The Tirion server needs a configuration file and a initialized backend. Both can be done by issuing the following commands.

```bash
cp $GOPATH/src/github.com/zimmski/tirion/tirion-server/conf/app.conf.sample $GOPATH/src/github.com/zimmski/tirion/tirion-server/conf/app.conf
sed -i "s/user=zimmski dbname=tirion/user=$USER dbname=$USER'/" $GOPATH/src/github.com/zimmski/tirion/tirion-server/conf/app.conf
psql < $GOPATH/src/github.com/zimmski/tirion/tirion-server/scripts/postgresql_ddl.sql
```

### 6. Start Tirion

Please have a look at the [main README](/) for a more complete look at how you can run Tirion. As an example the following commands start the Tirion server in development mode and run the Go example client through the Tirion agent.

```bash
revel run github.com/zimmski/tirion/tirion-server dev
```

This start the HTTP server on port 9000. Now open up a new terminal for the following command.

```bash
tirion-agent -verbose -interval 100 -metrics-filename $GOPATH/src/github.com/zimmski/tirion/clients/example-metrics.json -exec go-client -exec-arguments "-verbose -runtime 2" -socket /tmp/tirion.sock -server "localhost:9000"
```

After the agent has finished you can look at the results via the HTTP server [http://localhost:9000](http://localhost:9000).
