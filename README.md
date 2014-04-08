Akadok is the game server for [Deimos](http://deimos-ga.me).

# Building

The best way to build Akadok for your system is to download the latest version of the [Go](http://golang.org) compiler. Then you just have to run the following commands in a terminal:

    go get bitbucket.org/deimosgame/go-akadok
    cd $GOPATH/bitbucket.org/deimosgame/go-akadok
    go install

Akadok can now be ran using the `go-akadok` command. It will generate its default config file in its directory. You can edit it as you want.

# Testing

Akadok is using multiple sub-packages to make its components modular. However, the standard `go test` command only tests the current package. To run all the test suites of Akadok, this command should be ran in Akadok's root directory:

    go test ./...

# Configuration

In the configuration file (*server.cfg* by default, but this can be customized further by running the command `go-akadok /path/to/config/file.cfg`), the directives follow the format **param = value**. For instance, this is a valid config file that might be used with Akadok:

    ; My wonderful config file!
    name = The best server in the world
    pport = 1337
    maxplayers = 42

Here are a list of parameters that may be used in Akadok config files (unknown parameters are ignored by the software):

**name**: Changes the server name as it appears in the in-game server list. Default: Akadok server

**ip**: Use this directive to force a binding IP. By default, Akadok will try to resolve server's external IP through master server ; if that fails, it will bind to 127.0.0.1.

**port**: Akadok's port. Default: 1518

**max_players**: Maximum online players at the same time. Default: 15

**maps**: Maps used for map rotation. Map names are separated by commas. Default: map1, map2, map3

**verbose**: Used for debugging purposes. Outputs every event on the server to logs. Default: false

**log_file**: Changes server's logs location. Default: server.log

**register_server**: Determines wether or not server will try to contact master server in order to be registered on public server list. Default: false