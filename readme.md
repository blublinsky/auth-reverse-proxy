# Authorization using reverse proxy

Usage of the reverse proxy makes authorization quite simple. This simple code shows implementation of several
authorization methods. It is inspired by the following things:

* [Basic reverse proxy](https://github.com/angelbarrera92/basic-auth-reverse-proxy)
* [This Gist](https://gist.github.com/dbrinegar/88c0acf0bc4b0f0fc0c3b2bdbb0a62d3)

## Token based authorization

This is a very simple authorisation based on the string token, shared between proxy and client. The token is submitted
as a header `auth-token`. The implementation of handler is [here](auth/token_auth.go). The [main](cmd/proxy/proxy.go)
defines proxy parameters, and starts the proxy server using handler.

To test it execute:

```shell
curl http://localhost:9090/foo
```

returns `Unauthorised`

If you use a proper token:

```shell
curl http://localhost:9090/hello -H "auth-token: 12345"
```

the request will be forwarded  to the actial [server](cmd/server/server.go)

## Basic authorization

This is a very simple authorisation based on the user name/password, shared between proxy and client. The credentials are submitted using a standard basic auth header `Authorization`. The implementation of handler is[here](auth/basic_auth.go). The [main](cmd/proxy/proxy.go) defines proxy parameters, and starts the proxy server using handler.

To test it execute:

```shell
curl http://localhost:9090/foo
```

returns `Unauthorised`

If you use a proper token:

```shell
curl http://boris:12345@localhost:9090/
```

or using curl user:

```shell
curl -u "boris:12345" http://localhost:9090/
```

or using bas64 encoded credentials:

```shell
http://localhost:9090/ -H "Authorization: Basic Ym9yaXM6MTIzNDU="
```

the request will be forwarded  to the actial [server](cmd/server/server.go)

## JWT based authentication

Implementation is based on [this blog post](https://www.sohamkamani.com/golang/jwt-authentication/). Some other options are rescibed [here](https://medium.com/geekculture/securing-apis-via-jwt-in-golang-9d3659a32c34). The authorisation is based on the JWT token. It provides a signon method allowing user to get a JWT token based on token based authentication (see above). Once JWT token is obtained by client it can use use it by submitting it as a header `jwt-token`. The implementation of handler is [here](auth/jwt_auth.go). The [main](cmd/proxy/proxy.go) defines proxy parameters, and starts the proxy server using handler.

To test it first execute:

```shell
curl http://localhost:9090/signon -H "auth-token: 12345"
```

which returns jwt token, that can be used in the subsequent request

```shell
curl http://localhost:9090/hello -H "jwt-token: <token>"
```

the request will be forwarded  to the actial [server](cmd/server/server.go)

## Using HTTPS

The same proxy can be used for TLS-terminating of the client's request. For more details reffer to this [blog post](https://eli.thegreenplace.net/2022/go-and-proxy-servers-part-2-https-proxies/) and this [gist](https://gist.github.com/matishsiao/8270e18923d8f78f56c2). Note that in this case `ListenAndServeTLS` is used which requires certificate

## Supporting GRPC

GRPC support is based on the [GRPC proxy](https://github.com/mwitkow/grpc-proxy). Currently only Token based authorization is implemented, but additional authentication mechanisms, described above can be implemented in a similar fashion, see, for example, [here](https://dev.to/techschoolguru/use-grpc-interceptor-for-authorization-with-jwt-1c5h)

For local testing of handler you can use [handler_test](grpcproxy/handler_test.go) and for proxy - [proxy_test](grpcproxy/proxy_test.go). To test the whole thing, start [server](cmd/grpc_server/server.go), and then [proxy](cmd/grpc_proxy/proxy.go). With them running, you can use a simple [client](cmd/grpc_client/client.go) to exercise everything.
