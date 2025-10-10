package renderer

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"

	"rogchap.com/v8go"
)

// Renderer renders a React application to HTML.
type Renderer struct {
	pool          *IsolatePool
	ssrScriptName string
}

type Result struct {
	HTML string
	Head string
}

// NewRenderer creates a new server side renderer for a given script.
func NewRenderer(scriptContents string) *Renderer {
	ssrScriptName := "server.js"

	return &Renderer{
		pool:          NewIsolatePool(scriptContents, ssrScriptName),
		ssrScriptName: ssrScriptName,
	}
}

// Render renders the provided path to HTML with optional data payload.
func (r *Renderer) Render(urlPath string, payload map[string]any) (Result, error) {
	iso := r.pool.Get()
	defer r.pool.Put(iso)

	ctx := v8go.NewContext(iso.Isolate)
	defer ctx.Close()

	if len(payload) > 0 {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return Result{}, err
		}

		escaped := template.JSEscapeString(string(jsonData))
		script := fmt.Sprintf(`globalThis.__SSR_DATA__ = JSON.parse("%s");`, escaped)
		if _, err := ctx.RunScript(script, "ssr-data.js"); err != nil {
			return Result{}, formatError(err)
		}
	}

	iso.RenderScript.Run(ctx)

	quotedPath := strconv.Quote(urlPath)
	renderCmd := fmt.Sprintf("ssrRender(%s)", quotedPath)
	val, err := ctx.RunScript(renderCmd, r.ssrScriptName)
	if err != nil {
		return Result{}, formatError(err)
	}

	renderedHtml := ""

	if val.IsPromise() {
		result, err := resolvePromise(ctx, val, err)
		if err != nil {
			return Result{}, formatError(err)
		}

		renderedHtml = result.String()
	} else {
		renderedHtml = val.String()
	}

	headVal, err := ctx.RunScript("globalThis.__SSR_HEAD__ || ''", "ssr-head.js")
	if err != nil {
		return Result{}, formatError(err)
	}

	headContent := ""
	if headVal != nil {
		headContent = headVal.String()
	}

	return Result{
		HTML: renderedHtml,
		Head: headContent,
	}, nil
}
