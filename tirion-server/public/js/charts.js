function createChart(id, series, options) {
	options = options || {
		live: false,
	};

	var pushExtremes = true;

	var cChart = new Highcharts.StockChart({
		chart: {
			events : {
				load : function() {
					var chart = this;

					if (options.live && (! options.isRunning || options.isRunning())) {
						var fetchMetricsDataInterval = setInterval(function() {
							var loaded = 0;
							var needed = 0;

							$.each(chart.series, function(i, serie) {
								if (serie.type != 'line') {
									return;
								}

								needed++;

								$.getJSON(serie.options.url + "&from=" + serie.xData[serie.xData.length - 1], function(data) {
									for (var i = 0; i < data.length; i++) {
										serie.addPoint(data[i], false, false);
									}

									if (++loaded == needed) {
										pushExtremes = false;

										chart.redraw();

										pushExtremes = true;
									}
								});
							});

							if (options.isRunning && ! options.isRunning()) {
								clearInterval(fetchMetricsDataInterval);

								return;
							}
						}, 3000);
					}
				},
			},
			renderTo: id,
			title: series.name,
			zoomType: 'x',
		},

		credits: {
			enabled: false,
		},

		navigator: {
			series: {
				dataGrouping: {
					smoothed: false,
				},
			},
		},

		plotOptions: {
			series: {
				animation: false,
			},
		},

		rangeSelector: {
			buttons: [
				{
					type: 'second',
					count: 1,
					text: '1s'
				}, {
					type: 'second',
					count: 5,
					text: '5s'
				}, {
					type: 'minute',
					count: 1,
					text: '1m'
				}, {
					type: 'minute',
					count: 5,
					text: '5m'
				}, {
					type: 'hour',
					count: 1,
					text: '1h'
				}, {
					type: 'hour',
					count: 5,
					text: '5h'
				}, {
					type: 'all',
					text: 'All'
				}
			],
			inputEnabled: false,
			selected: 6,
		},

		series: series,

		xAxis: {
			minRange: 1, // one ms
		},
	});

	pushChart(cChart);

	var xAxis = cChart.xAxis[0];

	$(window).bind('popstate', function(e) {
		var state = e.originalEvent.state;

		if (xAxis && state && state.id && state.id == id) {
			pushExtremes = false;

			xAxis.setExtremes(state.from, state.to);

			pushExtremes = true;
		}
	});

	$(xAxis).bind('afterSetExtremes', function (e) {
		if (pushExtremes) {
			history.pushState({
				id: id,
				from: e.min,
				to: e.max,
			}, '');
		}
	});

	var extremes = xAxis.getExtremes();

	history.replaceState({
		id: id,
		from: extremes.min,
		to: extremes.max,
	}, '');
}

function createMultiChart(id, series, options) {
	options = options || {
		live: false,
	};

	var pushExtremes = true;

	var multiA = [];

	$.each(series, function(i, serie) {
		var div = document.createElement('div');

		var cChart = new Highcharts.StockChart({
			chart: {
				events : {
					load : function() {
						var chart = this;

						if (options.live && (! options.isRunning || options.isRunning())) {
							var fetchMetricsDataInterval = setInterval(function() {
								var loaded = 0;
								var needed = 0;

								$.each(chart.series, function(i, serie) {
									if (serie.type != 'line') {
										return;
									}

									needed++;

									$.getJSON(serie.options.url + "&from=" + serie.xData[serie.xData.length - 1], function(data) {
										for (var i = 0; i < data.length; i++) {
											serie.addPoint(data[i], false, false);
										}

										if (++loaded == needed) {
											pushExtremes = false;

											chart.redraw();

											pushExtremes = true;
										}
									});
								});

								if (options.isRunning && ! options.isRunning()) {
									clearInterval(fetchMetricsDataInterval);

									return;
								}
							}, 3000);
						}
					},
				},
				renderTo: div,
				zoomType: 'x',
			},

			credits: {
				enabled: false,
			},

			navigator: {
				series: {
					dataGrouping: {
						smoothed: false,
					},
				},
			},

			plotOptions: {
				series: {
					animation: false,
				},
			},

			rangeSelector: (i == 0) ? {
				buttons: [
					{
						type: 'second',
						count: 1,
						text: '1s'
					}, {
						type: 'second',
						count: 5,
						text: '5s'
					}, {
						type: 'minute',
						count: 1,
						text: '1m'
					}, {
						type: 'minute',
						count: 5,
						text: '5m'
					}, {
						type: 'hour',
						count: 1,
						text: '1h'
					}, {
						type: 'hour',
						count: 5,
						text: '5h'
					}, {
						type: 'all',
						text: 'All'
					}
				],
				inputEnabled: false,
				selected: 6,
			} : {
				enabled: false,
			},

			series: [serie],

			title: {
		 		text: serie.name
			},

			xAxis: {
				minRange: 1, // one ms
			},
		});

		multiA.push(cChart);
		pushChart(cChart);

		var xAxis = cChart.xAxis[0];

		$(xAxis).bind('afterSetExtremes', function (e) {
			if (pushExtremes) {
				history.pushState({
					id: id,
					from: e.min,
					to: e.max,
				}, '');

				pushExtremes = false;

				for (var i = 0; i < multiA.length; i++) {
					if (multiA[i] != cChart) {
						multiA[i].xAxis[0].setExtremes(e.min, e.max);
					}
				}

				pushExtremes = true;
			}
		});

		$('#' + id).append(div);
	});

	$(window).bind('popstate', function(e) {
		var state = e.originalEvent.state;

		if (state && state.id && state.id == id) {
			pushExtremes = false;

			for (var i = 0; i < multiA.length; i++) {
				multiA[i].xAxis[0].setExtremes(state.from, state.to);
			}

			pushExtremes = true;
		}
	});

	var extremes = multiA[0].xAxis[0].getExtremes();

	history.replaceState({
		id: id,
		from: extremes.min,
		to: extremes.max,
	}, '');
}

function createCombinedMultiChart(id, series, options) {
	options = options || {
		live: false,
	};

	var div = document.createElement('div');

	var cSeries = [];
	var cYAxis = [];
	var multiA = [];
	var pushExtremes = true;

	var top = 0;

	$.each(series, function(i, serie) {
		cYAxis.push({
			offset: 0,
	        title: {
	            text: serie.name
	        },
			top: top,
	        height: 200
	    });
		cSeries.push(serie);

		top += 200
	});

	var cChart = new Highcharts.StockChart({
		chart: {
			events : {
				load : function() {
					var chart = this;

					if (options.live && (! options.isRunning || options.isRunning())) {
						var fetchMetricsDataInterval = setInterval(function() {
							var loaded = 0;
							var needed = 0;

							$.each(chart.series, function(i, serie) {
								if (serie.type != 'line') {
									return;
								}

								needed++;

								$.getJSON(serie.options.url + "&from=" + serie.xData[serie.xData.length - 1], function(data) {
									for (var i = 0; i < data.length; i++) {
										serie.addPoint(data[i], false, false);
									}

									if (++loaded == needed) {
										pushExtremes = false;

										chart.redraw();

										pushExtremes = true;
									}
								});
							});

							if (options.isRunning && ! options.isRunning()) {
								clearInterval(fetchMetricsDataInterval);

								return;
							}
						}, 3000);
					}
				},
			},
			height: top + 100,
			renderTo: div,
			width: 1000,
			zoomType: 'x',
		},

		credits: {
			enabled: false,
		},

		navigator: {
			series: {
				dataGrouping: {
					smoothed: false,
				},
			},
		},

		plotOptions: {
			series: {
				animation: false,
			},
		},

		rangeSelector: {
			buttons: [
				{
					type: 'second',
					count: 1,
					text: '1s'
				}, {
					type: 'second',
					count: 5,
					text: '5s'
				}, {
					type: 'minute',
					count: 1,
					text: '1m'
				}, {
					type: 'minute',
					count: 5,
					text: '5m'
				}, {
					type: 'hour',
					count: 1,
					text: '1h'
				}, {
					type: 'hour',
					count: 5,
					text: '5h'
				}, {
					type: 'all',
					text: 'All'
				}
			],
			inputEnabled: false,
			selected: 6,
		},

	    series: cSeries,

		xAxis: {
			minRange: 1, // one ms
		},

		yAxis: cYAxis
	});

	multiA.push(cChart);
	pushChart(cChart);

	var xAxis = cChart.xAxis[0];

	$(xAxis).bind('afterSetExtremes', function (e) {
		if (pushExtremes) {
			history.pushState({
				id: id,
				from: e.min,
				to: e.max,
			}, '');

			pushExtremes = false;

			for (var i = 0; i < multiA.length; i++) {
				if (multiA[i] != cChart) {
					multiA[i].xAxis[0].setExtremes(e.min, e.max);
				}
			}

			pushExtremes = true;
		}
	});

	$('#' + id).append(div);

	$(window).bind('popstate', function(e) {
		var state = e.originalEvent.state;

		if (state && state.id && state.id == id) {
			pushExtremes = false;

			for (var i = 0; i < multiA.length; i++) {
				multiA[i].xAxis[0].setExtremes(state.from, state.to);
			}

			pushExtremes = true;
		}
	});

	var extremes = multiA[0].xAxis[0].getExtremes();

	history.replaceState({
		id: id,
		from: extremes.min,
		to: extremes.max,
	}, '');
}

function pushChart(chart) {
	if (window.chart) {
		window.chart.push(chart);
	}
	else {
		window.chart = [chart];
	}
}

function getJSONSynchron(url) {
	var ret;

	$.ajax({
		type: 'GET',
		url: url,
		dataType: 'json',
		success: function(data) { ret = data },
		data: {},
		async: false,
	});

	return ret;
}
