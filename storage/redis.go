package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	raven "github.com/getsentry/raven-go"
	"github.com/kpango/glg"
)

const (
	sessionsPrefix = "sessions"
)

// Cache will be a generic wrapper around a Redis cache.
type Cache struct {
	*redis.Pool
}

// NewCache will create a new cache instance and the required Redis connection pool.
func NewCache(addr string) *Cache {
	// 25 is the maximum number of active connections for the Heroku Redis free tier
	return &Cache{&redis.Pool{
		MaxIdle:     3,
		MaxActive:   25,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(addr) },
	}}
}

// GetSession will attempt to read a session from the cache, if an existing one is not found,
// an empty session will be created with the specified sessionID.
func (c *Cache) GetSession(id string, session interface{}) error {

	conn := c.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s:%s", sessionsPrefix, id)
	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		// NOTE: This is a normal situation, if the session is not stored in the cache,
		// it will hit this condition.
		return err
	}

	json.Unmarshal([]byte(reply), session)

	return nil
}

// SaveSession will persist the given session to the cache. This will allow support for long running
// Alexa sessions that continually prompt the user for more information.
func (c *Cache) SaveSession(id string, session interface{}) {

	conn := c.Get()
	defer conn.Close()

	sessionBytes, err := json.Marshal(session)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Couldn't marshal session to string: %s", err.Error())
		return
	}

	key := fmt.Sprintf("%s:%s", sessionsPrefix, id)
	_, err = conn.Do("SET", key, string(sessionBytes))
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to set session: %s", err.Error())
	}
}

// ClearSession will remove the specified session from the local cache, this will be done
// when the user completes a full request session.
func (c *Cache) ClearSession(id string) {

	conn := c.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s:%s", sessionsPrefix, id)
	_, err := conn.Do("DEL", key)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to delete the session from the Redis cache: %s", err.Error())
	}
}
