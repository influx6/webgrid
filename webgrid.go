/*

	Web.Grid is the collection of http based grid blocks for the masses

*/

package webgrid

import (
	"fmt"
	"github.com/influx6/fsgrid"
	"github.com/influx6/grids"
	"github.com/influx6/reggy"
	"mime"
	"net/http"
	fpath "path/filepath"
	"strings"
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

type ActionMethod func(res http.ResponseWriter, req *http.Request, p *grids.GridPacket)

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

type View struct {
	*grids.Grid
	Render *Render
}

type FileView struct {
	*View
	File *fsgrid.FileReader
}

type Render struct {
	Handler func(i interface{}, r interface{})
	*grids.Grid
}

type MethodFilter struct {
	method string
	*grids.Grid
}

type Reply struct {
	RequestCallback ActionMethod
	*grids.Grid
}

type Route struct {
	*grids.Grid
	Route *reggy.ClassicMatchMux
}

func UseHttpPacket(resq *grids.GridPacket, caller func(rw http.ResponseWriter, r *http.Request)) {
	req, ok := resq.Body["req"].(*http.Request)

	if !ok {
		return
	}

	res, ek := resq.Body["res"].(http.ResponseWriter)

	if !ek {
		return
	}

	caller(res, req)
}

func NewRender(format string, m func(i interface{}, f interface{})) *Render {
	r := &Render{m, grids.NewGrid(format)}

	r.NewIn("req")
	r.NewOut("err")

	reqChan := r.In("req")

	reqChan.Receive(func(i interface{}) {
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

		var ctype string

		uri := req.URL.Path

		if file, ok := resq.Body["file"].(string); !ok {
			ctype = mime.TypeByExtension(fpath.Ext(uri))
		} else {
			ctype = mime.TypeByExtension(fpath.Ext(file))
		}

		data := resq.Obj().([]interface{})
		bu := make([]byte, len(data))

		for k, v := range data {
			bu[k] = v.(byte)
		}

		if ctype == "" {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		} else {
			res.Header().Set("Content-Type", ctype)
		}

		res.Write(bu)
	})
	return r
}

func NewRenderView() *View {
	v := &View{grids.NewGrid("FileView"), FileRender()}
	v.NewIn("req")
	v.NewOut("err")
	return v
}

func NewFileView(file string) *FileView {
	v := &FileView{NewRenderView(), fsgrid.ReadFile(file)}

	v.InBind("req", v.File.In("file"))
	v.File.OutBind("res", v.Render.In("req"))
	v.File.OutBind("err", v.Out("err"))

	return v
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

func StripPath(stripBase string) *grids.Grid {
	r := grids.NewGrid("web.StripPath")
	r.NewIn("req")
	r.NewOut("res")

	ssb := strings.TrimPrefix(stripBase, "/")
	reqs := r.In("req")

	reqs.Decide(func(packet interface{}, next func(i interface{})) {
		resq, err := packet.(*grids.GridPacket)
		if !err {
			return
		}

		req, ok := resq.Body["req"].(*http.Request)

		if !ok {
			return
		}

		_, ek := resq.Body["res"].(http.ResponseWriter)

		if !ek {
			return
		}

		path := strings.TrimPrefix(fpath.Clean(req.URL.Path), "/")
		hs := strings.HasPrefix(path, stripBase)
		if hs {
			path = strings.Replace(path, stripBase, "", 1)
		}

		hs = strings.HasPrefix(path, ssb)

		if hs {
			path = strings.Replace(path, ssb, "", 1)
		}

		path = fpath.Clean(path)

		resq.Body["file"] = path

		r.OutSend("res", resq)

		next(nil)
	})

	return r
}

func NewHttp() *HttpServ {
	sv := &HttpServ{grids.NewGrid("ServIO")}
	sv.NewIn("req")
	sv.NewOut("res")
	sv.NewOut("err")
	sv.InBind("req", sv.Out("res"))
	return sv
}

func NewStaticBase(base string, stripbase string) *StaticServ {
	control, err := fsgrid.CreateFSControl(base)

	if err != nil {
		return nil
	}

	reader := fsgrid.CreateFSFile()
	stripper := StripPath(stripbase)

	st := &StaticServ{grids.NewGrid("web.StaticServer"), base, control, reader}
	st.NewIn("req")
	st.NewOut("res")
	st.NewOut("err")

	rd := reader.Out("res")

	rd.Receive(func(i interface{}) {
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
	cerror.Receive(errorHandler)
	rerror := reader.Out("err")
	rerror.Receive(errorHandler)

	cres := control.Out("res")
	cres.Receive(func(i interface{}) {
		resq, err := i.(*grids.GridPacket)

		if !err {
			return
		}

		reader.InSend("read", resq)
	})

	st.InBind("req", stripper.In("req"))

	stripper.Out("res").Decide(func(packet interface{}, next func(i interface{})) {
		resq, err := packet.(*grids.GridPacket)

		if !err {
			return
		}

		path, ok := resq.Body["file"]

		if !ok {
			return
		}

		if path == "/" || path == "./" {
			st.OutSend("err", resq)
			next(nil)
			return
		}

		control.InSend("file", resq)

		next(nil)
	})

	return st
}

func NewStatic(base string, strip string) *StaticServ {
	m := NewStaticBase(base, strip)
	r := FileRender()

	m.OutBind("res", r.In("req"))
	return m
}

func NewStaticServo(base string, strip string) *StaticServ {
	st := NewStaticBase(base, strip)

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
	serr.Receive(func(i interface{}) {
		resq, err := i.(*grids.GridPacket)

		if !err {
			return
		}

		errorRender.InSend("req", resq)
	})

	sres := st.Out("res")
	sres.Receive(func(i interface{}) {
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
	h.InSend("req", grids.CreateGridPacket(vals))
}

func NewMethodFilter(method string) *MethodFilter {
	m := &MethodFilter{method, grids.NewGrid("MethodFilter")}

	m.NewIn("req")
	m.NewOut("res")
	m.NewOut("rej")

	m.AndIn("req", func(p *grids.GridPacket, next func(f interface{})) {

		UseHttpPacket(p, func(res http.ResponseWriter, req *http.Request) {
			med := strings.ToLower(req.Method)

			if method == med {
				m.OutSend("res", p)
				return
			}
			m.OutSend("res", p)
			next(p)
			return
		})

	})

	return m
}

func NewReply(action ActionMethod) *Reply {
	m := &Reply{action, grids.NewGrid("Reply.Action")}
	m.NewIn("req")

	m.AndIn("req", func(p *grids.GridPacket, next func(f interface{})) {
		UseHttpPacket(p, func(res http.ResponseWriter, req *http.Request) {
			action(res, req, p)
			next(p)
		})
	})

	return m
}

func NewRoute(pattern string, strict bool) *Route {
	m := &Route{grids.NewGrid("Route.Action"), reggy.CreateClassic(pattern)}
	m.NewIn("req")

	m.NewOut("no")
	m.NewOut("yes")

	m.AndIn("req", func(p *grids.GridPacket, next func(f interface{})) {
		UseHttpPacket(p, func(res http.ResponseWriter, req *http.Request) {
			path := req.URL.Path

			passed, params := m.Route.Validate(path, strict)

			if !passed {
				m.OutSend("no", p)
				return
			}

			p.Body["params"] = params
			m.OutSend("yes", p)
			next(p)
		})
	})

	return m
}

func NewConsoler() *grids.Grid {
	g := grids.NewGrid("web.Consoler")
	g.NewIn("in")

	g.OrIn("in", func(g *grids.GridPacket) {
		UseHttpPacket(g, func(_ http.ResponseWriter, req *http.Request) {
			fmt.Println("New Request:", req.URL.Path, "Method:", req.Method)
		})
	})

	return g
}

func NewHttpConsole(s *HttpServ) *grids.Grid {
	c := NewConsoler()
	s.InBind("req", c.In("in"))
	return c
}

func NewStaticConsole(s *StaticServ) *grids.Grid {
	c := NewConsoler()
	s.InBind("req", c.In("in"))
	s.InBind("err", c.In("in"))
	return c
}
