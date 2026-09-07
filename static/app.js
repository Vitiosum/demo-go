// Polling of /stats — external file so the page can ship a strict CSP (script-src 'self').
(function () {
  var prev = {};
  var baseUptime = 0;
  var lastFetch = Date.now();
  var pollPill = document.getElementById('poll-pill');

  // Flash the metric in Clever orange for 300 ms when its value changes.
  function flash(id) {
    var el = document.getElementById(id);
    if (!el) return;
    el.classList.remove('cc-stat--flash');
    void el.offsetWidth;
    el.classList.add('cc-stat--flash');
    setTimeout(function () { el.classList.remove('cc-stat--flash'); }, 300);
  }

  function set(valId, val, flashId) {
    var el = document.getElementById(valId);
    if (!el) return;
    var s = String(val);
    if (prev[valId] !== undefined && prev[valId] !== s && flashId !== null) flash(flashId || valId);
    prev[valId] = s;
    el.textContent = s;
  }

  function fmtUptime(sec) {
    if (sec < 60) return sec + 's';
    if (sec < 3600) return Math.floor(sec / 60) + 'm ' + (sec % 60) + 's';
    var h = Math.floor(sec / 3600), m = Math.floor((sec % 3600) / 60), s = sec % 60;
    return h + 'h ' + m + 'm ' + s + 's';
  }

  function setOnline(ok) {
    if (!pollPill) return;
    pollPill.classList.toggle('cc-pill--ok', ok);
    pollPill.classList.toggle('cc-pill--muted', !ok);
    pollPill.lastChild.textContent = ok ? 'Rafraîchi toutes les 2 s' : 'Hors ligne · nouvelle tentative';
  }

  function fetchStats() {
    fetch('/stats').then(function (r) { return r.json(); }).then(function (d) {
      baseUptime = d.uptime_sec;
      lastFetch = Date.now();
      set('v-goroutines', d.goroutines);
      set('v-heap', d.heap_mb, 's-heap');
      set('v-gc', d.gc_cycles);
      set('v-requests', d.requests);
      set('v-version', d.go_version);
      document.getElementById('last-update').textContent = new Date().toTimeString().slice(0, 8);
      setOnline(true);
    }).catch(function () { setOnline(false); });
  }

  // Uptime ticks every second between two polls, without flashing.
  function tickUptime() {
    var elapsed = Math.floor((Date.now() - lastFetch) / 1000);
    set('v-uptime', fmtUptime(baseUptime + elapsed), null);
  }

  fetchStats();
  setInterval(fetchStats, 2000);
  setInterval(tickUptime, 1000);
})();
