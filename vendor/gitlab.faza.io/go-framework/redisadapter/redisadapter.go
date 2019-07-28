package redisadapter

// this package is an adapter for redis go client.
// it provides an interface to create a singleton,
// thread-safe connection for a redis server. It does not
// supports clustering. It only supports separate read &
// and write. To have different connections, you must
// connect via to different host and ports. For a single host:port
// combination, only one global long-lived connection is created.

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/matryer/resync"
)

const (
	redisPolicySeparateInstances   = 101
	redisPolicyUseSameForReadWrite = 102
	redisTypeRead                  = 201
	redisTypeWrite                 = 202
)

var redisUriPattern = "%v:%v"

// this is a redis manager which manages the connections pool
// as well as read and write instances policy.
type Redis struct {
	// stores the current policy; two possible values
	// for this field;
	// redisPolicyUseSameForReadWrite which is used when the same instance of Redis
	// is used for read/write.
	// redisPolicySeparateInstances when two different instances are used for
	// read and write.
	policy int

	// The result of last SET call is stored in this variable
	LastSet *redis.StatusCmd

	// The result of last DEL call
	LastDel *redis.IntCmd

	ReadConfig  *RedisConfig
	WriteConfig *RedisConfig
}

// variables for thread-safe one-time (Singleton) initialization
// of Redis object.
var redisReadOncePool, redisWriteOncePool map[string]*resync.Once

// redisWriteConnPool is used only if policy is set to use separate
// instances
var redisReadConnPool, redisWriteConnPool map[string]*redis.Client

type RedisConfig struct {
	Host     string
	Port     int
	DB       int
	Password string
}

// returns a new (or existing, if any) Redis instance (not connection)
// the connection is created for the first time when "get" or "set" is called
// or if Connect() is called explicitly
// If the second parameter for writeConfig is passed nil,
// the policy automatically forces redisPolicyUseSameForReadWrite,
// which literally means it uses the same instance/connection for reading
// and writing. The connections are kept in a pool and each time NewRedis()
// is called, it checks first to see if the pool has any connection for
// current host+port combination or not.
func NewRedis(readConfig *RedisConfig, writeConfig *RedisConfig) (*Redis, error) {
	var policy int
	if readConfig == nil || readConfig.Host == "" || readConfig.Port == 0 {
		return nil, errors.New("redisAdapter: readConfig cannot be empty")
	}
	if writeConfig != nil {
		policy = redisPolicySeparateInstances
	} else {
		policy = redisPolicyUseSameForReadWrite
	}
	return &Redis{
		policy:      policy,
		ReadConfig:  readConfig,
		WriteConfig: writeConfig,
	}, nil
}

// Connect() is optional; If you call it, it connects and does a Ping()
// to the redis server; Else, without calling Connect(), on first calling
// of GET or SET methods of redis, a connection gets established.
func (r *Redis) Connect() error {
	var err error
	_, err = r.getReadInstance()
	if err != nil {
		return err
	}
	if r.policy == redisPolicySeparateInstances && r.WriteConfig != nil {
		_, err = r.getWriteInstance()
		if err != nil {
			return err
		}
	}
	return nil
}

// Disconnects the client (forceful), removes the connection from pool
// and removes its associated sync.Once mutex object
// It is rarely needed to use this method
func DisconnectRead(rc *RedisConfig) {
	c, ok := redisReadConnPool[createKey(rc.Host, rc.Port)]
	if ok && c != nil {
		_ = c.Close()
		delete(redisReadConnPool, createKey(rc.Host, rc.Port))
		delete(redisReadOncePool, createKey(rc.Host, rc.Port))
	}
}

// refer to DisconnectRead() func's doc
func DisconnectWrite(rc *RedisConfig) {
	c, ok := redisWriteConnPool[createKey(rc.Host, rc.Port)]
	if ok && c != nil {
		_ = c.Close()
		delete(redisWriteConnPool, createKey(rc.Host, rc.Port))
		delete(redisWriteOncePool, createKey(rc.Host, rc.Port))
	}
}

// returns the value of the ClientID() by redis
func (r *Redis) GetReadID() int64 {
	c, err := r.getReadInstance()
	if err != nil {
		return 0
	}
	id := c.ClientID().Val()
	h := redisReadConnPool
	_ = h
	return id
}

// returns the value of the ClientID() by redis
func (r *Redis) GetWriteID() int64 {
	c, err := r.getWriteInstance()
	if err != nil {
		return 0
	}
	return c.ClientID().Val()
}

/**
Returns a read/write instance, and creates one if it does not exists
*/
func (r *Redis) getReadInstance() (*redis.Client, error) {
	r.initializePools(2, redisTypeRead)
	var ins = r.getConnFromPoolIfExists(r.ReadConfig.Host, r.ReadConfig.Port, redisTypeRead)
	if ins != nil {
		return ins, nil
	}
	r.createOnceInPoolIfNotExists(r.ReadConfig.Host, r.ReadConfig.Port, redisTypeRead)
	var once = r.getOnceFromPoolIfExists(r.ReadConfig.Host, r.ReadConfig.Port, redisTypeRead)
	var err error
	once.Do(func() {
		var client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf(redisUriPattern, r.ReadConfig.Host, r.ReadConfig.Port),
			Password: r.ReadConfig.Password,
			DB:       r.ReadConfig.DB,
		})

		_, err = client.Ping().Result()
		if err != nil {
			err = errors.New("cannot connect to the redis, got error:" + err.Error())
			return
		}
		redisReadConnPool[r.ReadConfig.Host+":"+strconv.Itoa(r.ReadConfig.Port)] = client
		return
	})

	return redisReadConnPool[r.ReadConfig.Host+":"+strconv.Itoa(r.ReadConfig.Port)], err
}

/**
returns a write instance. Checks the policy to see how it must behave,
if policy says separate read and write instances, then it tries to
create a separate write instance or return an existing one, if the
policy forces the same instance, it call Redis.getReadInstance() method
*/
func (r *Redis) getWriteInstance() (*redis.Client, error) {
	if r.policy == redisPolicyUseSameForReadWrite {
		return r.getReadInstance() // because in such case, an instance is used for
		// for both reading and writing
	}

	r.initializePools(2, redisTypeWrite)
	var ins = r.getConnFromPoolIfExists(r.WriteConfig.Host, r.WriteConfig.Port, redisTypeWrite)
	if ins != nil {
		return ins, nil
	}
	r.createOnceInPoolIfNotExists(r.WriteConfig.Host, r.WriteConfig.Port, redisTypeWrite)
	var once = r.getOnceFromPoolIfExists(r.WriteConfig.Host, r.WriteConfig.Port, redisTypeWrite)
	var err error
	once.Do(func() {
		var client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf(redisUriPattern, r.WriteConfig.Host, r.WriteConfig.Port),
			Password: r.WriteConfig.Password,
			DB:       r.WriteConfig.DB,
		})

		_, err = client.Ping().Result()
		if err != nil {
			err = errors.New("cannot connect to the redis, got error:" + err.Error())
			return
		}
		redisWriteConnPool[r.WriteConfig.Host+":"+strconv.Itoa(r.WriteConfig.Port)] = client
		return
	})

	return redisWriteConnPool[r.WriteConfig.Host+":"+strconv.Itoa(r.WriteConfig.Port)], err
}

func (r *Redis) initializePools(size, connType int) {
	if connType == redisTypeRead {
		if redisReadConnPool == nil {
			redisReadConnPool = make(map[string]*redis.Client, size)
			redisReadOncePool = make(map[string]*resync.Once, size)
		}
	} else if connType == redisTypeWrite {
		if redisWriteConnPool == nil {
			redisWriteConnPool = make(map[string]*redis.Client, size)
			redisWriteOncePool = make(map[string]*resync.Once, size)
		}
	}
	return
}

func (r *Redis) getConnFromPoolIfExists(host string, port int, connType int) *redis.Client {
	var uri = createKey(host, port)
	if connType == redisTypeRead {
		if _, ok := redisReadConnPool[uri]; ok {
			return redisReadConnPool[uri]
		}
	} else if connType == redisTypeWrite {
		if _, ok := redisWriteConnPool[uri]; ok {
			return redisWriteConnPool[uri]
		}
	}
	return nil
}

func (r *Redis) getOnceFromPoolIfExists(host string, port int, connType int) *resync.Once {
	var uri = createKey(host, port)
	if connType == redisTypeRead {
		if _, ok := redisReadOncePool[uri]; ok {
			return redisReadOncePool[uri]
		}
	} else if connType == redisTypeWrite {
		if _, ok := redisWriteOncePool[uri]; ok {
			return redisWriteOncePool[uri]
		}
	}
	return nil
}

func (r *Redis) createOnceInPoolIfNotExists(host string, port int, connType int) *resync.Once {
	var uri = createKey(host, port)
	if connType == redisTypeRead {
		if _, ok := redisReadOncePool[uri]; !ok {
			redisReadOncePool[uri] = &resync.Once{}
		}
	} else if connType == redisTypeWrite {
		if _, ok := redisWriteOncePool[uri]; !ok {
			redisWriteOncePool[uri] = &resync.Once{}
		}
	}
	return nil
}

func createKey(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}

func GetReadConnPool() map[string]*redis.Client {
	return redisReadConnPool
}

func GetWriteConnPool() map[string]*redis.Client {
	return redisWriteConnPool
}

// checks to see if a given error is of type redis.Nil or not
// in other words, it checks if the error is for a not found key
// or not. Use it for get operations.
func (r *Redis) NotFound(err error) bool {
	return err == redis.Nil
}

func (r *Redis) Get(key string) (string, error) {
	conn, err := r.getReadInstance()
	if err != nil {
		return "", err
	}
	res := conn.Get(key)
	return res.Val(), res.Err()
}

func (r *Redis) Scan(key string, i interface{}) (interface{}, error) {
	conn, err := r.getReadInstance()
	if err != nil {
		return nil, err
	}
	err = conn.Get(key).Scan(i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// sets a key in redis with value passed by "value"
// The result of SET command is buffered in a variable r.LastSet
// for further logging and debugging purposes.
func (r *Redis) Set(key string, value interface{}, exp time.Duration) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	conn, err := r.getWriteInstance()
	if err != nil {
		return err
	}
	res := conn.Set(key, value, exp)
	return res.Err()
}

func (r *Redis) Del(keys ...string) bool {
	conn, _ := r.getWriteInstance()
	res := conn.Del(keys...)
	r.LastDel = res
	if res.Err() == nil {
		return true
	}
	return false
}
