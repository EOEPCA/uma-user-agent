package authcache

import "time"

// AuthCacheEntry provides a cache of authorization decisions keyed upon the unique
// aspects of the resource access request
type AuthCacheEntry struct {
	UserID        string
	IDTokenHash   string
	ResourcePath  string
	RequestMethod string
	Expiry        time.Time
}

// AuthCache is a collection of cached authorization decisions
type AuthCache map[string]AuthCacheEntry

// Hash generates a hash from the AuthCacheEntry structure elements
func (ac *AuthCacheEntry) Hash() string {
	return "zzz"
}
