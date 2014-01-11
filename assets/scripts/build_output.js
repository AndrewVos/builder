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
  $.getJSON("/build/" + $("#build_id").val() + "/output/raw/&start=" + window.downloadedOutputBytes, function(data) {
    if (data.output != "") {
      window.downloadedOutputBytes += data.length;
      $("#output").append($(data.output));
      if (window.scrolledToHash == false && location.hash != "") {
        window.scrolledToHash = true;
        index = location.hash.replace("#line", "");
        element = $(".line").get(index);
        selectLine($(element));
      } else {
        if (window.autoScroll) {
          window.scrollTo(0, document.body.scrollHeight);
        }
      }
      updateScroller();
    }
    setTimeout(update, 1000);
  });
}

function updateScroller() {
  var scroller = $(".scroller");
  $("#output .line").each(function() {
    var line = $(this);
    if (line.hasClass("added-to-scroller") == false) {
      line.addClass("added-to-scroller");
      if (line.find("span.red").length > 0) {
        red = $("<div class='scroller_line red'><span></span></div>");
        red.click(function() {
          selectLine(line);
        });
        scroller.append(red);
      } else {
        scroller.append($("<div class='scroller_line'><span></span></div>"));
      }
    }
  });
}
