<h2>BGPmap: {{ html .Target }}</h2>

<script src="https://cdn.jsdelivr.net/npm/viz.js@2.1.2/viz.min.js" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/viz.js@2.1.2/lite.render.js" crossorigin="anonymous"></script>
<script>
  var viz = new Viz();
  viz.renderSVGElement(`{{ .Result }}`)
  .then(element => {
    document.body.appendChild(element);
  })
  .catch(error => {
    document.body.innerHTML = "<pre>"+error+"</pre>"
  });
</script>
