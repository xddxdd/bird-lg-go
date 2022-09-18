{{ $ServerName := urlquery .ServerName }}

<table class="table table-striped table-bordered table-sm sortable">
  <thead>
{{ range .Header }}
    <th scope="col">{{ html . }}</th>
{{ end }}
  </thead>
  <tbody>
{{ range .Rows }}
    <tr class="table-{{ .MappedState }}">
      <td><a href="/detail/{{ $ServerName }}/{{ urlquery .Name }}">{{ html .Name }}</a></td>
      <td>{{ html .Proto }}</td>
      <td>{{ html .Table }}</td>
      <td>{{ html .State }}</td>
      <td>{{ html .Since }}</td>
      <td>{{ html .Info  }}</td>
    </tr>
{{ end }}
  </tbody>
</table>
