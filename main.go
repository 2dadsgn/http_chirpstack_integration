package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
)

type handler struct {
	json bool
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body) //writes r.Body into b
	if err != nil {
		panic(err)
	}

	event := r.URL.Query().Get("event") //gets the event name from the HTTP request

	switch event {
	case "up":
		err = h.up(b) //<--- UPLINK
	case "join":
		err = h.join(b)
	default:
		log.Printf("handler for event %s is not implemented", event)
		return
	}

	if err != nil {
		log.Printf("handling event '%s' returned error: %s", event, err)
	}
}

func (h *handler) up(b []byte) error {
	var up integration.UplinkEvent
	if err := h.unmarshal(b, &up); err != nil { //there's an issue with this line of code
		return err
	}
	log.Printf("Uplink received from %s with payload: %s", hex.EncodeToString(up.DevEui), hex.EncodeToString(up.Data))
	return nil
}

func (h *handler) join(b []byte) error {
	var join integration.JoinEvent
	if err := h.unmarshal(b, &join); err != nil {
		return err
	}
	log.Printf("Device %s joined with DevAddr %s", hex.EncodeToString(join.DevEui), hex.EncodeToString(join.DevAddr))
	return nil
}

func (h *handler) unmarshal(b []byte, v proto.Message) error {
	if h.json {
		fmt.Println("IT's a json")
		unmarshaler := &jsonpb.Unmarshaler{
			AllowUnknownFields: true, // we don't want to fail on unknown fields
		}
		return unmarshaler.Unmarshal(bytes.NewReader(b), v) //error comes from here
	}
	return proto.Unmarshal(b, v)
}

func main() {
	// json: false   - to handle Protobuf payloads (binary)
	// json: true    - to handle JSON payloads (Protobuf JSON mapping)
	handler := handler{json: true}

	http.Handle("/", &handler)
	log.Fatal(http.ListenAndServe(":3333", nil))
}
