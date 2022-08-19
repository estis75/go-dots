package libcoap

/*
#cgo LDFLAGS: -lcoap-3-openssl
#include <coap3/coap.h>
*/
import "time"
import cache "github.com/patrickmn/go-cache"

var caches *cache.Cache

var expirationDefault = 0
var cleanupIntervalDefault = 0

/*
 * Create new cache with with a default expiration time of 'expiration', and which purges expired items every 'cleanupInterval'
 */
func CreateNewCache(expiration int, cleanupInterval int) {
	caches = cache.New(time.Duration(expiration) * time.Second, time.Duration(cleanupInterval) * time.Second)
}
