window.autoScroll = true;
window.downloadedOutputBytes = 0;
window.scrolledToHash = false;

$(document).ready(function() {
  update();
});

$(document).on("click", ".line", function() {
  selectLine($(this));
});

$(window).scroll(function() {
   if (window.autoScroll = $(window).scrollTop() + $(window).height() == $(document).height()) {
     return true;
   } else {
     return false;
   }
});

function selectLine(element) {
  hash = "#line" + element.index();
  location.hash = hash;
  $(".line.focused").removeClass("focused");
  element.addClass("focused");
  window.scrollTo(0, element.offset().top);
}

function update() {
  $.getJSON("/build_output_raw?id=" + $("#build_id").val() + "&start=" + window.downloadedOutputBytes, function(data) {
    if (data.output != "") {
      window.downloadedOutputBytes += data.length;
      $("#output").append($(data.output));
      if (window.scrolledToHash == false && location.hash != "") {
        window.scrolledToHash = true;
        index = location.hash.replace("#line", "");
        console.log(index);
        element = $(".line").get(index);
        selectLine($(element));
      } else {
        if (window.autoScroll) {
          window.scrollTo(0, document.body.scrollHeight);
        }
      }
    }
    setTimeout(update, 1000);
  });
}
