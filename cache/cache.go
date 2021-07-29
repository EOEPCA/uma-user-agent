package cache

import "time"

// AuthCache provides a cache of authorization decisions keyed upon the unique
// aspects of the resource access request
type AuthCache struct {
	UserID        string
	IDTokenHash   string
	ResourcePath  string
	RequestMethod string
	Expiry        time.Time
}

// Hash generates a hash from the AuthCache structure elements
func (ac *AuthCache) Hash() string {
	return "zzz"
}
