package redis

import (
	xredis "github.com/gomodule/redigo/redis"
	"github.com/pfhds/live-stream-ai/models"
	"github.com/pfhds/live-stream-ai/utils/log"
)

var r *Pool

// pool
type Pool struct {
	xredis.Pool
}

func Init(c *models.RedisConfig) {
	r = New(c)
}

// new pool
func New(c *models.RedisConfig) (p *Pool) {
	if c.DialTimeout <= 0 || c.ReadTimeout <= 0 || c.WriteTimeout <= 0 {
		panic("must config redis timeout")
	}

	dialFunc := func() (xredis.Conn, error) {
		return xredis.Dial(
			"tcp",
			c.Addr,
			xredis.DialConnectTimeout(c.DialTimeout),
			xredis.DialReadTimeout(c.ReadTimeout),
			xredis.DialWriteTimeout(c.WriteTimeout),
			xredis.DialDatabase(c.Db))
	}

	return &Pool{xredis.Pool{
		MaxIdle:     c.MaxIdle,
		MaxActive:   c.MaxActive,
		IdleTimeout: c.IdleTimeout,
		Dial:        dialFunc}}
}

// set key value
func SetKey(key string, val string, expire int64) (err error) {
	conn := r.Get()
	defer conn.Close()

	if _, err = xredis.String(conn.Do("SETEX", key, expire*60, val)); err != nil {
		if err == xredis.ErrNil {
			err = nil
		} else {
			log.Error("SETEX(%s) error(%s)", key, err)
		}
		return err
	}
	return err
}

// get key
func GetKey(key string) (val string, err error) {
	conn := r.Get()
	defer conn.Close()
	if val, err = xredis.String(conn.Do("GET", key)); err != nil {
		if err == xredis.ErrNil {
			err = nil
		} else {
			log.Error("GetKey(%s) error(%v)", key, err)
		}
		return
	}
	return
}

// get all key
func GetAllKeys(prefix string, limit int) ([]string, error) {
	conn := r.Get()
	defer conn.Close()

	keys, err := xredis.Strings(conn.Do("KEYS", prefix+"*"))

	return keys, err
}

func RemoveKey(key string) error {
	conn := r.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)

	return err
}

func RemoveAllKey(prefix string, maxai int) (err error) {
	conn := r.Get()
	defer conn.Close()
	keys, err := GetAllKeys(prefix, maxai)
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err := conn.Do("DEL", key)
		if err != nil {
			break
		}
	}

	return err
}
