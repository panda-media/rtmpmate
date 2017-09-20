package HTTPListener

import (
	"fmt"
	"net/http"
	"regexp"
	"rtmpmate.com/net/rtmp/Application"
	"strconv"
)

var (
	urlRe, _ = regexp.Compile("/([a-z0-9.-_]+)(/([a-z0-9.-_]+))?/([a-z0-9-_]+)(.[a-z0-9]+)$")
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

	appName := arr[1]
	instName := arr[3]
	streamName := arr[4]
	extension := arr[5]

	if instName == "" {
		instName = "_definst_"
	}

	switch extension {
	case ".mpd":
		app, err := Application.Get(appName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		inst, _ := app.GetInstance(instName)
		stream, err := inst.GetStream(streamName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		mpd, err := stream.DASHMuxer.GetMPD()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		n, err := w.Write(mpd)
		fmt.Printf("MPD: %d.\n", n)

	default:
		name := "www/" + appName + "/" + instName + "/" + streamName + extension
		http.ServeFile(w, r, name)
	}
}
