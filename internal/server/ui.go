package server

import (
	"html/template"
	"io"
)

var pageTmpl = template.Must(template.New("index").Parse(`<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>landrop</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif; margin: 24px; max-width: 760px; }
    textarea { width: 100%; min-height: 160px; }
    .card { border: 1px solid #ddd; border-radius: 8px; padding: 16px; margin: 16px 0; }
    button { margin-top: 8px; margin-right: 8px; }
    #status { white-space: pre-wrap; background: #f8f8f8; padding: 12px; border-radius: 8px; }
  </style>
</head>
<body>
  <h1>landrop</h1>
  <p>Upload files or send text to this machine.</p>
  <div class="card">
    <h2>Files</h2>
    <form id="uploadForm" action="/upload{{if .Token}}?t={{.Token}}{{end}}" method="post" enctype="multipart/form-data">
      <input type="file" name="file" multiple />
      <br />
      <button type="submit">Upload</button>
    </form>
  </div>
  <div class="card">
    <h2>Text</h2>
    <form id="textForm" action="/text{{if .Token}}?t={{.Token}}{{end}}" method="post">
      <textarea name="text" placeholder="Type or paste text"></textarea>
      <br />
      <button type="submit">Save Text</button>
    </form>
    {{if .ClipboardEnabled}}
      {{if .ClipboardReady}}<p>Clipboard copy is enabled on server.</p>{{else}}<p>Clipboard not supported on this platform.</p>{{end}}
    {{end}}
  </div>
  <h3>Status</h3>
  <div id="status">Ready.</div>
  <script>
    async function submitForm(form, enctype, clearTextOnSuccess) {
      const status = document.getElementById('status');
      try {
        const body = enctype === 'multipart/form-data' ? new FormData(form) : new URLSearchParams(new FormData(form));
        const res = await fetch(form.action, { method: 'POST', body });
        const text = await res.text();
        status.textContent = text;
        if (clearTextOnSuccess && res.ok) {
          const textArea = form.querySelector('textarea[name="text"]');
          if (textArea) {
            textArea.value = '';
          }
        }
      } catch (e) {
        status.textContent = 'request failed: ' + e;
      }
    }
    document.getElementById('uploadForm').addEventListener('submit', (e) => {
      e.preventDefault();
      submitForm(e.target, 'multipart/form-data', false);
    });
    document.getElementById('textForm').addEventListener('submit', (e) => {
      e.preventDefault();
      submitForm(e.target, 'application/x-www-form-urlencoded', true);
    });
  </script>
</body>
</html>`))

type uiData struct {
	Token            string
	ClipboardEnabled bool
	ClipboardReady   bool
}

func renderIndex(w io.Writer, data uiData) error {
	return pageTmpl.Execute(w, data)
}
