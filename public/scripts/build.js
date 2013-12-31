$(document).ready(function() {
  update()
  setInterval(update, 1000);
});

function update() {
  $.getJSON("/builds", function(data) {
    $.each(data, function(i, build) {
      var container = $("#builds");

      if ($("#"+build.ID).length == 0) {
        var commits = "";

        if (build.Commits != null && build.Commits.length > 0) {
          for (i = 0; i < build.Commits.length; i++) {
            var commit = build.Commits[i];
            commits += "<div>";
            commits += '<span class="label label-info">' + commit.SHA.slice(0,7) + '</span>';
            commits += "<span> " + commit.Message + "</span>";
            commits += "</div>";
          }
        }

        var html = "<div id='"+build.ID+"' class='build'>" +
          "<h2>" +
            "<div class='ball-container'><div class='ball'></div></div>" +
            "<a href='/build_output?id=" + build.ID + "'>" +
              build.Repo + "/" + build.Ref +
            "</a>" +
          "</h2>" +
            commits +
            "<div><a href='" + build.GithubURL + "'>View on Github</a></div>" +
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
