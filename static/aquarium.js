// L'aquarium du runtime Go, dessiné sur un <canvas>.
// Un poisson par goroutine, l'eau au niveau du tas (heap), la pompe = le GC, une bulle par pause de GC.
// app.js lui passe chaque mesure de /stats via AquaTank.update(). Fichier séparé : CSP stricte (script-src 'self').
(function () {
  'use strict';
  var canvas = document.getElementById('tank');
  if (!canvas || !canvas.getContext) return;
  var ctx = canvas.getContext('2d');
  var reduced = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  var MAX_FISH = 2500;       // 1 poisson = 1 goroutine, jusqu'à 2500 (BURST_GOROUTINES va jusqu'à 2000)
  var MAX_RESIDENTS = 24;    // goroutines « de fond » (serveur, échantillonneur…) dessinées en gros poissons dorés

  var W = 0, H = 0, DPR = 1;
  var data = { g: 0, heap: 0, gc: null, pause: 0, alert: 32, scaleMax: 80, idle: 0 };
  var level = null;          // niveau d'eau dessiné (MB), qui rejoint data.heap en douceur
  var residents = [], school = [], bubbles = [], pops = [], notes = [];
  var plants = [], gravel = null;
  var pumpSpin = 0, pumpPower = 0, sweep = -1;
  var clock = 0, last = 0, running = false;

  function rand(a, b) { return a + Math.random() * (b - a); }
  function clamp(v, a, b) { return v < a ? a : v > b ? b : v; }

  // ---------- Géométrie du bassin ----------
  function geo() {
    var top = H * 0.08;                          // ligne d'eau la plus haute (sous le bord)
    var floor = H - Math.max(20, H * 0.085);     // dessus du gravier
    var zero = floor - (floor - top) * 0.3;      // 0 MB : le bassin n'est jamais à sec, la règle part d'ici
    var pumpW = clamp(W * 0.05, 30, 52);
    return { top: top, floor: floor, zero: zero, pumpW: pumpW, pumpX: W - pumpW - clamp(W * 0.02, 10, 22), ruler: W < 520 ? 40 : 52 };
  }
  function yFor(mb) {
    var g = geo();
    return Math.max(g.top - 6, g.zero - (mb / data.scaleMax) * (g.zero - g.top));
  }
  function swimBox() {
    var g = geo(), surf = yFor(level);
    return { x0: g.ruler + 10, x1: g.pumpX - 12, y0: surf + 10, y1: g.floor - 8 };
  }

  // ---------- Décor : gravier pré-rendu, plantes ----------
  function buildDecor() {
    var g = geo();
    gravel = document.createElement('canvas');
    gravel.width = Math.ceil(W * DPR); gravel.height = Math.ceil((H - g.floor + 6) * DPR);
    var c = gravel.getContext('2d');
    c.scale(DPR, DPR);
    var gh = H - g.floor + 6;
    var grd = c.createLinearGradient(0, 0, 0, gh);
    grd.addColorStop(0, '#5b5a52'); grd.addColorStop(1, '#2c2b2a');
    c.fillStyle = grd; c.fillRect(0, 4, W, gh);
    var tones = ['#8a8270', '#a39a85', '#6f6a5e', '#c2b79c', '#7d7566', '#5d584f'];
    var n = Math.round(W * gh / 26);
    for (var i = 0; i < n; i++) {
      var x = Math.random() * W, y = 3 + Math.pow(Math.random(), 0.8) * gh, r = rand(1.6, 4.2);
      c.fillStyle = tones[(Math.random() * tones.length) | 0];
      c.beginPath(); c.ellipse(x, y, r * 1.3, r, rand(0, 3), 0, Math.PI * 2); c.fill();
    }
    plants = [];
    var spots = [0.035, 0.09, 0.36, 0.41, 0.7, 0.76];
    spots.forEach(function (p, k) {
      var x = g.ruler + p * (g.pumpX - g.ruler);
      var blades = 4 + (k % 3);
      for (var b = 0; b < blades; b++) {
        plants.push({ x: x + rand(-12, 12), h: (g.floor - g.zero) * rand(0.35, 0.95), lean: rand(-0.5, 0.5),
          phase: rand(0, 6.3), w: rand(3, 6), col: Math.random() < 0.5 ? '#1f7d68' : '#2f9a73' });
      }
    });
  }

  function resize() {
    var r = canvas.getBoundingClientRect();
    if (!r.width || !r.height) return;
    DPR = Math.min(window.devicePixelRatio || 1, 2);
    var oldW = W, oldH = H;
    W = r.width; H = r.height;
    canvas.width = Math.round(W * DPR); canvas.height = Math.round(H * DPR);
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
    if (oldW && oldH) {
      residents.concat(school).forEach(function (f) { f.x *= W / oldW; f.y *= H / oldH; });
    }
    buildDecor();
    if (reduced) render();
  }

  // ---------- Poissons ----------
  function newResident() {
    var b = level == null ? null : swimBox();
    var f = { x: b ? rand(b.x0, b.x1) : W * 0.5, y: b ? rand(b.y0, b.y1) : H * 0.7, vx: rand(-20, 20), vy: 0,
      tx: 0, ty: 0, wait: 0, phase: rand(0, 6.3), dir: 1, size: rand(0.85, 1.15), hue: Math.random() };
    pickTarget(f);
    return f;
  }
  function pickTarget(f) {
    var b = swimBox();
    f.tx = rand(b.x0 + 20, b.x1 - 20); f.ty = rand(b.y0 + 6, b.y1 - 6); f.wait = rand(3, 7);
  }
  function release(n) {
    var b = swimBox();
    var x = b.x0 + 26, y = b.y0 + 6;
    for (var i = 0; i < n; i++) {
      var a = Math.random() * Math.PI * 2, r = Math.sqrt(Math.random());
      school.push({ x: x + rand(-14, 14), y: y + rand(0, 26), vx: rand(30, 160), vy: rand(-10, 110), sp: rand(0.55, 1.35),
        ox: r * Math.cos(a), oy: r * Math.sin(a), depth: Math.random(), phase: rand(0, 6.3), h: 0, alpha: 1, dying: false });
    }
    if (n >= 5) notes.push({ text: '+' + n + ' goroutines', x: x + 8, y: y + 26, life: 2.4 });
  }
  function retire(n) {
    var alive = school.filter(function (f) { return !f.dying; });
    for (var i = 0; i < n && alive.length; i++) {
      var k = (Math.random() * alive.length) | 0;
      alive[k].dying = true;
      alive.splice(k, 1);
    }
  }
  function populate() {
    var n = Math.min(data.g, MAX_FISH);
    var base = Math.min(data.idle, n, MAX_RESIDENTS); // goroutines de fond = poissons résidents, le reste = le banc
    while (residents.length < base) residents.push(newResident());
    if (residents.length > base) residents.length = base;
    var alive = 0;
    school.forEach(function (f) { if (!f.dying) alive++; });
    var want = n - base;
    if (want > alive) release(want - alive);
    else if (want < alive) retire(alive - want);
  }

  function steer(f, tx, ty, dt, k, vmax) {
    var c = 2.6 * Math.sqrt(k) * 0.8; // amortissement proche du critique : le banc s'installe sans rebondir
    f.vx += ((tx - f.x) * k - f.vx * c) * dt;
    f.vy += ((ty - f.y) * k - f.vy * c) * dt;
    var s = Math.sqrt(f.vx * f.vx + f.vy * f.vy);
    if (s > vmax) { f.vx *= vmax / s; f.vy *= vmax / s; }
    f.x += f.vx * dt; f.y += f.vy * dt;
  }
  function keepIn(f, b) {
    if (f.x < b.x0) { f.x = b.x0; f.vx = Math.abs(f.vx) * 0.4; }
    if (f.x > b.x1) { f.x = b.x1; f.vx = -Math.abs(f.vx) * 0.4; }
    if (f.y < b.y0) { f.y += (b.y0 - f.y) * 0.25; f.vy = Math.abs(f.vy) * 0.3; }
    if (f.y > b.y1) { f.y = b.y1; f.vy = -Math.abs(f.vy) * 0.4; }
  }

  // Centre du banc : il dérive lentement d'un bout à l'autre du bassin.
  function schoolFrame(b) {
    var n = school.length, k = clamp(W / 1000, 0.45, 1.1);
    return {
      cx: (b.x0 + b.x1) / 2 + (b.x1 - b.x0) * 0.27 * Math.sin(clock * 0.11),
      cy: (b.y0 + b.y1) / 2 + (b.y1 - b.y0) * 0.2 * Math.sin(clock * 0.19 + 1),
      sx: Math.min((b.x1 - b.x0) * 0.3, 26 + Math.sqrt(n) * 8.5 * k),
      sy: Math.max(8, Math.min((b.y1 - b.y0) * 0.42, 14 + Math.sqrt(n) * 4.2 * k))
    };
  }

  // ---------- Pas de simulation ----------
  function step(dt) {
    clock += dt;
    level += (data.heap - level) * Math.min(1, dt * 2.6);
    var b = swimBox(), sf = schoolFrame(b), i, f;

    for (i = 0; i < school.length; i++) {
      f = school[i];
      var tx = sf.cx + f.ox * sf.sx + Math.sin(clock * 0.8 + f.phase) * 7;
      var ty = sf.cy + f.oy * sf.sy + Math.cos(clock * 1.1 + f.phase) * 4;
      steer(f, tx, ty, dt, (1.2 + f.depth) * f.sp, 70 + 90 * f.sp);
      keepIn(f, b);
      var want = Math.atan2(f.vy * 0.55, f.vx);
      var d = want - f.h;
      while (d > Math.PI) d -= Math.PI * 2;
      while (d < -Math.PI) d += Math.PI * 2;
      f.h += d * Math.min(1, dt * 7);
      if (f.dying) { f.alpha -= dt * 1.2; f.vy -= 40 * dt; }
    }
    school = school.filter(function (s) { return s.alpha > 0; });

    for (i = 0; i < residents.length; i++) {
      f = residents[i];
      f.wait -= dt;
      if (f.wait <= 0 || (Math.abs(f.tx - f.x) < 12 && Math.abs(f.ty - f.y) < 8)) pickTarget(f);
      f.ty = clamp(f.ty, b.y0 + 4, b.y1 - 4);
      steer(f, f.tx, f.ty, dt, 0.5, 46);
      keepIn(f, b);
      if (Math.abs(f.vx) > 3) f.dir = f.vx > 0 ? 1 : -1;
    }

    var surf = yFor(level);
    for (i = 0; i < bubbles.length; i++) {
      var u = bubbles[i];
      u.y -= u.vy * dt;
      u.x += Math.sin(clock * 3 + u.phase) * 10 * dt;
      if (u.y - u.r <= surf) { u.done = true; pops.push({ x: u.x, y: surf, r: u.r, life: 0.5 }); }
    }
    bubbles = bubbles.filter(function (q) { return !q.done; });
    pops.forEach(function (p) { p.life -= dt; });
    pops = pops.filter(function (p) { return p.life > 0; });
    notes.forEach(function (n) { n.life -= dt; n.y -= 8 * dt; });
    notes = notes.filter(function (n) { return n.life > 0; });

    pumpPower = Math.max(0, pumpPower - dt * 0.45);
    pumpSpin += dt * (0.6 + pumpPower * 14);
    if (sweep >= 0) { sweep += dt / 1.6; if (sweep > 1) sweep = -1; }
  }

  // Une pause de GC = une bulle qui part de la crépine de la pompe. Taille selon la durée de la pause.
  function bubble(pauseUs, label) {
    var g = geo();
    var r = clamp(4 + Math.log2(1 + pauseUs / 25) * 2.4, 4, 17);
    var bx = g.pumpX - 14 - r;
    bubbles.push({ x: bx + rand(-3, 3), y: g.floor - 16 - bubbles.length % 3 * 26, r: r, vy: rand(40, 52), phase: rand(0, 6.3), label: label });
    for (var i = 0; i < 4; i++) {
      bubbles.push({ x: g.pumpX - 10 + rand(-5, 3), y: g.floor - 10 - rand(0, 34), r: rand(1.5, 3), vy: rand(50, 80), phase: rand(0, 6.3) });
    }
  }

  // ---------- Rendu ----------
  function waveY(x, surf) {
    return surf + Math.sin(x * 0.018 + clock * 1.6) * 2.2 + Math.sin(x * 0.043 - clock * 1.1) * 1.1;
  }
  function waterPath(surf) {
    ctx.beginPath();
    ctx.moveTo(0, H);
    for (var x = 0; x <= W + 12; x += 12) ctx.lineTo(x, waveY(x, surf));
    ctx.lineTo(W, H);
    ctx.closePath();
  }

  function drawWater(surf) {
    var air = ctx.createLinearGradient(0, 0, 0, H);
    air.addColorStop(0, '#16203d'); air.addColorStop(1, '#0d1631');
    ctx.fillStyle = air; ctx.fillRect(0, 0, W, H);
    // lumière de la rampe, au-dessus de l'eau
    var lamp = ctx.createLinearGradient(0, 0, 0, H * 0.25);
    lamp.addColorStop(0, 'rgba(220, 245, 255, 0.10)'); lamp.addColorStop(1, 'rgba(220, 245, 255, 0)');
    ctx.fillStyle = lamp; ctx.fillRect(0, 0, W, H * 0.25);

    var wg = ctx.createLinearGradient(0, geo().top, 0, H);
    wg.addColorStop(0, '#8fe3f3'); wg.addColorStop(0.3, '#34b8d8'); wg.addColorStop(0.7, '#127da3'); wg.addColorStop(1, '#0a4a6a');
    waterPath(surf);
    ctx.fillStyle = wg; ctx.fill();

    // rayons de lumière dans l'eau
    ctx.save();
    waterPath(surf); ctx.clip();
    ctx.globalCompositeOperation = 'lighter';
    for (var i = 0; i < 4; i++) {
      var x = W * (0.16 + i * 0.22) + Math.sin(clock * 0.3 + i * 1.7) * 18;
      var ray = ctx.createLinearGradient(0, surf, 0, H);
      ray.addColorStop(0, 'rgba(200, 245, 255, 0.10)'); ray.addColorStop(1, 'rgba(200, 245, 255, 0)');
      ctx.fillStyle = ray;
      ctx.beginPath();
      ctx.moveTo(x - 26, surf); ctx.lineTo(x + 30, surf); ctx.lineTo(x + 90, H); ctx.lineTo(x + 10, H);
      ctx.closePath(); ctx.fill();
    }
    ctx.restore();
  }

  function drawSurfaceLine(surf) {
    ctx.beginPath();
    for (var x = 0; x <= W + 12; x += 12) { var y = waveY(x, surf); if (x) ctx.lineTo(x, y); else ctx.moveTo(x, y); }
    ctx.strokeStyle = 'rgba(235, 252, 255, 0.85)'; ctx.lineWidth = 1.6; ctx.stroke();
  }

  function drawPlants() {
    var g = geo();
    plants.forEach(function (p) {
      var sway = reduced ? 0 : Math.sin(clock * 0.9 + p.phase) * 8;
      var tipX = p.x + p.lean * p.h * 0.5 + sway, tipY = g.floor - p.h;
      ctx.beginPath();
      ctx.moveTo(p.x - p.w / 2, g.floor + 2);
      ctx.quadraticCurveTo(p.x + p.lean * p.h * 0.2 - p.w, g.floor - p.h * 0.55, tipX, tipY);
      ctx.quadraticCurveTo(p.x + p.lean * p.h * 0.2 + p.w, g.floor - p.h * 0.5, p.x + p.w / 2, g.floor + 2);
      ctx.fillStyle = p.col; ctx.fill();
    });
  }

  // Poisson du banc (goroutine de la rafale) : un polygone, regroupés par profondeur pour un seul fill().
  function schoolShape(f, L) {
    var c = Math.cos(f.h), s = Math.sin(f.h);
    var flap = reduced ? 1 : 0.65 + 0.35 * Math.sin(clock * 13 + f.phase);
    var pts = [0.5, 0, 0.04, -0.21, -0.3, 0, -0.56, -0.2 * flap, -0.56, 0.2 * flap, -0.3, 0, 0.04, 0.21];
    for (var i = 0; i < pts.length; i += 2) {
      var lx = pts[i] * L, ly = pts[i + 1] * L;
      var x = f.x + lx * c - ly * s, y = f.y + lx * s + ly * c;
      if (i) ctx.lineTo(x, y); else ctx.moveTo(x, y);
    }
    ctx.closePath();
  }
  function drawSchool() {
    var base = clamp(W / 105, 7, 11.5);
    var bands = [
      { min: 0, max: 0.34, col: 'rgba(214, 96, 84, 0.7)', k: 0.78 },
      { min: 0.34, max: 0.68, col: '#f57461', k: 0.94 },
      { min: 0.68, max: 1.01, col: '#ff9a82', k: 1.1 }
    ];
    bands.forEach(function (bd) {
      ctx.beginPath();
      for (var i = 0; i < school.length; i++) {
        var f = school[i];
        if (f.dying || f.depth < bd.min || f.depth >= bd.max) continue;
        schoolShape(f, base * bd.k);
      }
      ctx.fillStyle = bd.col; ctx.fill();
    });
    ctx.fillStyle = '#ff9a82';
    school.forEach(function (f) {
      if (!f.dying) return;
      ctx.globalAlpha = Math.max(0, f.alpha);
      ctx.beginPath(); schoolShape(f, base); ctx.fill();
    });
    ctx.globalAlpha = 1;
  }

  // Poisson résident (goroutine de fond : serveur HTTP, échantillonneur…) : plus gros, doré, avec un œil.
  function drawResident(f) {
    var L = clamp(W / 38, 17, 30) * f.size;
    var flap = reduced ? 0 : Math.sin(clock * 6 + f.phase);
    var pitch = clamp(Math.atan2(f.vy, Math.abs(f.vx) + 8), -0.35, 0.35);
    ctx.save();
    ctx.translate(f.x, f.y);
    ctx.scale(f.dir, 1);
    ctx.rotate(pitch);
    ctx.fillStyle = f.hue < 0.5 ? '#e8a63a' : '#d98b2b';
    ctx.beginPath(); // queue
    ctx.moveTo(-L * 0.32, 0);
    ctx.lineTo(-L * 0.62, -L * (0.24 + flap * 0.05));
    ctx.quadraticCurveTo(-L * 0.52, 0, -L * 0.62, L * (0.24 - flap * 0.05));
    ctx.closePath(); ctx.fill();
    ctx.beginPath(); // nageoire dorsale
    ctx.moveTo(-L * 0.12, -L * 0.17); ctx.quadraticCurveTo(-L * 0.02, -L * 0.36, L * 0.14, -L * 0.18); ctx.fill();
    var body = ctx.createLinearGradient(0, -L * 0.22, 0, L * 0.22);
    body.addColorStop(0, '#ffd772'); body.addColorStop(0.55, '#f2c14e'); body.addColorStop(1, '#fff1c2');
    ctx.fillStyle = body;
    ctx.beginPath(); ctx.ellipse(0, 0, L * 0.42, L * 0.22, 0, 0, Math.PI * 2); ctx.fill();
    ctx.fillStyle = '#fff';
    ctx.beginPath(); ctx.arc(L * 0.24, -L * 0.05, L * 0.055, 0, Math.PI * 2); ctx.fill();
    ctx.fillStyle = '#14203a';
    ctx.beginPath(); ctx.arc(L * 0.255, -L * 0.05, L * 0.03, 0, Math.PI * 2); ctx.fill();
    ctx.restore();
  }

  function roundRect(x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y); ctx.lineTo(x + w - r, y); ctx.quadraticCurveTo(x + w, y, x + w, y + r);
    ctx.lineTo(x + w, y + h - r); ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
    ctx.lineTo(x + r, y + h); ctx.quadraticCurveTo(x, y + h, x, y + h - r);
    ctx.lineTo(x, y + r); ctx.quadraticCurveTo(x, y, x + r, y); ctx.closePath();
  }

  // La pompe du GC : une cuve de filtre accrochée à droite, sa crépine au fond, son rotor visible.
  function drawPump() {
    var g = geo(), x = g.pumpX, w = g.pumpW, y0 = H * 0.04, y1 = g.zero - (g.zero - g.top) * 0.15;
    var cx = x + w / 2;
    ctx.fillStyle = '#4a5184'; // tube d'aspiration
    ctx.fillRect(cx - w * 0.14, y1 - 4, w * 0.28, g.floor - y1 - 6);
    ctx.fillStyle = '#272b4d';
    for (var k = 0; k < 4; k++) ctx.fillRect(cx - w * 0.22, g.floor - 30 + k * 6, w * 0.44, 3);
    var body = ctx.createLinearGradient(x, 0, x + w, 0);
    body.addColorStop(0, '#3a4170'); body.addColorStop(0.42, '#6a74ad'); body.addColorStop(1, '#2a2f5a');
    roundRect(x, y0, w, y1 - y0, 6); ctx.fillStyle = body; ctx.fill();
    ctx.fillStyle = '#c9a35b';
    ctx.fillRect(x, y0 + 8, w, 3); ctx.fillRect(x, y1 - 11, w, 3);
    // hublot du rotor
    var ry = y0 + (y1 - y0) * 0.36, rr = w * 0.32;
    ctx.fillStyle = '#0c1430'; ctx.beginPath(); ctx.arc(cx, ry, rr + 2, 0, Math.PI * 2); ctx.fill();
    ctx.strokeStyle = pumpPower > 0.05 ? 'rgba(95, 227, 211, ' + (0.45 + pumpPower * 0.55) + ')' : 'rgba(222, 221, 238, 0.45)';
    ctx.lineWidth = 2;
    for (var b = 0; b < 3; b++) {
      var a = pumpSpin + b * 2.094;
      ctx.beginPath(); ctx.moveTo(cx, ry); ctx.lineTo(cx + Math.cos(a) * rr * 0.85, ry + Math.sin(a) * rr * 0.85); ctx.stroke();
    }
    ctx.fillStyle = '#e8eefc';
    ctx.font = '600 ' + clamp(w * 0.3, 9, 12) + 'px "JetBrains Mono", ui-monospace, monospace';
    ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
    ctx.fillText('GC', cx, ry + rr + 14);
    ctx.fillStyle = '#efd99c';
    ctx.fillText(data.gc == null ? '—' : String(data.gc), cx, ry + rr + 30);
    // voyant : allumé tant que la pompe tourne
    ctx.fillStyle = pumpPower > 0.05 ? '#5fe3d3' : '#2b3a52';
    ctx.beginPath(); ctx.arc(cx, y0 + 20, 3, 0, Math.PI * 2); ctx.fill();
  }

  // Le passage du filtre : une nappe claire qui traverse le bassin de droite à gauche.
  function drawSweep(surf) {
    if (sweep < 0) return;
    var g = geo(), x = g.pumpX - sweep * (g.pumpX - g.ruler + 60);
    ctx.save(); waterPath(surf); ctx.clip();
    var sw = ctx.createLinearGradient(x - 70, 0, x + 20, 0);
    var a = 0.22 * Math.sin(Math.PI * sweep);
    sw.addColorStop(0, 'rgba(95, 227, 211, 0)'); sw.addColorStop(0.75, 'rgba(180, 255, 245, ' + a + ')'); sw.addColorStop(1, 'rgba(95, 227, 211, 0)');
    ctx.fillStyle = sw; ctx.fillRect(x - 70, surf, 90, H - surf);
    ctx.restore();
  }

  function drawBubbles() {
    ctx.textAlign = 'right'; ctx.textBaseline = 'middle';
    bubbles.forEach(function (u) {
      ctx.beginPath(); ctx.arc(u.x, u.y, u.r, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(230, 252, 255, 0.18)'; ctx.fill();
      ctx.strokeStyle = 'rgba(240, 253, 255, 0.85)'; ctx.lineWidth = 1.2; ctx.stroke();
      ctx.beginPath(); ctx.arc(u.x - u.r * 0.35, u.y - u.r * 0.35, Math.max(1, u.r * 0.22), 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(255, 255, 255, 0.9)'; ctx.fill();
      if (u.label) {
        ctx.font = '600 12px "JetBrains Mono", ui-monospace, monospace';
        var tw = ctx.measureText(u.label).width;
        ctx.fillStyle = 'rgba(8, 40, 60, 0.55)';
        roundRect(u.x - u.r - tw - 16, u.y - 10, tw + 10, 20, 5); ctx.fill();
        ctx.fillStyle = '#f2fdff';
        ctx.fillText(u.label, u.x - u.r - 11, u.y + 0.5);
      }
    });
    pops.forEach(function (p) {
      ctx.beginPath(); ctx.ellipse(p.x, p.y, p.r * (1.6 - p.life), p.r * 0.4, 0, 0, Math.PI * 2);
      ctx.strokeStyle = 'rgba(240, 253, 255, ' + p.life * 1.4 + ')'; ctx.lineWidth = 1; ctx.stroke();
    });
  }

  // La règle graduée collée sur la vitre, et la ligne rouge gravée de HEAP_ALERT_MB.
  function drawRuler(surf) {
    var g = geo(), rw = g.ruler;
    ctx.fillStyle = 'rgba(9, 13, 32, 0.62)';
    ctx.fillRect(0, g.top - 28, rw, g.zero - g.top + 42);
    var step = data.scaleMax / 4;
    ctx.font = '500 ' + (W < 520 ? 10 : 11) + 'px "JetBrains Mono", ui-monospace, monospace';
    ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
    ctx.strokeStyle = 'rgba(232, 238, 252, 0.7)'; ctx.fillStyle = 'rgba(232, 238, 252, 0.9)'; ctx.lineWidth = 1;
    for (var i = 0; i <= 20; i++) {
      var mb = data.scaleMax * i / 20, y = Math.round(yFor(mb)) + 0.5, major = i % 5 === 0;
      ctx.beginPath(); ctx.moveTo(rw - (major ? 12 : 6), y); ctx.lineTo(rw, y); ctx.stroke();
      if (major) ctx.fillText(fmtMB(i / 5 * step), 5, y);
    }
    ctx.fillStyle = 'rgba(232, 238, 252, 0.6)';
    ctx.fillText('MB', 5, g.top - 18);
    // ligne d'alerte
    var ya = Math.round(yFor(data.alert)) + 0.5, over = data.heap > data.alert;
    ctx.save();
    ctx.setLineDash(over ? [] : [7, 5]);
    ctx.strokeStyle = over ? '#ff3d5e' : 'rgba(255, 90, 114, 0.95)'; ctx.lineWidth = over ? 2 : 1.5;
    ctx.beginPath(); ctx.moveTo(rw, ya); ctx.lineTo(g.pumpX - 4, ya); ctx.stroke();
    ctx.restore();
    var label = 'HEAP_ALERT_MB ' + data.alert + ' MB';
    ctx.font = '600 11px "JetBrains Mono", ui-monospace, monospace';
    var tw = ctx.measureText(label).width;
    ctx.fillStyle = over ? '#d0213f' : 'rgba(9, 13, 32, 0.8)';
    roundRect(rw + 8, ya - 21, tw + 12, 18, 4); ctx.fill();
    ctx.strokeStyle = '#ff5a72'; ctx.lineWidth = 1; ctx.stroke();
    ctx.fillStyle = over ? '#fff' : '#ff8a9b';
    ctx.fillText(label, rw + 14, ya - 11.5);
    // repère du niveau actuel sur la règle
    var tag = fmtMB(level, true);
    ctx.font = '600 11px "JetBrains Mono", ui-monospace, monospace';
    var lw = ctx.measureText(tag).width;
    ctx.fillStyle = '#e9fbff';
    ctx.beginPath(); ctx.moveTo(rw, surf); ctx.lineTo(rw + 7, surf - 6); ctx.lineTo(rw + 7, surf + 6); ctx.closePath(); ctx.fill();
    roundRect(rw + 7, surf - 9, lw + 10, 18, 4); ctx.fill();
    ctx.fillStyle = '#0a3d58';
    ctx.fillText(tag, rw + 12, surf + 0.5);
  }
  function fmtMB(v, precise) {
    if (precise) return (v < 10 ? v.toFixed(2) : v.toFixed(1)) + ' MB';
    return v >= 100 || v === Math.round(v) ? String(Math.round(v)) : v.toFixed(1);
  }

  function drawNotes() {
    ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
    ctx.font = '600 13px "Plus Jakarta Sans", system-ui, sans-serif';
    notes.forEach(function (n) {
      ctx.fillStyle = 'rgba(255, 255, 255, ' + Math.min(1, n.life) + ')';
      ctx.fillText(n.text, n.x, n.y);
    });
  }

  function render() {
    if (!W || level == null) return;
    var g = geo(), surf = yFor(level);
    ctx.clearRect(0, 0, W, H);
    drawWater(surf);
    drawPlants();
    if (gravel) ctx.drawImage(gravel, 0, g.floor - 6, W, gravel.height / DPR);
    drawSweep(surf);
    drawSchool();
    residents.forEach(drawResident);
    drawPump();
    drawBubbles();
    drawSurfaceLine(surf);
    drawRuler(surf);
    drawNotes();
  }

  function frame(now) {
    var dt = Math.min(0.05, (now - last) / 1000 || 0.016);
    last = now;
    step(dt);
    render();
    requestAnimationFrame(frame);
  }

  // Sans animation : chaque mesure place directement les poissons et les bulles.
  function settle() {
    level = data.heap;
    var b = swimBox(), sf = schoolFrame(b);
    school = school.filter(function (f) { return !f.dying; });
    school.forEach(function (f) { f.x = sf.cx + f.ox * sf.sx; f.y = sf.cy + f.oy * sf.sy; f.h = f.phase > 3.14 ? 0 : Math.PI; f.vx = 0; f.vy = 0; keepIn(f, b); });
    residents.forEach(function (f, i) {
      f.x = b.x0 + (i + 0.5) * (b.x1 - b.x0) / residents.length;
      f.y = b.y1 - 14 - (i % 3) * 22; f.dir = i % 2 ? 1 : -1; f.vx = 0; f.vy = 0; keepIn(f, b);
    });
    notes = []; pops = [];
    var g = geo();
    bubbles = data.pause ? [{ x: g.pumpX - 24, y: (yFor(level) + g.floor) / 2, r: clamp(4 + Math.log2(1 + data.pause / 25) * 2.4, 4, 17), vy: 0, phase: 0, label: pauseText(data.pause) }] : [];
    render();
  }

  function pauseText(us) { return us >= 1000 ? (us / 1000).toFixed(1) + ' ms' : Math.round(us) + ' µs'; }

  window.AquaTank = {
    // s : une mesure de /stats ({g, h, gc, p}) ; opts : {alert, scaleMax, idle}
    update: function (s, opts) {
      var first = level == null;
      data.alert = opts.alert; data.scaleMax = opts.scaleMax; data.idle = opts.idle;
      data.g = s.g; data.heap = s.h; data.pause = s.p;
      if (first) { resize(); level = s.h; }
      if (data.gc != null && s.gc > data.gc) {
        var cycles = Math.min(s.gc - data.gc, 5);
        for (var i = 0; i < cycles; i++) bubble(s.p, i === 0 ? pauseText(s.p) : null);
        pumpPower = 1; sweep = 0;
      } else if (first && s.p) {
        bubble(s.p, pauseText(s.p));
      }
      data.gc = s.gc;
      populate();
      if (first) residents.concat(school).forEach(function (f) { f.x = rand(swimBox().x0, swimBox().x1); f.y = rand(swimBox().y0, swimBox().y1); });
      if (reduced) { settle(); return; }
      if (!running) { running = true; last = performance.now(); requestAnimationFrame(frame); }
    }
  };

  if (window.ResizeObserver) new ResizeObserver(resize).observe(canvas);
  else window.addEventListener('resize', resize);
})();
