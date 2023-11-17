package leaderboard

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// 8171 years 才会用完
const shift = 32
const mask = (1 << shift) - 1

// 一分钟之内分数相同可以忽略不计，玩家看不出来
func NewScore(val int32) float64 {
	return float64(int64(val)<<shift | time.Now().Unix()/60)
}

func ParseVal(f float64) int32 {
	return int32(int64(f) >> shift)
}

// Leaderboard encapsulates leaderboard-related functionality
type Leaderboard struct {
	client *redis.Client
	opt    *redis.Options
	mu     sync.RWMutex // Mutex for synchronization
	die    chan struct{}
}

// NewLeaderboard creates a new leaderboard instance
func NewLeaderboard(addr, password string, db int) (*Leaderboard, error) {
	opt := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	client := redis.NewClient(opt)
	err := client.Ping().Err()
	if err != nil {
		return nil, fmt.Errorf("connect to redis failed: %v", err)
	}

	lb := &Leaderboard{client, opt, sync.RWMutex{}, make(chan struct{})}
	go lb.checkAndReconnect()
	return lb, nil
}

func (lb *Leaderboard) Close() {
	close(lb.die)
	lb.mu.Lock()
	lb.client.Close()
	lb.mu.Unlock()
}

// reconnect attempts to reconnect to Redis asynchronously
func (lb *Leaderboard) checkAndReconnect() {
	ticker := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-ticker.C:
			lb.mu.RLock()
			c := lb.client // Copy
			lb.mu.RUnlock()
			if c.Ping().Err() == nil { // check
				continue
			}

			c = redis.NewClient(lb.opt)
			if err := c.Ping().Err(); err != nil {
				continue
			}
			lb.mu.Lock()
			lb.client.Close()
			lb.client = c
			lb.mu.Unlock()
		case <-lb.die:
			return
		}
	}
}

func (lb *Leaderboard) Incr(name, id string, incr float64) error {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	// 执行lua 脚本
	return lb.client.Eval(`
	    local score =  redis.call('zscore', KEYS[1], ARGV[2])
		if score  then
		   score = score - score% ARGV[3]
		   ARGV[1] = ARGV[1] +score
		end
		return redis.call('zadd', KEYS[1], ARGV[1], ARGV[2])
	`, []string{name}, incr, id, mask).Err()

}

// CreateLeaderboard creates a leaderboard with the given name and sets the expiration time
func (lb *Leaderboard) CreateLeaderboard(name string, expirationTime time.Time) error {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	err := lb.client.ExpireAt(name, expirationTime).Err()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	return err
}

// UpdateScore updates a player's score in the specifiedferboard
func (lb *Leaderboard) Set(name, id string, score float64) error {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.client.ZAdd(name, redis.Z{Score: float64(score), Member: id}).Err()
}

// GetLeaderboard retrieves the leaderboard data by name
func (lb *Leaderboard) GetLeaderboard(name string) ([]redis.Z, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result, err := lb.client.ZRevRangeWithScores(name, 0, -1).Result()
	return result, err
}

// GetPlayerRank retrieves the rank of a specified player in the specified leaderboard
func (lb *Leaderboard) GetRank(name, id string) (int, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	rank, err := lb.client.ZRevRank(name, id).Result()
	if err == redis.Nil {
		return -1, fmt.Errorf("player '%s' not found in the leaderboard '%s'", id, name)
	} else if err != nil {
		return -1, err
	}
	return int(rank + 1), nil
}

// GetTopNPlayers retrieves the top N players from the specified leaderboard
func (lb *Leaderboard) TopN(name string, n int) ([]redis.Z, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.client.ZRevRangeWithScores(name, 0, int64(n-1)).Result()
}

// CheckAndCreateLeaderboard checks if the specified leaderboard exists, and creates it if not
func (lb *Leaderboard) CheckAndCreateLeaderboard(name string) error {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	key := name
	exists, err := lb.client.Exists(key).Result()
	if err != nil {
		return err
	}

	if exists == 0 {
		return lb.client.ExpireAt(name, time.Now().Add(24*time.Hour)).Err()

	}
	return nil
}
