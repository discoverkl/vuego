# Vuego

Call remote Go methods directly from Javascript through WebSocket.

See `examples` folder.

# Helloword

```shell
# cd ANY_GO_MODULE_DIR
go run github.com/discoverkl/vuego/examples/helloworld
```

**OR**

```shell
git clone https://github.com/discoverkl/vuego.git
cd vuego/examples/helloworld
go run .
```

# Build Examples

Run `go generate` if you modified any files under `fe/dist`. This will pack the directory into Go source code. (pkged.go)
