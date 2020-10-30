package main

import (
	"text/template"
)

type tmplArguments struct {
	// Global options
	Options map[string]string
	Servers []string

	// Parameters related to current request
	AllServersLinkActive bool
	AllServersURL        string

	// Whois specific handling (for its unique URL)
	IsWhois     bool
	WhoisTarget string

	URLOption  string
	URLServer  string
	URLCommand string

	// Generated content to be displayed
	Title   string
	Brand   string
	Content string
}

var tmpl = template.Must(template.New("tmpl").Parse(`
<!DOCTYPE html>
<html lang="en-US">
<head>
<meta http-equiv="Content-Type" content="text/html;charset=UTF-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
<meta name="renderer" content="webkit">
<title>{{ .Title }}</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@4.5.1/dist/css/bootstrap.min.css" integrity="sha256-VoFZSlmyTXsegReQCNmbXrS4hBBUl/cexZvPmPWoJsY=" crossorigin="anonymous">
<meta name="robots" content="noindex, nofollow">
</head>
<body>

<nav class="navbar navbar-expand-lg navbar-light bg-light">
	<a class="navbar-brand" href="/">{{ .Brand }}</a>
	<button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
		<span class="navbar-toggler-icon"></span>
	</button>

	<div class="collapse navbar-collapse" id="navbarSupportedContent">
		{{ $option := .URLOption }}
		{{ $server := .URLServer }}
		{{ $target := .URLCommand }}
		{{ if .IsWhois }}
			{{ $option = "summary" }}
			{{ $server = .AllServersURL }}
			{{ $target = "" }}
		{{ end }}
		<ul class="navbar-nav mr-auto">
			<li class="nav-item">
				<a class="nav-link{{ if .AllServersLinkActive }} active{{ end }}"
					href="/{{ $option }}/{{ .AllServersURL }}/{{ $target }}"> All Servers </a>
			</li>
			{{ range $k, $v := .Servers }}
			<li class="nav-item">
				<a class="nav-link{{ if eq $server $v }} active{{ end }}"
					href="/{{ $option }}/{{ $v }}/{{ $target }}">{{ $v }}</a>
			</li>
			{{ end }}
		</ul>
		{{ if .IsWhois }}
			{{ $target = .WhoisTarget }}
		{{ end }}
		<form class="form-inline" action="/redir" method="GET">
			<div class="input-group">
				<select name="action" class="form-control">
					{{ range $k, $v := .Options }}
					<option value="{{ $k }}"{{ if eq $k $.URLOption }} selected{{end}}>{{ $v }}</option>
					{{ end }}
				</select>
				<input name="server" class="d-none" value="{{ $server }}">
				<input name="target" class="form-control" placeholder="Target" aria-label="Target" value="{{ $target }}">
				<div class="input-group-append">
					<button class="btn btn-outline-success" type="submit">&raquo;</button>
				</div>
			</div>
		</form>
	</div>
</nav>

<div class="container">
	{{ .Content }}
</div>

<script src="https://cdn.jsdelivr.net/npm/jquery@3.5.1/dist/jquery.min.js" integrity="sha256-9/aliU8dGd2tb6OSsuzixeV4y/faTqgFtohetphbbj0=" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/bootstrap@4.5.1/dist/js/bootstrap.min.js" integrity="sha256-0IiaoZCI++9oAAvmCb5Y0r93XkuhvJpRalZLffQXLok=" crossorigin="anonymous"></script>
</body>
</html>
`))
