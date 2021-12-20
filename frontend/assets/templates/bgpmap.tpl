<h2>BGPmap: {{ html .Target }}</h2>
<div id="bgpmap">
</div>

<script src="/static/jsdelivr/npm/viz.js@2.1.2/viz.min.js" crossorigin="anonymous"></script>
<script src="/static/jsdelivr/npm/viz.js@2.1.2/lite.render.js" crossorigin="anonymous"></script>
<script>
  var viz = new Viz();
  viz.renderSVGElement(`{{ .Result }}`)
  .then(element => {
    document.getElementById("bgpmap").appendChild(element);
  })
  .catch(error => {
    document.getElementById("bgpmap").innerHTML = "<pre>"+error+"</pre>"
  });
</script>
