## Redis Adapter
this package is an adapter for redis go client.
it provides an interface to create a singleton,
thread-safe connection to a redis server. It does not
supports clustering. It only support separate read &
and write. To have different connections, you must
connect via different host and ports. For a single host:port
combination, only one global long-lived connection is created. No matter
how many times `NewRedis()` is called, it returns the same connection.

##### Usage
To simply connect to a redis server, do the following:

```go
redis, err := NewRedis(&RedisConfig{ Host: "127.0.0.1", Port: 6379}, nil)
if err != nil {
	// show some error
	os.Exit(1)
}
res := redis.Set("sample-key", "sample value", 25)
// to get a value
val, err := redis.Get("sample-key")
if err != nil {
	fmt.Println("failed to get key 'sample-key'")
}
```

If you want to have a separate read & write instances, then simply pass
a RedisConfig to the second param of NewRedis(), it manages all DEL and SET
to the write server internally:
```go
redis, err := NewRedis(&RedisConfig{ Host: "127.0.0.1", Port: 6379}, 
	&RedisConfig{ Host: "127.0.0.1", Port: 6380})
if err != nil {
	// show some error
	os.Exit(1)
}
res := redis.Set("sample-key", "sample value", 25) // sets on write server
// to get a value
val, err := redis.Get("sample-key") // gets from read server
if err != nil {
	fmt.Println("failed to get key 'sample-key'")
}
```

For deleting a key, use `Del()` command:
```go
var res = redis.Del("key1", "key2")
if !res {
	fmt.Println("failed to delete all keys")
}
```