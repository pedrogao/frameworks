# go frameworks

> implement some service frameworks for go.

## modules

-[log](./log) simple & easy-use logger library  
-[orm](./orm) tiny orm framework for go  
-[rpc](./rpc) tiny rpc framework for go  
-[web](./web) simple & easy-use web framework for go  

## example

see [hello](./hello/main.go), all implements by modules.

### setup & run

```sh
$ go run hello/main.go

2022-06-25T15:28:53+08:00 INFO raw.go:39 exec: DROP TABLE IF EXISTS User;  []
2022-06-25T15:28:53+08:00 INFO raw.go:39 exec: CREATE TABLE User(Name text);  []
2022-06-25T15:28:53+08:00 INFO router.go:18 Route  GET - /
2022-06-25T15:28:53+08:00 INFO router.go:18 Route POST - /users
2022-06-25T15:28:53+08:00 INFO router.go:18 Route  GET - /users
2022-06-25T15:28:53+08:00 INFO engine.go:32 start server at: :3000
```

### create & query user by curl

```sh
$ curl localhost:3000
simple & easy%

$ curl -X POST "http://localhost:3000/users?name=pedro"
{"message":"ok"}

$ curl -X GET "http://localhost:3000/users?name=pedro"
{"name":"pedro"}

...
2022-06-25T15:29:01+08:00 INFO router.go:26 handle route, path: /, method: GET, status: 200
2022-06-25T15:30:28+08:00 INFO router.go:26 handle route, path: /, method: POST, status: 404
2022-06-25T15:30:38+08:00 INFO raw.go:39 exec: INSERT INTO User(`Name`) values (?)  [pedro]
2022-06-25T15:30:38+08:00 INFO router.go:26 handle route, path: /users, method: POST, status: 200
2022-06-25T15:30:46+08:00 INFO raw.go:49 query row: SELECT * FROM User WHERE Name = ?  [pedro]
2022-06-25T15:30:46+08:00 INFO router.go:26 handle route, path: /users, method: GET, status: 200
```

## references

- [Getting started with multi-module workspaces](https://go.dev/doc/tutorial/workspaces)
- [7 days golang programs from scratch](https://github.com/geektutu/7days-golang)