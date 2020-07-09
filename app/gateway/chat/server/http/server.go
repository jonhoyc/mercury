package http

import (
	"net/http"
	"outgoing/app/gateway/chat/config"
	"outgoing/app/gateway/chat/session"
	"outgoing/app/gateway/chat/stats"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/ginx"
	"outgoing/x/log"
	"outgoing/x/secretboxer"
	"strings"

	"outgoing/x/websocket"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
)

type httpServer struct {
	id string
	e  *gin.Engine
	l  log.Logger
}

func Init(c config.Provider) {
	opts := []web.Option{
		web.Id(c.ID()),
		web.Name(c.Name()),
		web.Version(c.Version()),
		web.RegisterTTL(c.RegisterTTL()),
		web.RegisterInterval(c.RegisterInterval()),
		web.Address(c.Address()),
	}

	if c.Etcd().Enable {
		etcdv3Registry := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range c.Etcd().Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, web.Registry(etcdv3Registry))
	}

	microWeb := web.NewService(opts...)

	// Initialize service
	if err := microWeb.Init(); err != nil {
		panic("unable to initialize service:" + err.Error())
	}

	s := &httpServer{
		id: microWeb.Options().Id,
		e:  gin.New(),
		l:  c.Logger(),
	}
	s.middleware()
	s.setupRouter()

	microWeb.Handle("/", s)
	stats.Init(microWeb, "/debug/vars")

	// Run service
	go func() {
		if err := microWeb.Run(); err != nil {
			panic("unable to run http service:" + err.Error())
		}
	}()
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.e.ServeHTTP(w, req)
}

func (s *httpServer) middleware() {
	s.e.NoMethod(ginx.NoMethodHandler())
	s.e.NoRoute(ginx.NoRouteHandler())

	s.e.Use(ginx.Recovery(), ginx.Logger(), ginx.CORS())
}

func (s *httpServer) setupRouter() {
	v1 := ginx.NewGroup(s.e.Group("/chat/v1/"))
	{
		v1.GET("/decrypt", s.decryptPath)
		// Websocket
		v1.GET("/channels", s.serveWebSocket)
	}
}

// TODO
func (s *httpServer) decryptPath(c *ginx.Context) {
	encryptPath := c.Query("hash")
	if encryptPath == "" {
		c.Error(ecode.ErrBadRequest)
		return
	}
	timestamp := c.GetHeader(ginx.TimestampHeader)

	boxer, _ := secretboxer.SecretBoxerForType(secretboxer.WrapTypePassphrase, timestamp, secretboxer.EncodingTypeURL)
	realPath, err := boxer.Open(encryptPath)
	if err != nil {
		c.Error(ecode.ErrBadRequest)
		return
	}

	c.Request.URL.Path = string(realPath)
	s.e.HandleContext(c.Context)
}

func (s *httpServer) serveWebSocket(c *ginx.Context) {
	conn, err := websocket.Upgrade(c.Writer, c.Request)
	if err != nil {
		c.Error(err)
		return
	}

	session.GlobalSessionStore.NewSession(c, conn, s.id)
}