package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/creepfmd/jsonpath"
	"github.com/gorilla/mux"
)

// our main function
func main() {
	log.Println("Setting router...")
	router := mux.NewRouter()
	router.HandleFunc("/{arrayPath}", splitMessage).Methods("POST")
	log.Println("Listening 8082...")
	log.Fatal(http.ListenAndServe(":8082", router))
}

func splitMessage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var jsonData interface{}
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jpath, _ := jsonpath.Compile((string)(params["arrayPath"]) + "[:]")
	jpathSteps := jpath.GetSteps()

	res, _ := jpath.Lookup(jsonData)

	clearBody, _ := json.Marshal(jsonData)
	switch x := res.(type) {
	case []interface{}:
		nodeName := jpathSteps[len(jpathSteps)-1]
		rgxp := regexp.MustCompile(`"` + nodeName + `":\[.*\]`)
		var messageParts []string
		for _, e := range x {
			dummy, _ := json.Marshal(e)
			messageParts = append(messageParts, rgxp.ReplaceAllString((string)(clearBody[:]), `"`+nodeName+`":`+(string)(dummy[:])))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[` + strings.Join(messageParts, `,`) + `]`))
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(clearBody))
	}
}
