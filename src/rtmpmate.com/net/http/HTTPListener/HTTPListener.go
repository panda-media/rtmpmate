package HTTPListener

import (
	"fmt"
	"net/http"
	"strconv"
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
	w.WriteHeader(http.StatusNotFound)
}
