package templates

import (
	"html/template"
	"net/http"
)

const (
	ServerVersionString string = "interchange/0.1.0"
	errorTemplate       string = `<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error: {{ .Code }} {{ .Text }}</title>
</head>

<body>
    <div style="text-align: center; font-family: Arial, sans-serif; padding: 50px;">
        <h1 style="font-size: 72px; color: #ff6b6b;">{{ .Code }}</h1>
        <p style="font-size: 24px; color: #333;">{{ .Text }}</p>
        <p style="font-size: 18px; color: #666;">Server: {{ .Server }}</p>
    </div>
</body>

</html>`
)

// parameters to be passed to the error template
type errorParams struct {
	Code   int
	Text   string
	Server string
}

// writes the error template to the given http.ResponseWriter
func WriteError(w http.ResponseWriter, code int, text string) {
	params := errorParams{
		Code:   code,
		Text:   text,
		Server: ServerVersionString,
	}

	tmpl := template.Must(template.New("error").Parse(errorTemplate))
	tmpl.Execute(w, params)
}
