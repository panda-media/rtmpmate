package http

import ()

const (
	WEBROOT = "webroot/"
)

var (
	STREAM_HEADERS map[string]string
)

func init() {
	STREAM_HEADERS = make(map[string]string)
	STREAM_HEADERS["Access-Control-Allow-Origin"] = "*"
}
