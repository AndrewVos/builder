window.autoScroll = true;

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
  $.getJSON("/build_output_raw?id=" + $("#build_id").val(), function(data) {
    data.output = data.output.replace(">", "&gt;");
    data.output = data.output.replace("<", "&lt;");
    var html = ansi_up.ansi_to_html(data.output);
    document.getElementById("output").innerHTML = html;
    if (window.autoScroll) {
      window.scrollTo(0,document.body.scrollHeight);
    }
    setTimeout(update, 1000);
  });
}

