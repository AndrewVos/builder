$(document).ready(function() {
  update()
  setInterval(update, 1000);
});

function update() {
  $.getJSON("/builds", function(data) {
    $.each(data, function(i, build) {
      var container = $("#builds");

      if ($("#"+build.ID).length == 0) {
        var html = "<div id='"+build.ID+"' class='build'>" +
          "<div class='icon'></div>" +
          "<h2>" +
            "<a href='/build_output?id=" + build.ID + "'>" +
              build.Repo + "/" + build.Ref +
            "</a>" +
          "</h2>" +
            "<a href='" + build.GithubURL + "'>View on Github</a>" +
        "</div>";

        container.prepend($(html));
      }
      var buildLine = $("#" + build.ID);
      if (build.Complete == true) {
        buildLine.removeClass("blue");
        if (build.Success == true) {
          buildLine.addClass("green");
        } else {
          buildLine.addClass("red")
        }
      } else {
        buildLine.addClass("blue")
      }
    });
  });
}
