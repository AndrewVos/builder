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
    var txt  = "\n\n\033[1;33;40m 33;40  \033[1;33;41m 33;41  \033[1;33;42m 33;42  \033[1;33;43m 33;43  \033[1;33;44m 33;44  \033[1;33;45m 33;45  \033[1;33;46m 33;46  \033[1m\033[0\n\n\033[1;33;42m >> Tests OK\n\n"
    var html = ansi_up.ansi_to_html(data.output);
    document.getElementById("output").innerHTML = html;
    if (window.autoScroll) {
      window.scrollTo(0,document.body.scrollHeight);
    }
    setTimeout(update, 1000);
  });
}

