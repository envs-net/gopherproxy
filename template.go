package gopherproxy

var tpltext = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>{{.Title}}</title>
</head>
<body>
<section>
<pre>
{{range .Lines}} {{if .Link}}({{.Type}}) <a class="{{ .Type }}" href="{{.Link}}">{{.Text}}</a>{{else}}      {{.Text}}{{end}}
{{end}}</pre>
</section>
<script type="text/javascript">
var qry=document.getElementsByClassName('QRY')
var i=qry.length
while (i--) {
  qry[i].addEventListener('click', function(e) {
    e.preventDefault();
    var resp=prompt("Please enter required input: ", "")
    if (resp !== "") window.location = e.target.href + "?" + resp
    return false;
  })
}
</script>
</body>
</html>`
