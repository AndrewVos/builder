window.autoScroll = true;
window.downloadedOutputBytes = 0;

$(document).ready(function() {
  update(function() {
    if (location.hash == "") {
      if (window.autoScroll) {
        window.scrollTo(0, document.body.scrollHeight);
      }
    } else {
      updateSelectedLineNumber();
    }
  });
});

$(window).scroll(function() {
   window.autoScroll = $(window).scrollTop() + $(window).height() == $(document).height();
});

$(window).on('hashchange', function() {
  updateSelectedLineNumber();
});

function updateSelectedLineNumber() {
  $(".line.focused").removeClass("focused");
  $(location.hash).addClass("focused");
  window.scrollTo(0, $(location.hash).offset().top);
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
