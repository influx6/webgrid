# Webgrid
  http go plugs for Grids[1] framework. It provides the necessary struts for the different necessary http and web building structures and its constantly evolving implementing as best as possible the best practices

##Installation

      go get github.com/influx6/webgrid

  Then
  
      go install github.com/influx6/webgrid

##API
    The webgrid is the collection of Grids[1] structs each providing a part of the functionality for webserver

    -   HttpServ
        This struct creates the grid object for piping http request as a standard fbp component, it provides a ServeHTTP functon for binding into the server handle. It does not create the server itself so as to allow flexibility but its the root starting point for most webgrid grid structures

    -   StaticServ
        This structs provides a static file server that can be used as an asset/file serving struct or as a standard static server for a directory,it provides two helper functions that declare a fully ready file server or static server and it composes the HttpServ

    -   MethodFilter
        This struct encapsulates the operation of filtering http.Request based on the requests method,useful for routing needs

    -   Reply
        This struct encapsulates a callback like pattern for grids, where during creation it accepts a function which is invodked on every GridPacket which contains a http.Request and http.Response 

    -   Route
        This struct encapsulates the routing operations by accepting a specific route and allowing only http.requests which match that route to be allowed or reject them out through its rejection channel. It allows chaining of different routes together using this struct ,where each struct connects to the previous rejection channel for different routes or the previous response channel as the next link from the previous route.

    -   HttpConsoler
        This is a simple grids.Grids struct which uses `fmt` to print to the console basic information of the current incoming request from the server

    -   Render
        This struct is the base for most struct that encapsulates a response action in the webgrid. It requires the passing of a custom function to be called as the rendering action for responses

    -   FileRender
        This struct encapsulates the rendering of a file path as a standard http.response based on the request received,usually when a request wrapped GridPacket comes into its request channel. It has a file reader from the FSGrid[2] package for reading the filepath givien and return it as its result through its response channel

    -   View
        This is the base struct that defines the properties of a view,it is composed the webgrid.FileView and has a render struct where it pipes its incoming packets to be rendered

    
##Function Helpers
    These helpers provided the necessary struct creation of the different different grid structures for this package

    -   UseHttpPacket(packet *GridPacket,callback func(rw http.ResponseWriter,req *http.Request))
        This is a helper function to be attach to grid channels to easily gain access and ensure the gridpackets comes with the necessary http.ResponseWriter and http.Request
    
        ```

                m := &MethodFilter{method, grids.NewGrid("MethodFilter")}

                m.NewIn("req")
                m.NewOut("res")
                m.NewOut("rej")

                m.AndIn("req", func(p *grids.GridPacket, next func(f *grids.GridPacket)) {

                    UseHttpPacket(p, func(res http.ResponseWriter, req *http.Request) {
                        //do something
                    })

                })
        
        ```

    -   NewHttpConsole(s *HttpServ) *grids.Grid
        This returns a new grid struct which prints out http request from the HttpServ struct passed to it

        ```
            app := webgrid.NewHttp()
            console = webgrid.NewHttpConsole(app)

            err := http.ListenAndServe(":3000", app)

            if err != nil {
                panic("server exploded")
            }
        ```

    -   NewStaticConsole(s *StaticServ)  *grids.Grid
        This returns a new consoler struct for printing from a StaticServ struct when used as a static server.

        ```
            app := webgrid.NewStaticServo("..", "")
            console = webgrid.NewStaticConsole(app)

            err := http.ListenAndServe(":3000", app)

            if err != nil {
                panic("server exploded")
            }

        ```

    -   NewRoute(pattern string,strict bool) *Route
        This returns a new route grid struct ready to filter out request packets coming in

        ```
            home := NewRoute(`/app/{id:[/d+]}/log`,false)
        ```

    -   NewMethodFilter(method string) *MethodFilter
        This returns a new methodfilter grid struct ready to filter out request packets coming in based on the request method

        ```
            gets := NewMethodFIlter(“get”)
        ```

    -   NewFileView(file_path string) *View
        This returns a new view grid struct ready to render the file_path content as the response to the request

        ```
            app := webgrid.NewHttp()

            index := webgrid.NewRoute("/", true)
            home := webgrid.NewFileView("./index.html")

            index.OutBind("yes", home.In("req"))

            app.OutBind("res", index.In("req"))
        ```

    -   FileRender() *Render
        This returns a renderer that renders the data within the packets buffer as a file response for the http.request

        ```
            render := FileRender()
            render.OutSend(“res”,...)
        ```

    -   NewReply(action func(res http.ResponseWriter, req *http.Request, p *grids.GridPacket)
        This returns a reply struct which takes a callback for execution on every http gridpacket

        ```
            hello := NewReply(func (res http.Responsewriter,req *http.Request,packet grids.GridPacket){
                res.WriteStatus(200)
                res.Write([]byte(“Hello!”))
                res.Close()
            })
        ```

    -   NewRender(format string,callback func(...)) *Render
        This returns a basic render struct with the channels added and set to executive on each packet comming in

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

                req, ok := resq.Get("req").(*http.Request)

                if !ok {
                    return
                }
            }  


##Example
    Below is a simple server built using webgrid:

    ```
        var _ interface{}

        app := webgrid.NewHttp()
         = webgrid.NewHttpConsole(app)

        index := webgrid.NewRoute("/", true)
        home := webgrid.NewFileView("./index.html")

        index.OutBind("yes", home.In("req"))

        assets := webgrid.NewRoute("/assets", false)
        static := webgrid.NewStatic(".", "/assets/")

        assets.OutBind("yes", static.In("req"))

        io := webgrid.NewRoute("/io", false)
        io.OrOut("yes", func(p *grids.GridPacket) {
            webgrid.UseHttpPacket(p, func(res http.ResponseWriter, req *http.Request) {
                unit.Serve(res, req)
            })
        })

        app.OutBind("res", assets.In("req"))
        app.OutBind("res", index.In("req"))
        app.OutBind("res", io.In("req"))

        err := http.ListenAndServe(":3000", app)

        if err != nil {
            panic("server exploded")
        }
    
    ```


[1](https://github.com/influx6/grids)
[2](https://github.com/influx6/fsgrid)
