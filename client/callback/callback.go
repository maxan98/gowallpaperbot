package callback

import (
	"log"
	"net/http"
)

func goCheck(ch chan bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ch <- true
		return
	}
}

func CheckNastyCallback(ch chan bool) {
	http.HandleFunc("/", goCheck(ch))
	log.Fatal(http.ListenAndServe(":10001", nil))
}
