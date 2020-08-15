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

	URLProto   string
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
<meta name="viewport" content="width=device-width,initial-scale=1,shrink-to-fit=no">
<meta name="renderer" content="webkit">
<title>{{ .Title }}</title>
<link href="https://cdn.jsdelivr.net/npm/bootstrap@4.4.1/dist/css/bootstrap.min.css" rel="stylesheet">
<meta name="robots" content="noindex, nofollow">
</head>
<body>

<nav class="navbar navbar-expand-lg navbar-light bg-light">
	<a class="navbar-brand" href="/">{{ .Brand }}</a>
	<button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
		<span class="navbar-toggler-icon"></span>
	</button>

	<div class="collapse navbar-collapse" id="navbarSupportedContent">
		<ul class="navbar-nav mr-auto">
			<li class="nav-item"><a class="nav-link{{ if eq "ipv4" .URLProto }} active{{ end }}" href="/ipv4/{{ .URLOption }}/{{ .URLServer }}/{{ .URLCommand }}"> IPv4 </a></li>
			<li class="nav-item"><a class="nav-link{{ if eq "ipv6" .URLProto }} active{{ end }}" href="/ipv6/{{ .URLOption }}/{{ .URLServer }}/{{ .URLCommand }}"> IPv6 </a></li>
			<span class="navbar-text">|</span>
			<li class="nav-item">
				<a class="nav-link{{ if .AllServersLinkActive }} active{{ end }}" href="/{{ .URLProto }}/{{ .URLOption }}/{{ .AllServersURL }}/{{ .URLCommand }}"> All Servers </a>
			</li>
			{{ range $k, $v := .Servers }}
			<li class="nav-item">
				<a class="nav-link{{ if eq $.URLServer $v }} active{{ end }}" href="/{{ $.URLProto }}/{{ $.URLOption }}/{{ $v }}/{{ $.URLCommand }}">{{ $v }}</a>
			</li>
			{{ end }}
		</ul>
		{{ $option := .URLOption }}
		{{ $target := .URLCommand }}
		{{ if .IsWhois }}
			{{ $option = "whois" }}
			{{ $target = .WhoisTarget }}
		{{ end }}
		<form class="form-inline" action="/redir" method="GET">
			<div class="input-group">
				<select name="action" class="form-control">
					{{ range $k, $v := .Options }}
					<option value="{{ $k }}"{{ if eq $k $option }} selected{{end}}>{{ $v }}</option>
					{{ end }}
				</select>
				<input name="proto" class="d-none" value="{{ .URLProto }}">
				<input name="server" class="d-none" value="{{ .URLServer }}">
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
</body>
</html>
`))
