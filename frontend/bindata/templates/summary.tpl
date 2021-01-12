{{ $ServerName := urlquery .ServerName }}

<table class="table table-striped table-bordered table-sm">
  <thead>
{{ range .Header }}
    <th scope="col">{{ html . }}</th>
{{ end }}
  </thead>
  <tbody>
{{ range .Rows }}
    <tr class="table-{{ .MappedState }}">
      <td><a href="/detail/{{ $ServerName }}/{{ urlquery .Name }}">{{ html .Name }}</a></td>
      <td>{{ .Proto }}</td>
      <td>{{ .Table }}</td>
      <td>{{ .State }}</td>
      <td>{{ .Since }}</td>
      <td>{{ .Info  }}</td>
    </tr>
{{ end }}
  </tbody>
</table>

<!-- 
{{ .Raw }}
-->

