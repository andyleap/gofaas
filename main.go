// gofaas project main.go
package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/andyleap/gofaas/builder"
	"github.com/andyleap/gofaas/processor"
)

var p *processor.Processor
var code = `
package function

type In struct {
	A int
	B int
}

func Handle(in In) int {
	return in.A + in.B
}`

func main() {
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/api", HandleAPI)

	http.ListenAndServe(":8080", nil)
}

var tpl = `
<html>
<head>
</head>
<body>
<form method="post">
<input type="submit"><br/>
<textarea name="code" style="width: 100%; height:50%;">{{.Code}}</textarea><br/>
</form>
<pre>
{{.Output}}
</pre>
</body>
</html>
`

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	output := ""
	if r.FormValue("code") != "" {
		binary, err := builder.Compile([]byte(r.FormValue("code")))
		if err == nil {
			if p != nil {
				p.Stop()
			}
			ioutil.WriteFile("function.exe", binary, 0777)
			p = processor.New("function.exe", 32)
			err = p.Start()
			if err != nil {
				log.Println(err)
			}
		} else {
			output = err.Error()
		}
		code = r.FormValue("code")
	}
	tplData := struct {
		Code   string
		Output string
	}{code, output}
	template.Must(template.New("index.html").Parse(tpl)).Execute(w, tplData)
}

func HandleAPI(w http.ResponseWriter, r *http.Request) {
	if p == nil {
		return
	}
	in, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	out := p.Process(json.RawMessage(in))
	w.Write(out)
}
