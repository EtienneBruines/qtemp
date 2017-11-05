package qtemp

import (
	"html/template"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

var (
	TemplateDir = "templates"
	mgr         = templateManager{templates: make(map[string]*template.Template)}
)

type templateManager struct {
	templates map[string]*template.Template
	master    *template.Template
	mu        sync.RWMutex
	handlers  []VariableHandler
}

func (m *templateManager) Master(tpl string) {
	var err error
	m.master, err = template.New(tpl).ParseFiles(filepath.Join(TemplateDir, tpl))
	if err != nil {
		log.Println("Unable to parse master template:", err)
	}
}

func (m *templateManager) Get(tplFiles []string) (*template.Template, error) {
	key := strings.Join(tplFiles, "|")

	m.mu.RLock()
	t, _ := m.templates[key]
	m.mu.RUnlock()
	if t == nil {
		var (
			err   error
			files []string
		)
		for _, tplFile := range tplFiles {
			files = append(files, filepath.Join(TemplateDir, tplFile))
		}

		if m.master == nil {
			mgr.Master("master.html")
		}

		t, err = template.Must(m.master.Clone()).ParseFiles(files...)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse template %q", tplFiles)
		}
		m.mu.Lock()
		m.templates[key] = t
		m.mu.Unlock()
	}

	return t, nil
}

func Master(tpl string) {
	mgr.Master(tpl)
}

func Render(ctx *fasthttp.RequestCtx, data map[string]interface{}, tpl ...string) {
	RenderWithStatus(ctx, data, fasthttp.StatusOK, tpl...)
}

func RenderWithStatus(ctx *fasthttp.RequestCtx, data Variables, status int, tpl ...string) {
	t, err := mgr.Get(tpl)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("text/plain")
		ctx.WriteString("Unable to parse template: " + err.Error())
		return
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	for _, h := range mgr.handlers {
		data = h(ctx, data)
	}

	err = t.Execute(ctx, data)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("text/plain")
		ctx.WriteString("Unable to execute template: " + err.Error())
		return
	}

	ctx.SetStatusCode(status)
	ctx.SetContentType("text/html")
}

type Variables = map[string]interface{}
type VariableHandler = func(ctx *fasthttp.RequestCtx, p Variables) Variables

func RegisterHandler(handler VariableHandler) {
	mgr.handlers = append(mgr.handlers, handler)
}
