package jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"net/http"
	"sync"
)

type (
	// Key represents a public key used to validate the incoming token signatures.
	// The value must be of type *rsa.PublicKey, *ecdsa.PublicKey, []byte or string.
	// Keys of type []byte or string are interpreted depending on the incoming request JWT token
	// method (HMAC, RSA, etc.).
	Key interface{}

	// KeyResolver allows the management of keys used by the middleware to verify the signature of
	// incoming requests. Keys are grouped by name allowing the authorization algorithm to select a
	// group depending on the incoming request state (e.g. a header). The use of groups enables key
	// rotation.
	KeyResolver interface {
		// SelectKeys returns the group of keys to be used for the incoming request.
		SelectKeys(req *http.Request) []Key
	}

	// GroupResolver is a key resolver that switches on the value of a specified request header
	// for selecting the key group used to authorize the incoming request.
	GroupResolver struct {
		*sync.RWMutex
		keyHeader string
		keyMap    map[string][]Key
	}

	// simpleResolver uses a single immutable key group.
	simpleResolver []Key
)

// NewResolver returns a GroupResolver that uses the value of the request header with the given name
// to select the key group used for authorization. keys contains the initial set of key groups
// indexed by name.
func NewResolver(keys map[string][]Key, header string) (*GroupResolver, error) {
	if header == "" {
		return nil, ErrEmptyHeaderName
	}
	keyMap := make(map[string][]Key)
	for name := range keys {
		for _, keys := range keys[name] {
			switch keys := keys.(type) {
			case *rsa.PublicKey, *ecdsa.PublicKey, string, []byte:
				keyMap[name] = append(keyMap[name], keys)
			case []*rsa.PublicKey:
				for _, key := range keys {
					keyMap[name] = append(keyMap[name], key)
				}
			case []*ecdsa.PublicKey:
				for _, key := range keys {
					keyMap[name] = append(keyMap[name], key)
				}
			case [][]byte:
				for _, key := range keys {
					keyMap[name] = append(keyMap[name], key)
				}
			case []string:
				for _, key := range keys {
					keyMap[name] = append(keyMap[name], key)
				}
			default:
				return nil, ErrInvalidKey
			}
		}
	}
	return &GroupResolver{
		RWMutex:   &sync.RWMutex{},
		keyMap:    keyMap,
		keyHeader: header,
	}, nil
}

// NewSimpleResolver returns a simple resolver.
func NewSimpleResolver(keys []Key) KeyResolver {
	return simpleResolver(keys)
}

// AddKeys can be used to add keys to the resolver which will be referenced
// by the provided name. Acceptable types for keys include string, []string,
// *rsa.PublicKey or []*rsa.PublicKey. Multiple keys are allowed for a single
// key name to allow for key rotation.
func (kr *GroupResolver) AddKeys(name string, keys Key) error {
	kr.Lock()
	defer kr.Unlock()
	switch keys := keys.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey, []byte, string:
		kr.keyMap[name] = append(kr.keyMap[name], keys)
	case []*rsa.PublicKey:
		for _, key := range keys {
			kr.keyMap[name] = append(kr.keyMap[name], key)
		}
	case []*ecdsa.PublicKey:
		for _, key := range keys {
			kr.keyMap[name] = append(kr.keyMap[name], key)
		}
	case [][]byte:
		for _, key := range keys {
			kr.keyMap[name] = append(kr.keyMap[name], key)
		}
	case []string:
		for _, key := range keys {
			kr.keyMap[name] = append(kr.keyMap[name], key)
		}
	default:
		return ErrInvalidKey
	}
	return nil
}

// RemoveAllKeys removes all keys from the resolver.
func (kr *GroupResolver) RemoveAllKeys() {
	kr.Lock()
	defer kr.Unlock()
	kr.keyMap = make(map[string][]Key)
	return
}

// RemoveKeys removes all keys from the resolver stored under the provided name.
func (kr *GroupResolver) RemoveKeys(name string) {
	kr.Lock()
	defer kr.Unlock()
	delete(kr.keyMap, name)
	return
}

// RemoveKey removes only the provided key stored under the provided name from
// the resolver.
func (kr *GroupResolver) RemoveKey(name string, key Key) {
	kr.Lock()
	defer kr.Unlock()
	if keys, ok := kr.keyMap[name]; ok {
		for i, keyItem := range keys {
			if keyItem == key {
				kr.keyMap[name] = append(keys[:i], keys[i+1:]...)
			}
		}
	}
	return
}

// GetAllKeys returns a list of all the keys stored in the resolver.
func (kr *GroupResolver) GetAllKeys() []Key {
	kr.RLock()
	defer kr.RUnlock()
	var keys []Key
	for name := range kr.keyMap {
		for _, key := range kr.keyMap[name] {
			keys = append(keys, key)
		}
	}
	return keys
}

// GetKeys returns a list of all the keys stored in the resolver under the
// provided name.
func (kr *GroupResolver) GetKeys(name string) ([]Key, error) {
	kr.RLock()
	defer kr.RUnlock()
	if keys, ok := kr.keyMap[name]; ok {
		return keys, nil
	}
	return nil, ErrKeyDoesNotExist
}

// SelectKeys returns the keys in the group with the name identified by the request key selection
// header. If the header does value does not match a specific group then all keys are returned.
func (kr *GroupResolver) SelectKeys(req *http.Request) []Key {
	keyName := req.Header.Get(kr.keyHeader)
	kr.RLock()
	defer kr.RUnlock()
	if keyName != "" {
		return kr.keyMap[keyName]
	}
	var keys []Key
	for _, ks := range kr.keyMap {
		keys = append(keys, ks...)
	}
	return keys
}

// SelectKeys returns the keys used to create the simple resolver.
func (sr simpleResolver) SelectKeys(req *http.Request) []Key {
	return []Key(sr)
}
