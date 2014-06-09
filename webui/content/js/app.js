(function (){
	$(document).ready(function () {
		setFromNow();
		setLogSize();

		$('.search-button').click(function () {
			var from = getFrom(),
				range = getRange(),
				minSev = getSev(true),
				maxSev = getSev(false),
				sources = getSources(),
				keywords = getKeywords(),
				query = {
					from: from,
					range: range,
					minSev: minSev,
					maxSev: maxSev,
					sources: sources,
					keywords: keywords
				};

			if (from && range && minSev && maxSev) {
				$('.log').
					empty().
					append('<img src="/img/loading.gif" class="loading">');

				$.get('log', query).done(function (data) {
					var log = $('.log').empty(),
						i;

					if (data.length > 0) {
						log.append(
							$('<p></p>').append(
								$('<a></a>').
									attr('href', String.prototype.concat('log?', $.param(query), '&text=1')).
									attr('target', '_blank').
									text('Get as text.')
							)
						);
					}

					for (i = 0; i < data.length; i++) {
						log.append(
							$('<pre></pre>').text(
								formatLogEntry(data[i])
							)
						);
					}
				}).fail(function(result) {
					$('.log').empty().append(
						$('<pre></pre>').text(
							result.status + ': ' + result.statusText
						)
					);
				});
			}
		});

		$('.now-button').click(function() {
			setFromNow();
		});
	});

	$(window).resize(function (){
		setLogSize();
	});

	function setLogSize() {
		var log = $('.log'),
			logPadding = parseInt(log.css('padding')),
			bodyMargin = parseInt($('body').css('margin-top')),
			windowHeight = $(window).innerHeight();

		log.css('height', (windowHeight - 2*bodyMargin - 2*logPadding) + 'px');
	}

	function formatLogEntry(ent) {
		var ts = new Date(Math.floor(ent.Ts / 1000000))

		return String.prototype.concat(
			ts.toLocaleDateString(),
			' ',
			ts.toLocaleTimeString(),
			' [',
			ent.Src,
			'] ',
			ent.Sev,
			': ',
			ent.Msg
		);
	}

	function setFromNow() {
		var now = new Date();

		$('#date-from').val(String.prototype.concat(
			now.getFullYear(),
			'-',
			zeroPrefix(now.getMonth() + 1),
			'-',
			zeroPrefix(now.getDate())
		));

		$('#time-from').val(String.prototype.concat(
			zeroPrefix(now.getHours()),
			':',
			zeroPrefix(now.getMinutes()),
			':',
			zeroPrefix(now.getSeconds())
		));

		function zeroPrefix(n) {
			return ((n < 10) ? '0' : '') + n;
		}
	}

	function getFrom() {
		var dateFormat = /^\d{4}-\d{2}-\d{2}$/,
			timeFormat = /^\d{2}:\d{2}:\d{2}$/,
			date = $('#date-from'),
			time = $('#time-from'),
			dateVal = date.val(),
			timeVal = time.val(),
			ok;

		if (!dateFormat.test(dateVal)) {
			date.addClass('error');
		} else {
			date.removeClass('error');
			ok = true;
		}

		if (!timeFormat.test(timeVal)) {
			time.addClass('error');
		} else {
			time.removeClass('error');
			ok = true;
		}

		return (ok) ? dateVal.concat(' ', timeVal) : void 0;
	}

	function getRange() {
		var format = /^[+-]?\d+$/,
			range = $('#range'),
			rangeVal = range.val(),
			uomVal = $('#range-unit-s:checked').val() || 
					 $('#range-unit-m:checked').val() ||
					 $('#range-unit-h:checked').val();

		if (!format.test(rangeVal)) {
			range.addClass('error');
			return void 0;
		} else {
			range.removeClass('error');
		}

		return rangeVal + uomVal;
	}

	function getSev(min) {
		var format = /^\d+$/,
			sev = (min) ? $('#min-sev') : $('#max-sev'),
			sevVal = sev.val();

		if (!format.test(sevVal)) {
			sev.addClass('error');
			return void 0;
		} else {
			sev.removeClass('error');
		}

		return sevVal;
	}

	function getSources() {
		return $('#sources').val();
	}

	function getKeywords() {
		return $('#keywords').val();
	}
})();