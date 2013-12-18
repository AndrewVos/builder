window.autoScroll = true;
window.downloadedOutputBytes = 0;

$(window).scroll(function() {
   if($(window).scrollTop() + $(window).height() == $(document).height()) {
     window.autoScroll = true;
   } else {
     window.autoScroll = false;
   }
});

$(document).ready(function() {
  document.scroll
  update();
});

function update() {
  $.getJSON("/build_output_raw?id=" + $("#build_id").val() + "&start=" + window.downloadedOutputBytes, function(data) {
    e = document.getElementById("output");
    window.downloadedOutputBytes += data.output.length;
    e.innerHTML = e.innerHTML + data.output;
    if (window.autoScroll) {
      window.scrollTo(0,document.body.scrollHeight);
    }
    setTimeout(update, 1000);
  });
}
