package share

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

type httpHandler func(http.ResponseWriter, *http.Request)

func AddExperimentHandler(db *bolt.DB) httpHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		in, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		fmt.Println("add:", string(in))

		exp := &experiment{}
		err = json.Unmarshal(in, exp)
		if err != nil {
			http.Error(w, "bad json request", http.StatusBadRequest)
			return
		}

		// TODO do validate empty fields etc
		rawUUID := uuid.New()
		exp.UUID = strings.Replace(rawUUID.String(), "-", "", -1)

		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(experimentBucket))
			id, _ := b.NextSequence()
			exp.ID = int(id)

			buf, err := json.Marshal(exp)
			if err != nil {
				return err
			}
			// persist experiment in its own bucket
			return b.Put(itob(exp.ID), buf)
		})

		res := &result{true, exp.UUID}
		json.NewEncoder(w).Encode(res)
	}
}

func ListExperimentHandler(db *bolt.DB) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		sel := []*experiment{}

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(experimentBucket))

			b.ForEach(func(k, v []byte) error {
				exp := new(experiment)
				err := json.Unmarshal(v, exp)
				if err == nil {
					sel = append(sel, exp)
				}
				return nil
			})
			return nil
		})
		res := []*resultExp{&resultExp{
			OK:   true,
			Data: sel,
		},
		}
		json.NewEncoder(w).Encode(res)
	}
}

// TODO I'm mixing up render + handler here. Decouple :)
func RenderExperimentByUUID(db *bolt.DB) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
