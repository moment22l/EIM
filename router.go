package EIM

import (
	"EIM/wire/pkt"
	"fmt"
	"sync"
)

// FuncTree HandlerFun的树结构
type FuncTree struct {
	nodes map[string]HandlerChain
}

func NewTree() *FuncTree {
	return &FuncTree{nodes: make(map[string]HandlerChain, 10)}
}

// Add 将若干HandlerFun放入某个节点中
func (t *FuncTree) Add(path string, handlers ...HandlerFun) {
	if t.nodes[path] == nil {
		t.nodes[path] = HandlerChain{}
	}

	t.nodes[path] = append(t.nodes[path], handlers...)
}

// Get 获取树中的一条chain
func (t *FuncTree) Get(path string) (HandlerChain, bool) {
	chain, ok := t.nodes[path]
	return chain, ok
}

// Router 路由
type Router struct {
	middlewares []HandlerFun // 中间件
	handlers    *FuncTree    // 注册的监听器列表
	pool        sync.Pool    // 对象池
}

func NewRouter() *Router {
	r := &Router{
		middlewares: make([]HandlerFun, 0),
		handlers:    NewTree(),
	}
	r.pool.New = func() interface{} {
		return BuildContext()
	}
	return r
}

func (r *Router) Serve(packet *pkt.LoginPkt, dispatcher Dispatcher, cache SessionStorage, session Session) error {
	if dispatcher == nil {
		return fmt.Errorf("dispatcher is nil")
	}
	if cache == nil {
		return fmt.Errorf("cache is nil")
	}
	ctx := r.pool.Get().(*ContextImpl)
	ctx.reset()
	ctx.request = packet
	ctx.Dispatcher = dispatcher
	ctx.SessionStorage = cache
	ctx.session = session
	r.serveContext(ctx)
	// ctx使用后放回对象池中
	r.pool.Put(ctx)
	return nil
}

func (r *Router) serveContext(ctx *ContextImpl) {
	chain, ok := r.handlers.Get(ctx.Header().Command)
	if !ok {
		ctx.handlers = []HandlerFun{handleNoFound}
		ctx.Next()
		return
	}
	ctx.handlers = chain
	ctx.Next()
}

func handleNoFound(ctx Context) {
	_ = ctx.Resp(pkt.Status_NotImplemented, &pkt.ErrorResp{Message: "NotImplemented"})
}
