package miniredis

// Commands to modify and query our databases directly.

import (
	"errors"
)

var (
	// ErrKeyNotFound is returned when a key doesn't exist.
	ErrKeyNotFound = errors.New(msgKeyNotFound)
	// ErrWrongType when a key is not the right type.
	ErrWrongType = errors.New(msgWrongType)
	// ErrIntValueError can returned by INCRBY
	ErrIntValueError = errors.New(msgInvalidInt)
	// ErrFloatValueError can returned by INCRBYFLOAT
	ErrFloatValueError = errors.New(msgInvalidFloat)
)

// Select sets the DB id for all direct commands.
func (m *Miniredis) Select(i int) {
	m.Lock()
	defer m.Unlock()
	m.selectedDB = i
}

// Get returns string keys added with SET.
func (m *Miniredis) Get(k string) (string, error) {
	return m.DB(m.selectedDB).Get(k)
}

// Get returns a string key
func (db *RedisDB) Get(k string) (string, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return "", ErrKeyNotFound
	}
	if db.t(k) != "string" {
		return "", ErrWrongType
	}
	return db.get(k), nil
}

// Set sets a string key. Removes expire.
func (m *Miniredis) Set(k, v string) error {
	return m.DB(m.selectedDB).Set(k, v)
}

// Set sets a string key. Removes expire.
// Unlike redis the key can't be an existing non-string key.
func (db *RedisDB) Set(k, v string) error {
	db.master.Lock()
	defer db.master.Unlock()

	if db.exists(k) && db.t(k) != "string" {
		return ErrWrongType
	}
	db.del(k, true) // Remove expire
	db.set(k, v)
	return nil
}

// Incr changes a int string value by delta.
func (m *Miniredis) Incr(k string, delta int) (int, error) {
	return m.DB(m.selectedDB).Incr(k, delta)
}

// Incr changes a int string value by delta.
func (db *RedisDB) Incr(k string, delta int) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if db.exists(k) && db.t(k) != "string" {
		return 0, ErrWrongType
	}

	return db.incr(k, delta)
}

// Incrfloat changes a float string value by delta.
func (m *Miniredis) Incrfloat(k string, delta float64) (float64, error) {
	return m.DB(m.selectedDB).Incrfloat(k, delta)
}

// Incrfloat changes a float string value by delta.
func (db *RedisDB) Incrfloat(k string, delta float64) (float64, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if db.exists(k) && db.t(k) != "string" {
		return 0, ErrWrongType
	}

	return db.incrfloat(k, delta)
}

// List returns the list k, or an error if it's not there or something else.
// This is the same as the Redis command `LRANGE 0 -1`, but you can do your own
// range-ing.
func (m *Miniredis) List(k string) ([]string, error) {
	return m.DB(m.selectedDB).List(k)
}

// List returns the list k, or an error if it's not there or something else.
// This is the same as the Redis command `LRANGE 0 -1`, but you can do your own
// range-ing.
func (db *RedisDB) List(k string) ([]string, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if !db.exists(k) {
		return nil, ErrKeyNotFound
	}
	if db.t(k) != "list" {
		return nil, ErrWrongType
	}
	return db.listKeys[k], nil
}

// Lpush is an unshift. Returns the new length.
func (m *Miniredis) Lpush(k, v string) (int, error) {
	return m.DB(m.selectedDB).Lpush(k, v)
}

// Lpush is an unshift. Returns the new length.
func (db *RedisDB) Lpush(k, v string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if db.exists(k) && db.t(k) != "list" {
		return 0, ErrWrongType
	}
	return db.lpush(k, v), nil
}

// Lpop is a shift. Returns the popped element.
func (m *Miniredis) Lpop(k string) (string, error) {
	return m.DB(m.selectedDB).Lpop(k)
}

// Lpop is a shift. Returns the popped element.
func (db *RedisDB) Lpop(k string) (string, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if !db.exists(k) {
		return "", ErrKeyNotFound
	}
	if db.t(k) != "list" {
		return "", ErrWrongType
	}
	return db.lpop(k), nil
}

// Push add element at the end. Is called RPUSH in redis. Returns the new length.
func (m *Miniredis) Push(k string, v ...string) (int, error) {
	return m.DB(m.selectedDB).Push(k, v...)
}

// Push add element at the end. Is called RPUSH in redis. Returns the new length.
func (db *RedisDB) Push(k string, v ...string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if db.exists(k) && db.t(k) != "list" {
		return 0, ErrWrongType
	}
	return db.push(k, v...), nil
}

// Pop removes and returns the last element. Is called RPOP in Redis.
func (m *Miniredis) Pop(k string) (string, error) {
	return m.DB(m.selectedDB).Pop(k)
}

// Pop removes and returns the last element. Is called RPOP in Redis.
func (db *RedisDB) Pop(k string) (string, error) {
	db.master.Lock()
	defer db.master.Unlock()

	if !db.exists(k) {
		return "", ErrKeyNotFound
	}
	if db.t(k) != "list" {
		return "", ErrWrongType
	}

	return db.pop(k), nil
}

// SetAdd adds keys to a set. Returns the number of new keys.
func (m *Miniredis) SetAdd(k string, elems ...string) (int, error) {
	return m.DB(m.selectedDB).SetAdd(k, elems...)
}

// SetAdd adds keys to a set. Returns the number of new keys.
func (db *RedisDB) SetAdd(k string, elems ...string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if db.exists(k) && db.t(k) != "set" {
		return 0, ErrWrongType
	}
	return db.setadd(k, elems...), nil
}

// Members gives all set keys. Sorted.
func (m *Miniredis) Members(k string) ([]string, error) {
	return m.DB(m.selectedDB).Members(k)
}

// Members gives all set keys. Sorted.
func (db *RedisDB) Members(k string) ([]string, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return nil, ErrKeyNotFound
	}
	if db.t(k) != "set" {
		return nil, ErrWrongType
	}
	return db.members(k), nil
}

// IsMember tells if value is in the set.
func (m *Miniredis) IsMember(k, v string) (bool, error) {
	return m.DB(m.selectedDB).IsMember(k, v)
}

// IsMember tells if value is in the set.
func (db *RedisDB) IsMember(k, v string) (bool, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return false, ErrKeyNotFound
	}
	if db.t(k) != "set" {
		return false, ErrWrongType
	}
	return db.isMember(k, v), nil
}

// HKeys returns all keys ('fields') for a hash key.
func (m *Miniredis) HKeys(k string) ([]string, error) {
	return m.DB(m.selectedDB).HKeys(k)
}

// HKeys returns all keys ('fields') for a hash key.
func (db *RedisDB) HKeys(key string) ([]string, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(key) {
		return nil, ErrKeyNotFound
	}
	if db.t(key) != "hash" {
		return nil, ErrWrongType
	}
	return db.hkeys(key), nil
}

// Del deletes a key and any expiration value. Returns whether there was a key.
func (m *Miniredis) Del(k string) bool {
	return m.DB(m.selectedDB).Del(k)
}

// Del deletes a key and any expiration value. Returns whether there was a key.
func (db *RedisDB) Del(k string) bool {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return false
	}
	db.del(k, true)
	return true
}

// Expire value. As set by the client (via EXPIRE, PEXPIRE, EXPIREAT, PEXPIREAT and
// similar commands). 0 if not set.
func (m *Miniredis) Expire(k string) int {
	return m.DB(m.selectedDB).Expire(k)
}

// Expire value. As set by the client (via EXPIRE, PEXPIRE, EXPIREAT, PEXPIREAT and
// similar commands). 0 if not set.
func (db *RedisDB) Expire(k string) int {
	db.master.Lock()
	defer db.master.Unlock()
	return db.expire[k]
}

// SetExpire sets expiration of a key.
func (m *Miniredis) SetExpire(k string, ex int) {
	m.DB(m.selectedDB).SetExpire(k, ex)
}

// SetExpire sets expiration of a key.
func (db *RedisDB) SetExpire(k string, ex int) {
	db.master.Lock()
	defer db.master.Unlock()
	db.expire[k] = ex
	db.keyVersion[k]++
}

// Type gives the type of a key, or ""
func (m *Miniredis) Type(k string) string {
	return m.DB(m.selectedDB).Type(k)
}

// Type gives the type of a key, or ""
func (db *RedisDB) Type(k string) string {
	db.master.Lock()
	defer db.master.Unlock()
	return db.t(k)
}

// Exists tells whether a key exists.
func (m *Miniredis) Exists(k string) bool {
	return m.DB(m.selectedDB).Exists(k)
}

// Exists tells whether a key exists.
func (db *RedisDB) Exists(k string) bool {
	db.master.Lock()
	defer db.master.Unlock()
	return db.exists(k)
}

// HGet returns hash keys added with HSET.
// This will return an empty string if the key is not set. Redis would return
// a nil.
// Returns empty string when the key is of a different type.
func (m *Miniredis) HGet(k, f string) string {
	return m.DB(m.selectedDB).HGet(k, f)
}

// HGet returns hash keys added with HSET.
// Returns empty string when the key is of a different type.
func (db *RedisDB) HGet(k, f string) string {
	db.master.Lock()
	defer db.master.Unlock()
	h, ok := db.hashKeys[k]
	if !ok {
		return ""
	}
	return h[f]
}

// HSet sets a hash key.
// If there is another key by the same name it will be gone.
func (m *Miniredis) HSet(k, f, v string) {
	m.DB(m.selectedDB).HSet(k, f, v)
}

// HSet sets a hash key.
// If there is another key by the same name it will be gone.
func (db *RedisDB) HSet(k, f, v string) {
	db.master.Lock()
	defer db.master.Unlock()
	db.hset(k, f, v)
}

// HDel deletes a hash key.
func (m *Miniredis) HDel(k, f string) {
	m.DB(m.selectedDB).HDel(k, f)
}

// HDel deletes a hash key.
func (db *RedisDB) HDel(k, f string) {
	db.master.Lock()
	defer db.master.Unlock()
	db.hdel(k, f)
}

func (db *RedisDB) hdel(k, f string) {
	if _, ok := db.hashKeys[k]; !ok {
		return
	}
	delete(db.hashKeys[k], f)
	db.keyVersion[k]++
}

// HIncr increases a key/field by delta (int).
func (m *Miniredis) HIncr(k, f string, delta int) (int, error) {
	return m.DB(m.selectedDB).HIncr(k, f, delta)
}

// HIncr increases a key/field by delta (int).
func (db *RedisDB) HIncr(k, f string, delta int) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()
	return db.hincr(k, f, delta)
}

// HIncrfloat increases a key/field by delta (float).
func (m *Miniredis) HIncrfloat(k, f string, delta float64) (float64, error) {
	return m.DB(m.selectedDB).HIncrfloat(k, f, delta)
}

// HIncrfloat increases a key/field by delta (float).
func (db *RedisDB) HIncrfloat(k, f string, delta float64) (float64, error) {
	db.master.Lock()
	defer db.master.Unlock()
	return db.hincrfloat(k, f, delta)
}

// SRem removes fields from a set. Returns number of deleted fields.
func (m *Miniredis) SRem(k string, fields ...string) (int, error) {
	return m.DB(m.selectedDB).SRem(k, fields...)
}

// SRem removes fields from a set. Returns number of deleted fields.
func (db *RedisDB) SRem(k string, fields ...string) (int, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return 0, ErrKeyNotFound
	}
	if db.t(k) != "set" {
		return 0, ErrWrongType
	}
	return db.setrem(k, fields...), nil
}

// ZAdd adds a score,member to a sorted set.
func (m *Miniredis) ZAdd(k string, score float64, member string) (bool, error) {
	return m.DB(m.selectedDB).ZAdd(k, score, member)
}

// ZAdd adds a score,member to a sorted set.
func (db *RedisDB) ZAdd(k string, score float64, member string) (bool, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if db.exists(k) && !db.exists(k) && db.t(k) != "zset" {
		return false, ErrWrongType
	}
	return db.zadd(k, score, member), nil
}

// ZMembers returns all members by score
func (m *Miniredis) ZMembers(k string) ([]string, error) {
	return m.DB(m.selectedDB).ZMembers(k)
}

// ZMembers returns all members by score
func (db *RedisDB) ZMembers(k string) ([]string, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return nil, ErrKeyNotFound
	}
	if db.t(k) != "zset" {
		return nil, ErrWrongType
	}
	return db.zmembers(k), nil
}

// SortedSet returns a raw string->float64 map.
func (m *Miniredis) SortedSet(k string) (map[string]float64, error) {
	return m.DB(m.selectedDB).SortedSet(k)
}

// SortedSet returns a raw string->float64 map.
func (db *RedisDB) SortedSet(k string) (map[string]float64, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return nil, ErrKeyNotFound
	}
	if db.t(k) != "zset" {
		return nil, ErrWrongType
	}
	return db.sortedSet(k), nil
}

// ZRem deletes a member. Returns whether the was a key.
func (m *Miniredis) ZRem(k, member string) (bool, error) {
	return m.DB(m.selectedDB).ZRem(k, member)
}

// ZRem deletes a member. Returns whether the was a key.
func (db *RedisDB) ZRem(k, member string) (bool, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return false, ErrKeyNotFound
	}
	if db.t(k) != "zset" {
		return false, ErrWrongType
	}
	return db.zrem(k, member), nil
}

// ZScore gives the score of a sorted set member.
func (m *Miniredis) ZScore(k, member string) (float64, error) {
	return m.DB(m.selectedDB).ZScore(k, member)
}

// ZScore gives the score of a sorted set member.
func (db *RedisDB) ZScore(k, member string) (float64, error) {
	db.master.Lock()
	defer db.master.Unlock()
	if !db.exists(k) {
		return 0, ErrKeyNotFound
	}
	if db.t(k) != "zset" {
		return 0, ErrWrongType
	}
	return db.zscore(k, member), nil
}
