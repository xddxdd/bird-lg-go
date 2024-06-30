<h2>BGPmap: {{ html .Target }}</h2>
<div id="bgpmap">
</div>

<script src="/static/jsdelivr/npm/viz.js@2.1.2/viz.min.js" crossorigin="anonymous"></script>
<script src="/static/jsdelivr/npm/viz.js@2.1.2/lite.render.js" crossorigin="anonymous"></script>
<script>
  function decodeBase64(base64) {
    const text = atob(base64);
    const length = text.length;
    const bytes = new Uint8Array(length);
    for (let i = 0; i < length; i++) {
        bytes[i] = text.charCodeAt(i);
    }
    const decoder = new TextDecoder();
    return decoder.decode(bytes);
  }

  var viz = new Viz();
  viz.renderSVGElement(decodeBase64({{ .Result }}))
  .then(element => {
    document.getElementById("bgpmap").appendChild(element);
  })
  .catch(error => {
    document.getElementById("bgpmap").innerHTML = "<pre>"+error+"</pre>"
  });
</script>
