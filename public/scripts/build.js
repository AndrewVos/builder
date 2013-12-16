$(document).ready(function() {
  update()
  setInterval(update, 1000);
});

function update() {
  $.getJSON("/builds", function(data) {
    $.each(data, function(i, build) {
      console.log(build.ID + "-" + build.Complete);
      var container = $("#builds");
      var id = "build_" + build.ID;

      if ($("#"+id).length == 0) {
        var buildLine = $("<div id='"+id+"' class='build'></div>");
        var title = $("<h1>"+ build.Owner + "/" + build.Repo + "</h1>");
        var icon = $("<div class='icon'></div>");
        buildLine.append(icon);
        buildLine.append(title);
        container.prepend(buildLine)
      }
      var buildLine = $("#" + id);
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
