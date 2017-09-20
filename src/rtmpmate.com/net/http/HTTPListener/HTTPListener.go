package HTTPListener

import (
	"fmt"
	"net/http"
	"regexp"
	"rtmpmate.com/net/rtmp/Application"
	"strconv"
)

var (
	urlRe, _ = regexp.Compile("^http[s]?://[a-z0-9.-]+(:[0-9]+)?/([a-z0-9.-_]+)(/([a-z0-9.-_]+))?(/[a-z0-9.-_]+)(.[a-z0-9]+)$")
)

type HTTPListener struct {
	network string
	port    int
	exiting bool
}

func New() (*HTTPListener, error) {
	var ln HTTPListener
	return &ln, nil
}

func (this *HTTPListener) Listen(network string, port int) {
	if network == "" {
		network = "tcp4"
	}

	if port == 0 {
		port = 80
	}

	address := strconv.Itoa(port)

	http.HandleFunc("/", this.connHandler)
	http.ListenAndServe(":"+address, nil)

	fmt.Printf("%v exiting...\n", this)
}

func (this *HTTPListener) connHandler(w http.ResponseWriter, r *http.Request) {
	arr := urlRe.FindStringSubmatch(r.URL.Path)
	if arr == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if arr[4] == "" {
		arr[4] = "_definst_"
	}
	if arr[3] == "" {
		arr[3] = "/" + arr[4]
	}

	switch arr[6] {
	case ".mpd":
		app, err := Application.Get(arr[2])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		inst, _ := app.GetInstance(arr[3])
		stream, err := inst.GetStream(arr[5])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		mpd, err := stream.DASHMuxer.GetMPD()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(mpd)

	default:
		name := "../www/" + arr[2] + arr[3] + arr[5] + arr[6]
		http.ServeFile(w, r, name)
	}
}
