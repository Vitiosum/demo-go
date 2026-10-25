// Oscilloscope and load buttons. External file so the page keeps a strict CSP (script-src 'self').
(function () {
  var W = 900, H = 260, SPAN = 90; // viewBox width/height, seconds shown
  var root = document.querySelector('.go-pulse');
  var alertMB = Number(root && root.dataset.alertMb) || 32;
  var samples = [];
  var baseUptime = 0, lastFetch = Date.now(), online = true;

  function $(id) { return document.getElementById(id); }
  function text(id, v) { var el = $(id); if (el && el.firstChild) el.firstChild.nodeValue = String(v); }

  function fmtUptime(sec) {
    if (sec < 60) return sec + 's';
    if (sec < 3600) return Math.floor(sec / 60) + 'm ' + String(sec % 60).padStart(2, '0') + 's';
    return Math.floor(sec / 3600) + 'h ' + String(Math.floor((sec % 3600) / 60)).padStart(2, '0') + 'm';
  }

  function nice(v) { // round an axis maximum up to 1, 2 or 5 × 10^n
    var p = Math.pow(10, Math.floor(Math.log10(v))), m = v / p;
    return (m <= 1 ? 1 : m <= 2 ? 2 : m <= 5 ? 5 : 10) * p;
  }

  function add(list) {
    var last = samples.length ? samples[samples.length - 1].t : 0;
    list.forEach(function (s) { if (s.t > last) { samples.push(s); last = s.t; } });
    if (samples.length > SPAN) samples = samples.slice(samples.length - SPAN);
  }

  function draw() {
    if (!samples.length) return;
    var now = samples[samples.length - 1].t;
    var x = function (t) { return W - ((now - t) / 1000) * (W / (SPAN - 1)); };
    var gMax = 10, hMax = alertMB * 1.25;
    samples.forEach(function (s) { gMax = Math.max(gMax, s.g); hMax = Math.max(hMax, s.h); });
    gMax = nice(gMax * 1.15); hMax = nice(hMax * 1.1);
    var yg = function (v) { return H - 8 - (v / gMax) * (H - 40); };
    var yh = function (v) { return H - (v / hMax) * (H - 32); };
    var g = '', h = '', gc = '';
    samples.forEach(function (s, i) {
      var cmd = i ? 'L' : 'M', px = x(s.t).toFixed(1);
      g += cmd + px + ' ' + yg(s.g).toFixed(1);
      h += cmd + px + ' ' + yh(s.h).toFixed(1);
      if (i && s.gc > samples[i - 1].gc) gc += 'M' + px + ' ' + (H - 22) + 'V' + H;
    });
    var x0 = x(samples[0].t).toFixed(1);
    $('p-g').setAttribute('d', g);
    $('p-heap').setAttribute('d', h);
    $('p-heap-area').setAttribute('d', h + 'L' + W + ' ' + H + 'L' + x0 + ' ' + H + 'Z');
    $('p-gc').setAttribute('d', gc);
    var ya = yh(alertMB);
    $('l-alert').setAttribute('y1', ya); $('l-alert').setAttribute('y2', ya);
    $('lbl-alert').style.top = (ya / H * 100) + '%';
    $('ax-heap').textContent = hMax + ' MB';
    $('ax-g').textContent = gMax + ' goroutines';
  }

  function show(d) {
    var s = d.now;
    baseUptime = d.uptime_sec; lastFetch = Date.now();
    $('v-uptime').textContent = fmtUptime(baseUptime);
    $('v-g').textContent = s.g;
    $('v-heap-sub').textContent = s.h.toFixed(2);
    $('v-gc-sub').textContent = s.gc;
    text('v-heap', s.h.toFixed(2));
    $('v-gc').textContent = s.gc;
    $('v-pause').textContent = s.p >= 1000 ? (s.p / 1000).toFixed(1) + ' ms' : Math.round(s.p) + ' µs';
    $('v-req').textContent = d.requests;
    var over = s.h > alertMB;
    $('v-heap').classList.toggle('is-over', over);
    $('over').hidden = !over;
    state('goroutines', d.burst_left_ms);
    state('heap', d.heap_left_ms);
  }

  function state(kind, leftMs) {
    var btn = document.querySelector('[data-load="' + kind + '"]');
    var el = document.querySelector('[data-state="' + kind + '"]');
    if (!btn || !el) return;
    btn.setAttribute('aria-disabled', leftMs > 0 ? 'true' : 'false');
    el.textContent = leftMs > 0 ? Math.ceil(leftMs / 1000) + ' s' : '';
  }

  function poll(withHistory) {
    fetch('/stats' + (withHistory ? '?history=1' : ''), { cache: 'no-store' })
      .then(function (r) { if (!r.ok) throw new Error(r.status); return r.json(); })
      .then(function (d) {
        add(d.history || [d.now]); show(d); draw();
        if (!online) { online = true; toast('', false); }
      })
      .catch(function () { online = false; toast('Hors ligne · nouvelle tentative', true); });
  }

  var toastTimer;
  function toast(msg, busy) {
    var el = $('toast');
    el.textContent = msg; el.classList.toggle('is-busy', !!busy);
    clearTimeout(toastTimer);
    if (msg && online) toastTimer = setTimeout(function () { el.textContent = ''; }, 4000);
  }

  document.querySelectorAll('[data-load]').forEach(function (btn) {
    btn.addEventListener('click', function () {
      fetch('/load/' + btn.dataset.load, { method: 'POST' })
        .then(function (r) { return r.json(); })
        .then(function (d) { toast(d.message, !d.ok); setTimeout(function () { poll(false); }, 150); })
        .catch(function () { toast('Requête refusée', true); });
    });
  });

  setInterval(function () {
    var el = $('v-uptime');
    if (el) el.textContent = fmtUptime(baseUptime + Math.floor((Date.now() - lastFetch) / 1000));
  }, 1000);

  poll(true);
  setInterval(function () { poll(false); }, 1000);
})();
