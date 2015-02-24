/*

	Web.Grid is the collection of http based grid blocks for the masses

*/

package webgrid

import (
	"github.com/influx6/fsgrid"
	"github.com/influx6/grids"
	"net/http"
	// "encoding/json"
	// "errors"
	// "io"
	// "io/ioutil"
	// "net"
	// "net/url"
	// "os"
	// "strconv"
	// "strings"
	// "sync"
	// "time"
	// "unicode/utf8"
)

type HttpServ struct {
	*grids.Grid
}

type StaticServ struct {
	*grids.Grid
	Base    string
	Control *fsgrid.FSControl
	Reader  *fsgrid.FSFile
	// Render      *Render
	// ErrorRender *Render
}

type MethodFilter struct {
	method string
	*grids.Grid
}

type Render struct {
	Handler func(i interface{}, r interface{})
	*grids.Grid
}

func NewRender(format string, m func(i interface{}, f interface{})) *Render {
	r := &Render{m, grids.NewGrid(format)}

	r.NewIn("req")
	r.NewOut("err")

	reqChan := r.In("req")

	reqChan.Or(func(i interface{}) {
		r.Handler(i, r)
	})

	return r
}

func FileRender() *Render {
	r := NewRender("FileRender", func(i interface{}, t interface{}) {
		_, erc := t.(*Render)

		if !erc {
			return
		}

		resq, err := i.(*grids.GridPacket)

		if !err {
			// core.OutSend("err", resq)
			return
		}

		req, ok := resq.Body["req"].(*http.Request)

		if !ok {
			return
		}

		res, ek := resq.Body["res"].(http.ResponseWriter)

		if !ek {
			return
		}

		// uri := req.URL.Path
		var _ = req

		data := resq.Obj().([]interface{})
		bu := make([]byte, len(data))

		for k, v := range data {
			bu[k] = v.(byte)
		}

		res.Write(bu)
	})
	return r
}

func ErrorRender(response func(rw http.ResponseWriter, r *http.Request)) *Render {
	r := NewRender("ErrorRender", func(i interface{}, t interface{}) {
		_, erc := t.(*Render)

		if !erc {
			return
		}

		resq, err := i.(*grids.GridPacket)

		if !err {
			// core.OutSend("err", resq)
			return
		}

		req, ok := resq.Body["req"].(*http.Request)

		if !ok {
			return
		}

		res, ek := resq.Body["res"].(http.ResponseWriter)

		if !ek {
			return
		}

		response(res, req)
	})

	return r
}

func NewHttp() *HttpServ {
	sv := &HttpServ{grids.NewGrid("ServIO")}
	return sv
}

func NewStatic(base string) *StaticServ {
	control, err := fsgrid.CreateFSControl(base)

	if err != nil {
		return nil
	}

	reader := fsgrid.CreateFSFile()

	st := &StaticServ{grids.NewGrid("web.StaticServer"), base, control, reader}
	st.NewIn("req")
	st.NewOut("res")
	st.NewOut("err")

	reqs := st.In("req")
	rd := reader.Out("res")

	rd.Or(func(i interface{}) {
		resq, err := i.(*grids.GridPacket)

		if !err {
			return
		}
		st.OutSend("res", resq)
	})

	errorHandler := func(i interface{}) {
		resq, err := i.(*grids.GridPacket)
		if !err {
			return
		}
		st.OutSend("err", resq)
	}

	cerror := control.Out("err")
	cerror.Or(errorHandler)
	rerror := reader.Out("err")
	rerror.Or(errorHandler)

	cres := control.Out("res")
	cres.Or(func(i interface{}) {
		resq, err := i.(*grids.GridPacket)

		if !err {
			return
		}

		reader.InSend("file", resq)
	})

	reqs.End(func(packet interface{}, next func(i interface{})) {
		resq, err := packet.(*grids.GridPacket)

		if !err {
			return
		}

		req, ok := resq.Body["req"].(*http.Request)

		if !ok {
			return
		}

		res, ek := resq.Body["res"].(http.ResponseWriter)

		if !ek {
			return
		}

		path := req.URL.Path

		if path == "/" || path == "./" {
			st.OutSend("err", resq)
			return
		}

		pkmap := grids.GridMap{"res": res, "req": req, "file": path}
		pk := grids.CreateGridPacket(pkmap)

		control.InSend("file", pk)

		next(nil)
	})

	return st
}

func NewStaticServo(base string) *StaticServ {
	st := NewStatic(base)

	startTemp := `
		<!doctype>
		<html>
			<head>
				<style>
					*{
						box-sizing: border-box
						-moz-box-sizing: border-box
						-o-box-sizing: border-box
						-webkit-box-sizing: border-box
					}

					body {
						width: 100%;
						height: 100%;
						margin: 5;
						padding: 5;
						background: #ccc;
					}

					h1{
						font-family: helvetica, sans-serif, serif;
						padding: 10%;
						font-size: 4.0em;
						color: #222
					}

					h1 span {
						color: #9e2b0c
					}
				</style>
			</head>
			<body>
	`

	endTemp := `	
			</body>
		</html>
	`

	render := FileRender()
	errorRender := ErrorRender(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html;charset=UTF8")
		res.WriteHeader(http.StatusForbidden)
		res.Write([]byte(startTemp + "<h1> 304! ;(  Sorry! <br/> <span>" + req.URL.Path + " Not Available :p </span></h1>" + endTemp))
	})

	serr := st.Out("err")
	serr.Or(func(i interface{}) {
		resq, err := i.(*grids.GridPacket)

		if !err {
			return
		}

		errorRender.InSend("req", resq)
	})

	sres := st.Out("res")
	sres.Or(func(i interface{}) {
		resq, err := i.(*grids.GridPacket)

		if !err {
			return
		}

		render.InSend("req", resq)
	})

	return st
}

func (s *StaticServ) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	vals := map[interface{}]interface{}{"res": rw, "req": r}
	s.InSend("req", grids.CreateGridPacket(vals))
}

func (h *HttpServ) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	vals := map[interface{}]interface{}{"res": rw, "req": r}
	h.OutSend("req", grids.CreateGridPacket(vals))
}
