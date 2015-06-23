package main

const (
	displayTmpl = `
<html>
<head>
<title>fcs-sql - {{ .Title }}</title>
<!-- <meta http-equiv="refresh" content="10"> -->
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.time.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.selection.min.js"></script>
<script type="text/javascript">
	var sock = null;
	var wsuri = "ws://127.0.0.1:8080/data";
	var data = [
	        { label: "temperature", 
			  data: [],
			  yaxis: 1
		    },
	        { label: "pressure", 
			  data: [],
		      yaxis: 2
		    },
	        { label: "hygrometry", 
			  data: [],
			  yaxis: 3
		    },
	];

	var options = {
		legend: {
			position: "nw",
			noColumns: 2,
			backgroundOpacity: 0.2
		},
		xaxis: {
			mode: "time",
			timezone: "browser",
			timeformat: "%Y/%m/%d\n%H:%M:%S"
		},
		selection: {
			mode: "x"
		},
	};

	var overviewOpts = {
		legend: { show: false},
		series: {
			lines: {
				show: true,
				lineWidth: 1
			},
			shadowSize: 0
		},
		xaxis: {
			ticks: [],
			mode: "time"
		},
		/*
		yaxis: {
			ticks: [],
			min: 0,
			autoscaleMargin: 0.1
		},
		*/
		selection: {
			mode: "x"
		}
	};

	var plotRange = {
		min: 0,
		max: 0,
	};

	var plot = null;
	var overview = null;

	function update_display() {
		plot = $.plot("#placeholder", data, options);
		overview = $.plot("#overview", data, overviewOpts);
		// now connect the two
		$("#placeholder").bind("plotselected", function (event, ranges) {
			plotRange.min = ranges.xaxis.from;
			plotRange.max = ranges.xaxis.to;
			// do the zooming
			$.each(plot.getXAxes(), function(_, axis) {
				var opts = axis.options;
				opts.min = plotRange.min;
				opts.max = plotRange.max;
			});
			plot.setupGrid();
			plot.draw();
			plot.clearSelection();
			// don't fire event on the overview to prevent eternal loop
			overview.setSelection(ranges, true);
		});

		$("#overview").bind("plotselected", function (event, ranges) {
			plotRange.min = ranges.xaxis.from;
			plotRange.max = ranges.xaxis.to;
			plot.setSelection(ranges);
		});
		if (plotRange.min < plotRange.max) {
			console.log("==> zoom... ["+plotRange.min+", "+plotRange.max+"]");
			
			var ranges = {
				xaxis: {
					from: plotRange.min,
					to: plotRange.max
				}
			};
			
			// do the zooming
			$.each(plot.getXAxes(), function(_, axis) {
				var opts = axis.options;
				opts.min = plotRange.min;
				opts.max = plotRange.max;
			});
			
			plot.setSelection(ranges);
			//overview.setSelection(ranges, true);
			
		}
	};

	window.onload = function() {

            console.log("onload");

            sock = new WebSocket(wsuri);

            sock.onopen = function() {
                console.log("connected to " + wsuri);
            }

            sock.onclose = function(e) {
                console.log("connection closed (" + e.code + ")");
            }

             sock.onmessage = function(e) {
				var obj = JSON.parse(e.data);
				data[0]["data"] = obj["temperature"];
				data[1]["data"] = obj["pressure"];
				data[2]["data"] = obj["hygrometry"];
				update_display();
         }
    };


	/*
	// hard-code color indices to prevent them from shifting as
	// labels are turned on/off

	var i = 0;
	$.each(datasets, function(key, val) {
		val.color = i;
		++i;
	});

	// insert checkboxes 
	var choiceContainer = $("#choices");
	$.each(datasets, function(key, val) {
		choiceContainer.append("<br/><input type='checkbox' name='" + key +
			"' checked='checked' id='id" + key + "'></input>" +
			"<label for='id" + key + "'>"
			+ val.label + "</label>");
	});

	choiceContainer.find("input").click(plotAccordingToChoices);

		function plotAccordingToChoices() {

			var data = [];

			choiceContainer.find("input:checked").each(function () {
				var key = $(this).attr("name");
				if (key && datasets[key]) {
					data.push(datasets[key]);
				}
			});

			if (data.length > 0) {
				$.plot("#placeholder", data, {
					yaxis: {
						min: 0
					},
					xaxis: {
						tickDecimals: 0
					}
				});
			}
		}

		plotAccordingToChoices();

		// Add the Flot version string to the footer

		$("#footer").prepend("Flot " + $.plot.version + " &ndash; ");
	*/

	$(document).ready(update_display);

</script>
<style>
#content {
	margin: 0 auto;
	padding: 10px;
}
.demo-container {
	box-sizing: border-box;
	width: 1200px;
	height: 450px;
	padding: 20px 15px 15px 15px;
	margin: 15px auto 30px auto;
	border: 1px solid #ddd;
	background: #fff;
	background: linear-gradient(#f6f6f6 0, #fff 50px);
	background: -o-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -ms-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -moz-linear-gradient(#f6f6f6 0, #fff 50px);
	background: -webkit-linear-gradient(#f6f6f6 0, #fff 50px);
	box-shadow: 0 3px 10px rgba(0,0,0,0.15);
	-o-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-ms-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-moz-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
	-webkit-box-shadow: 0 3px 10px rgba(0,0,0,0.1);
}
.demo-placeholder {
	width: 100%;
	height: 100%;
	font-size: 14px;
	line-height: 1.2em;
}
</style>
</head>
<body>
<pre>{{ .Title }}</pre>
<div id="content">
	<div class="demo-container">
		<div id="placeholder" class="demo-placeholder"></div>
	</div>
	<div class="demo-container" style="height:150px;">
		<div id="overview" class="demo-placeholder"></div>
	</div>
	<p>The smaller plot is linked to the main plot, so it acts as an overview. Try dragging a selection on either plot, and watch the behavior of the other.</p>
</div>
<pre><b>Legend</b>
Temperature: in Celsius
Pressure: in mbar
Hygrometry: in per-cents
</pre>
</body>
</html>
	`
)
