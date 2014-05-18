(function (){
	$(document).ready(function () {
		setFromNow();

		$('.search-button').click(function () {
			var from = getFrom(),
				range = getRange(),
				minSev = getSev(true),
				maxSev = getSev(false),
				sources = getSources(),
				keywords = getKeywords();

			if (from && range && minSev && maxSev) {
				$.get('log', {
					from: from,
					range: range,
					minSev: minSev,
					maxSev: maxSev,
					sources: sources,
					keywords: keywords
				}).done(function () {
					alert('ok');
				}).fail(function() {
					alert('error');
				});
			}
		});
	});

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
					 $('#range-unit-h:checked').val() || 
					 $('#range-unit-d:checked').val();

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