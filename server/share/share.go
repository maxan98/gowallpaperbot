package share

import (
	"bytes"
	clients2 "client/clients"
	"client/wallpaper"
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func getWallpaper(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr)
	clients := clients2.GetInstance()
	splitted := strings.Split(r.RemoteAddr, ":")
	if len(splitted) != 0 {
		clients.AppendClient(splitted[0])
	}

	buffer := new(bytes.Buffer)
	img := wallpaper.GetPicture()
	data, _, err := img.GetPicture()
	if err == nil {
		if err := jpeg.Encode(buffer, data, nil); err != nil {
			log.Println("unable to encode image.")
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
		var res int
		if res, err = w.Write(buffer.Bytes()); err != nil {
			log.Println("unable to write image.")
		}
		fmt.Println("Wrote bytes: ", res)
	} else {
		w.WriteHeader(500)
		w.Write([]byte("Error writing image"))
	}

}

func HandleRequests() {
	http.HandleFunc("/", getWallpaper)
	log.Fatal(http.ListenAndServe(":10000", nil))
}
