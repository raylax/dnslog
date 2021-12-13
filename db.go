package main

import (
	"fmt"
	"sync"
	"time"
)

var emptyRecords = make([]*Record, 0, 0)

type Record struct {
	Domain string   `json:"domain"`
	IP     string   `json:"ip"`
	Time   JSONTime `json:"time"`
}

type nameCache struct {
	time    time.Time
	records []*Record
	m       sync.Mutex
}

type db struct {
	cacheMap map[string]*nameCache
	ttl      time.Duration
	max      int
	m        sync.Mutex
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

func newDB(ttl time.Duration, max int) *db {
	db := &db{
		cacheMap: make(map[string]*nameCache),
		ttl:      ttl,
		max:      max,
	}
	db.initRecovery()
	return db
}

func (db *db) AddRecord(name string, domain string, ip string) {
	now := time.Now()
	db.m.Lock()
	cache, ok := db.cacheMap[name]
	if !ok {
		cache = &nameCache{
			records: make([]*Record, 0, db.max),
		}
		db.cacheMap[name] = cache
	}
	cache.time = now
	db.m.Unlock()
	cache.m.Lock()
	defer cache.m.Unlock()
	if len(cache.records) == db.max {
		cache.records = cache.records[1:]
	}
	cache.records = append(cache.records, &Record{
		Domain: domain,
		IP:     ip,
		Time:   JSONTime(now),
	})
}

func (db *db) GetRecords(name string) []*Record {
	if cache, ok := db.cacheMap[name]; ok {
		return cache.records
	}
	return emptyRecords
}

func (db *db) initRecovery() {
	go func() {
		for {
			expireAt := time.Now().Add(-db.ttl)
			time.Sleep(1 * time.Second)
			db.m.Lock()
			for k, cache := range db.cacheMap {
				if cache.time.Before(expireAt) {
					delete(db.cacheMap, k)
				}
			}
			db.m.Unlock()
		}
	}()
}
