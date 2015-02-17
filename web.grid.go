/*

	Web.Grid is the collection of http based grid blocks for the masses

*/

package webgrid

import (
	// "encoding/json"
	// "errors"
	"github.com/influx6/grids"
	// "io"
	// "io/ioutil"
	// "net"
	"net/http"
	// "net/url"
	// "os"
	// "strconv"
	// "strings"
	// "sync"
	// "time"
	// "unicode/utf8"
)

type httpPacket struct {
	Id  string
	res http.ResponseWriter
	req *http.Request
}

type HttpServ struct {
	*grids.Grid
}

type MethodFilter struct {
	method string
	*grids.Grid
}

func (h *HttpServ) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	vals := &httpPacket{"httpPacket", rw, r}

	if c := h.Out("req"); c != nil {
		go func() {
			c <- vals
		}()
	}
}

func NewServo() *HttpServ {
	sv := &HttpServ{grids.NewGrid("ServIO")}
	return sv
}
