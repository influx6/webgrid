package webgrid

import (
  "testing"
  "net/http"
  "log"
  "github.com/influx6/grids"
)

func MakeRequest(t *testing.T){
  res,err := http.Get("http://127.0.0.1:3000/assets")
  defer res.Body.Close()

  if err != nil {
    t.Fatalf("Couldnt get response from request",err,res)
  }

  if err == nil {
    panic("existing...")
  }
}

func TestServer(t *testing.T){

	var _ interface{}

	app := NewHttp()
	_ = NewHttpConsole(app)

	assets := NewRoute("/assets", false)
  reply := NewReply(func (res http.ResponseWriter,req *http.Request,p *grids.GridPacket){
    // res.WriteStatus(200)
    res.Write([]byte("Welcome!"))
  })

	app.OrOut("res", func (g *grids.GridPacket){
     if g == nil {
       t.Fatalf("recieved no packet")
     }
  })

  app.OutBind("res",assets.In("req"))
  assets.OutBind("yes",reply.In("req"))

	err := http.ListenAndServe("127.0.0.1:3000", app)

	if err != nil {
    log.Println("exploding")
		panic("server exploded")
	}


}
