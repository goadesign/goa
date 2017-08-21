package cellar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
)

// ErrNotFound is the error returned when attempting to load a record that does
// not exist.
var ErrNotFound = fmt.Errorf("missing record")

// Bolt is the database driver.
type Bolt struct {
	// client is the Bolt client.
	client *bolt.DB
}

// NewBoltDB creates a Bolt DB database driver given an underlying client.
func NewBoltDB(client *bolt.DB) (*Bolt, error) {
	err := client.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("CELLAR"))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Bolt{client}, nil
}

// NewID returns a unique ID for the given bucket.
func (b *Bolt) NewID(bucket string) (string, error) {
	var sid string
	err := b.client.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		id, err := bkt.NextSequence()
		if err != nil {
			return err
		}
		sid = strconv.FormatUint(id, 10)
		return nil
	})
	return sid, err
}

// Save writes the record to the DB and returns the corresponding new ID.
// data must contain a value that can be marshaled by the encoding/json package.
func (b *Bolt) Save(bucket, id string, data interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return b.client.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if err := bkt.Put([]byte(id), buf); err != nil {
			return err
		}
		return nil
	})
}

// Delete deletes a record by ID.
func (b *Bolt) Delete(bucket, id string) error {
	return b.client.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Delete([]byte(id))
	})
}

// Load reads a record by ID. data is unmarshaled into and should hold a pointer.
func (b *Bolt) Load(bucket, id string, data interface{}) error {
	return b.client.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		v := bkt.Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, data)
	})
}

// LoadAll returns all the records in the given bucket. data should be a pointer
// to a slice. Don't do this in a real service :-)
func (b *Bolt) LoadAll(bucket string, data interface{}) error {
	buf := &bytes.Buffer{}
	err := b.client.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		buf.WriteByte('[')
		if bkt != nil {
			bkt.ForEach(func(_, v []byte) error {
				buf.Write(v)
				return fmt.Errorf("done")
			})
			first := true
			bkt.ForEach(func(_, v []byte) error {
				if first {
					first = false
					return nil
				}
				buf.WriteByte(',')
				buf.Write(v)
				return nil
			})
		}
		buf.WriteByte(']')
		return nil
	})
	if err != nil {
		return err
	}
	return json.Unmarshal(buf.Bytes(), data)
}
