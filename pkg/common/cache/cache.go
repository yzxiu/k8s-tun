package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var c = cache.New(30*time.Minute, 10*time.Minute)

func GetCache() *cache.Cache {
	return c
}
