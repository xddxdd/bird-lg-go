<!DOCTYPE html>
<html lang="en-US">
<head>
<link rel="icon" href="/favicon.ico" type="image/x-icon" />
<meta http-equiv="Content-Type" content="text/html;charset=UTF-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
<meta name="renderer" content="webkit">
<title>{{ html .Title }}</title>
<link rel="stylesheet" href="/static/jsdelivr/npm/bootstrap@4.5.1/dist/css/bootstrap.min.css" integrity="sha256-VoFZSlmyTXsegReQCNmbXrS4hBBUl/cexZvPmPWoJsY=" crossorigin="anonymous">
<meta name="robots" content="noindex, nofollow">
</head>
<body>

<nav class="navbar navbar-expand-lg navbar-light bg-light">
	<a class="navbar-brand" href="{{ .BrandURL }}">{{ .Brand }}</a>
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
				{{ if eq .AllServersURLCustom "all" }}
				<a class="nav-link{{ if .AllServersLinkActive }} active{{ end }}"
					href="/{{ $option }}/{{ .AllServersURL }}/{{ $target }}"> {{ .AllServerTitle }} </a>
				{{ else }}
				<a class="nav-link active"
					href="{{ .AllServersURLCustom }}"> {{ .AllServerTitle }} </a>
				{{ end }}
			</li>
			{{ $length := len .Servers }} 
			{{ range $k, $v := .Servers }}
			<li class="nav-item">
				{{ if gt $length 1 }}
				<a class="nav-link{{ if eq $server $v }} active{{ end }}"
					href="/{{ $option }}/{{ $v }}/{{ $target }}">{{ html (index $.ServersDisplay $k) }}</a>
				{{ else }}
				<a class="nav-link{{ if eq $server $v }} active{{ end }}"
					href="/">{{ html (index $.ServersDisplay $k) }}</a>
				{{ end }}
			</li>
			{{ end }}
		</ul>
		{{ if .IsWhois }}
			{{ $target = .WhoisTarget }}
		{{ end }}
		<form name="goto" class="form-inline" action="javascript:goto();">
			<div class="input-group">
				<select name="action" class="form-control">
					{{ range $k, $v := .Options }}
					<option value="{{ html $k }}"{{ if eq $k $.URLOption }} selected{{end}}>{{ html $v }}</option>
					{{ end }}
				</select>
				<input name="server" class="d-none" value="{{ html $server }}">
				<input name="target" class="form-control" placeholder="Target" aria-label="Target" value="{{ html $target }}">
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

<script src="/static/jsdelivr/npm/jquery@3.5.1/dist/jquery.min.js" integrity="sha256-9/aliU8dGd2tb6OSsuzixeV4y/faTqgFtohetphbbj0=" crossorigin="anonymous"></script>
<script src="/static/jsdelivr/npm/bootstrap@4.5.1/dist/js/bootstrap.min.js" integrity="sha256-0IiaoZCI++9oAAvmCb5Y0r93XkuhvJpRalZLffQXLok=" crossorigin="anonymous"></script>
<script src="/static/sortTable.js"></script>

<script>
function goto() {
	let action = $('[name="action"]').val();
	let server = $('[name="server"]').val();
	let target = $('[name="target"]').val();
	let url = "";

	if (action == "whois") {
		url = "/" + action + "/" + target;
	} else if (action == "summary") {
		url = "/" + action + "/" + server + "/";
	} else {
		url = "/" + action + "/" + server + "/" + target;
	}

	window.location.href = url;
}
</script>
</body>
</html>
