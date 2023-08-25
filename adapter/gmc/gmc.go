package gmc

import (
	"bytes"
	"errors"
	"github.com/GoAdminGroup/go-admin/adapter"
	gmctx "github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/plugins"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/template/types"
	gcore "github.com/snail007/gmc/core"
	gtemplate "github.com/snail007/gmc/http/template"
	_ "github.com/snail007/gmc/using/web"
	"net/http"
	"net/url"
	"strings"
)

// GMC structure value is a GMC GoAdmin adapter.
type GMC struct {
	adapter.BaseAdapter
	ctx gcore.Ctx
	app gcore.HTTPServer
}

func init() {
	gtemplate.SetBinBytes(map[string][]byte{"": []byte("")})
	engine.Register(new(GMC))
}

// User implements the method Adapter.User.
func (s *GMC) User(ctx interface{}) (models.UserModel, bool) {
	return s.GetUser(ctx, s)
}

// Use implements the method Adapter.Use.
func (s *GMC) Use(app interface{}, plugs []plugins.Plugin) error {
	return s.GetUse(app, plugs, s)
}

// Content implements the method Adapter.Content.
func (s *GMC) Content(ctx interface{}, getPanelFn types.GetPanelFn, fn gmctx.NodeProcessor, navButtons ...types.Button) {
	s.GetContent(ctx, getPanelFn, s, navButtons, fn)
}

type HandlerFunc func(ctx gcore.Ctx) (types.Panel, error)

func Content(handler HandlerFunc) gcore.Middleware {
	return func(ctx gcore.Ctx) bool {
		engine.Content(ctx, func(ctx interface{}) (types.Panel, error) {
			return handler(ctx.(gcore.Ctx))
		})
		return false
	}
}

// SetApp implements the method Adapter.SetApp.
func (s *GMC) SetApp(app interface{}) error {
	var (
		eng gcore.HTTPServer
		ok  bool
	)
	if eng, ok = app.(gcore.HTTPServer); !ok {
		return errors.New("gmc adapter SetApp: wrong parameter")
	}
	s.app = eng
	return nil
}

// AddHandler implements the method Adapter.AddHandler.
func (s *GMC) AddHandler(method, path string, handlers gmctx.Handlers) {
	s.app.Ctx().WebServer().Router().Handle(strings.ToUpper(method), path, func(w http.ResponseWriter, r *http.Request, ps gcore.Params) {
		for _, v := range ps {
			key := v.Key
			value := v.Value
			if r.URL.RawQuery == "" {
				r.URL.RawQuery += strings.ReplaceAll(key, ":", "") + "=" + value
			} else {
				r.URL.RawQuery += "&" + strings.ReplaceAll(key, ":", "") + "=" + value
			}
		}
		ctx := gmctx.NewContext(r)
		ctx.SetHandlers(handlers).Next()
		for key, head := range ctx.Response.Header {
			w.Header().Add(key, head[0])
		}
		w.WriteHeader(ctx.Response.StatusCode)
		if ctx.Response.Body != nil {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(ctx.Response.Body)
			w.Write(buf.Bytes())
		}
	})
}

// Name implements the method Adapter.Name.
func (*GMC) Name() string {
	return "gmc"
}

// SetContext implements the method Adapter.SetContext.
func (*GMC) SetContext(contextInterface interface{}) adapter.WebFrameWork {
	var (
		ctx gcore.Ctx
		ok  bool
	)
	if ctx, ok = contextInterface.(gcore.Ctx); !ok {
		panic("gmc adapter SetContext: wrong parameter")
	}
	return &GMC{ctx: ctx}
}

// Redirect implements the method Adapter.Redirect.
func (s *GMC) Redirect() {
	http.Redirect(s.ctx.Response(), s.ctx.Request(), config.Url(config.GetLoginUrl()), http.StatusFound)
}

// SetContentType implements the method Adapter.SetContentType.
func (s *GMC) SetContentType() {
	s.ctx.Response().Header().Set("Content-Type", s.HTMLContentType())
}

// Write implements the method Adapter.Write.
func (s *GMC) Write(body []byte) {
	_, _ = s.ctx.Response().Write(body)
}

// GetCookie implements the method Adapter.GetCookie.
func (s *GMC) GetCookie() (string, error) {
	return s.ctx.Cookie(s.CookieKey()), nil
}

// Lang implements the method Adapter.Lang.
func (s *GMC) Lang() string {
	return s.ctx.Request().URL.Query().Get("__ga_lang")
}

// Path implements the method Adapter.Path.
func (s *GMC) Path() string {
	return s.ctx.Request().URL.Path
}

// Method implements the method Adapter.Method.
func (s *GMC) Method() string {
	return s.ctx.Request().Method
}

// FormParam implements the method Adapter.FormParam.
func (s *GMC) FormParam() url.Values {
	_ = s.ctx.Request().ParseMultipartForm(32 << 20)
	return s.ctx.Request().PostForm
}

// IsPjax implements the method Adapter.IsPjax.
func (s *GMC) IsPjax() bool {
	return s.ctx.Request().Header.Get(constant.PjaxHeader) == "true"
}

// Query implements the method Adapter.Query.
func (s *GMC) Query() url.Values {
	return s.ctx.Request().URL.Query()
}
