window.autoScroll = true;
window.downloadedOutputBytes = 0;

$(document).ready(function() {
  update(function() {
    if (location.hash == "") {
      if (window.autoScroll) {
        window.scrollTo(0, document.body.scrollHeight);
      }
    } else {
      index = location.hash.replace("#line", "");
      console.log(index);
      element = $(".line").get(index);
      selectLine($(element));
    }
  });
});

$(document).on("click", ".line", function() {
  selectLine($(this));
});

$(window).scroll(function() {
   window.autoScroll = $(window).scrollTop() + $(window).height() == $(document).height();
});

function selectLine(element) {
  hash = "#line" + element.index();
  location.hash = hash;
  $(".line.focused").removeClass("focused");
  element.addClass("focused");
  window.scrollTo(0, element.offset().top);
}

function update(success) {
  $.getJSON("/build_output_raw?id=" + $("#build_id").val() + "&start=" + window.downloadedOutputBytes, function(data) {
    if (data.output != "") {
      window.downloadedOutputBytes += data.length;
      $("#output").append($(data.output));
      if(typeof success == 'function') {
        success();
      }
    }
    setTimeout(update, 1000);
  });
}
