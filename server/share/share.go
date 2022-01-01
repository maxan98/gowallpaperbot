package share

import (
	"bytes"
	clients2 "client/clients"
	"client/settings"
	"client/wallpaper"
	"fmt"
	log "github.com/sirupsen/logrus"
	"image/jpeg"
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
	u, p, ok := r.BasicAuth()
	if !ok {
		log.Info("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}
	set := settings.GetInstance()
	setu := set.GetUsername()
	setp := set.GetPassword()
	if u != setu || p != setp {
		log.Info("Bad auth..")
		w.WriteHeader(401)
		return
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
			log.Info("unable to write image.")
		}
		log.Info("Wrote bytes: ", res)
	} else {
		w.WriteHeader(500)
		w.Write([]byte("Error writing image"))
	}

}

func HandleRequests() {
	set := settings.GetInstance()
	http.HandleFunc("/", getWallpaper)
	log.Fatal(http.ListenAndServeTLS(":10000", set.GetCertFile(), set.GetCertKey(), nil))
}
