package core

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/engine"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/baetyl/baetyl/sync"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
	"github.com/valyala/fasthttp"
)

type Core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn sync.Sync
	svr *http.Server
}

// NewCore creates a new core
func NewCore(cfg config.Config) (*Core, error) {
	c := &Core{}
	var err error
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.sha, err = node.NewNode(c.sto)
	if err != nil {
		c.Close()
		return nil, errors.Trace(err)
	}
	c.syn, err = sync.NewSync(cfg.Sync, c.sto, c.sha)
	if err != nil {
		c.Close()
		return nil, errors.Trace(err)
	}
	c.syn.Start()

	c.eng, err = engine.NewEngine(cfg, c.sto, c.sha, c.syn)
	if err != nil {
		c.Close()
		return nil, errors.Trace(err)
	}
	c.eng.Start()

	c.svr = http.NewServer(cfg.Server, c.initRouter())
	c.svr.Start()
	return c, nil
}

func (c *Core) Close() {
	if c.svr != nil {
		c.svr.Close()
	}
	if c.eng != nil {
		c.eng.Close()
	}
	if c.sto != nil {
		c.sto.Close()
	}
	if c.syn != nil {
		c.syn.Close()
	}
}

func (c *Core) initRouter() fasthttp.RequestHandler {
	router := routing.New()
	router.Get("/node/stats", c.sha.GetStats)
	router.Get("/services/<service>/log", c.eng.GetServiceLog)
	return router.HandleRequest
}
