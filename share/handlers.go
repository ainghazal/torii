package share

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	petname "github.com/dustinkirkland/golang-petname"
	bolt "go.etcd.io/bbolt"
)

type httpHandler func(http.ResponseWriter, *http.Request)

// randomPetname returns a two-word petname in the form "fluffy-foobar"
func randomPetname() string {
	petname.NonDeterministicMode()
	return petname.Generate(2, "-")
}

func AddExperimentHandler(db *bolt.DB) httpHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		in, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		fmt.Println("add:", string(in))

		exp := &Experiment{}
		err = json.Unmarshal(in, exp)
		if err != nil {
			http.Error(w, "bad json request", http.StatusBadRequest)
			return
		}
		if exp.Name == "" {
			exp.Name = randomPetname()
			log.Printf("Assigned experiment name: %s\n", exp.Name)
		}

		log.Println("exp max:", exp.Max)

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
			return b.Put([]byte(exp.UUID), buf)
		})

		res := &result{true, exp.UUID}
		json.NewEncoder(w).Encode(res)
	}
}

func ListExperimentHandler(db *bolt.DB) httpHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		sel := []*Experiment{}

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(experimentBucket))

			b.ForEach(func(k, v []byte) error {
				exp := new(Experiment)
				err := json.Unmarshal(v, exp)
				if err == nil {
					sel = append(sel, exp)
				}
				return err
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

func GetExperimentByUUID(db *bolt.DB, uuid string) []*Experiment {
	sel := []*Experiment{}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(experimentBucket))
		v := b.Get([]byte(uuid))

		exp := new(Experiment)
		err := json.Unmarshal(v, exp)
		if err == nil {
			sel = append(sel, exp)
		}
		return err
	})
	return sel
}

func RenderJSONExperimentByUUID(db *bolt.DB) httpHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		uuid := mux.Vars(r)["uuid"]
		sel := GetExperimentByUUID(db, uuid)

		res := []*resultExp{&resultExp{
			OK:   true,
			Data: sel,
		},
		}
		json.NewEncoder(w).Encode(res)
	}
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
