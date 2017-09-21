package HTTPListener

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"rtmpmate.com/muxer/DASHMuxer"
	HTTP "rtmpmate.com/net/http"
	RTMP "rtmpmate.com/net/rtmp"
	"rtmpmate.com/net/rtmp/Application"
	"strconv"
)

var (
	urlRe, _ = regexp.Compile("/([a-z0-9.-_]+)(/([a-z0-9.-_]+))?/([a-z0-9-_]+)/([a-z0-9-_]+)(.[a-z0-9]+)$")
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
	if _, err := os.Stat(HTTP.WEBROOT); os.IsNotExist(err) {
		err = os.MkdirAll(HTTP.WEBROOT, os.ModeDir)
		if err != nil {
			return
		}
	}

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
	if arr != nil {
		appName := arr[1]
		instName := arr[3]
		streamName := arr[4]
		fileName := arr[5]
		extension := arr[6]

		if instName == "" {
			instName = "_definst_"
		}

		this.appendHeadersForStream(w)

		switch extension {
		case ".mpd":
			this.serveMPD(w, r, appName, instName, streamName, fileName)
			return

		default:
			name := RTMP.APPLICATIONS + appName + "/" + instName + "/" + streamName + "/" + fileName + extension
			http.ServeFile(w, r, name)
			return
		}
	}

	http.ServeFile(w, r, HTTP.WEBROOT+r.URL.Path)
}

func (this *HTTPListener) serveMPD(w http.ResponseWriter, r *http.Request,
	appName string, instName string, streamName string, fileName string) {
	if fileName != DASHMuxer.MPD_FILENAME {
		w.WriteHeader(http.StatusNotFound)
		return
	}

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
}

func (this *HTTPListener) appendHeadersForStream(w http.ResponseWriter) {
	for k, v := range HTTP.STREAM_HEADERS {
		w.Header().Set(k, v)
	}
}
