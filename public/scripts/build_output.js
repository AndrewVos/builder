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
  update();
});

function update() {
  $.getJSON("/build_output_raw?id=" + $("#build_id").val() + "&start=" + window.downloadedOutputBytes, function(data) {
    if (data.output != "") {
      window.downloadedOutputBytes += data.length;
      $("#output").append($("<span>"+data.output+"<span>"));
      if (window.autoScroll) {
        window.scrollTo(0, document.body.scrollHeight);
      }
    }
    setTimeout(update, 1000);
  });
}
