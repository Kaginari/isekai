// net3d.js — the neural net in depth: a hand-written WebGL2 stage (no library). The tree stands as
// planes: the throne at the top, each rank a ring below it (a parent sits at the middle of its
// children's arc), the minds a halo beneath. Nodes are glowing sprites (a core and a soft radial
// halo, additive on a dark theme so they bloom, alpha-blended on a light one), bonds are
// screen-space ribbons weighted by the commissions they carried, and particles drift along the
// bonds the way work travels — down. Drag orbits, the wheel or a pinch zooms, a click focuses (the
// camera eases to it); a dispatch sends a spark down and a report one back up. It renders only
// while seen, idles at 30 frames a second, and is still (no particles, no spin) under
// prefers-reduced-motion. Colours are the css tokens, read through probe elements.
(() => {
  'use strict';
  const TAU = Math.PI * 2;
  const clamp = (v, a, b) => Math.max(a, Math.min(b, v));

  function supported() {
    if (supported.v !== undefined) return supported.v;
    try { const c = document.createElement('canvas'); supported.v = !!c.getContext('webgl2'); } catch (_) { supported.v = false; }
    return supported.v;
  }

  // ---------- tiny matrix kit (column-major) ----------
  function perspective(fovy, aspect, near, far) {
    const f = 1 / Math.tan(fovy / 2), nf = 1 / (near - far);
    return new Float32Array([f / aspect, 0, 0, 0, 0, f, 0, 0, 0, 0, (far + near) * nf, -1, 0, 0, 2 * far * near * nf, 0]);
  }
  function lookAt(e, t, u) {
    let zx = e[0] - t[0], zy = e[1] - t[1], zz = e[2] - t[2];
    let l = Math.hypot(zx, zy, zz) || 1; zx /= l; zy /= l; zz /= l;
    let xx = u[1] * zz - u[2] * zy, xy = u[2] * zx - u[0] * zz, xz = u[0] * zy - u[1] * zx;
    l = Math.hypot(xx, xy, xz) || 1; xx /= l; xy /= l; xz /= l;
    const yx = zy * xz - zz * xy, yy = zz * xx - zx * xz, yz = zx * xy - zy * xx;
    return new Float32Array([xx, yx, zx, 0, xy, yy, zy, 0, xz, yz, zz, 0,
      -(xx * e[0] + xy * e[1] + xz * e[2]), -(yx * e[0] + yy * e[1] + yz * e[2]), -(zx * e[0] + zy * e[1] + zz * e[2]), 1]);
  }
  function mul(a, b) {
    const o = new Float32Array(16);
    for (let c = 0; c < 4; c++) for (let r = 0; r < 4; r++) {
      let s = 0;
      for (let k = 0; k < 4; k++) s += a[k * 4 + r] * b[c * 4 + k];
      o[c * 4 + r] = s;
    }
    return o;
  }

  // ---------- shaders ----------
  const POINT_VS = `#version 300 es
  in vec3 aPos; in vec4 aColor; in float aSize; in float aCore;
  uniform mat4 uVP; uniform float uScale;
  out vec4 vColor; out float vCore;
  void main() {
    vec4 p = uVP * vec4(aPos, 1.0);
    gl_Position = p;
    gl_PointSize = clamp(aSize * uScale / max(p.w, 0.01), 1.0, 512.0);
    vColor = aColor; vCore = aCore;
  }`;
  const POINT_FS = `#version 300 es
  precision mediump float;
  in vec4 vColor; in float vCore;
  out vec4 o;
  void main() {
    vec2 d = gl_PointCoord * 2.0 - 1.0;
    float r = length(d);
    if (r > 1.0) discard;
    float core = vCore > 0.0 ? 1.0 - smoothstep(vCore - 0.12, vCore, r) : 0.0;
    float rim = vCore > 0.0 ? smoothstep(vCore - 0.3, vCore - 0.12, r) * core : 0.0;
    float glow = exp(-r * r * 4.5) * (1.0 - core);
    vec3 c = mix(vColor.rgb, vColor.rgb * 1.25 + 0.08, rim);
    float a = vColor.a * clamp(core + glow * 0.6, 0.0, 1.0);
    o = vec4(c * a, a);
  }`;
  const LINE_VS = `#version 300 es
  in vec2 aCorner;
  in vec3 aA; in vec3 aB; in vec4 aCol; in float aWidth;
  uniform mat4 uVP; uniform vec2 uView;
  out vec4 vColor; out float vSide;
  void main() {
    vec4 pa = uVP * vec4(aA, 1.0), pb = uVP * vec4(aB, 1.0);
    vec2 sa = pa.xy / pa.w, sb = pb.xy / pb.w;
    vec2 d = (sb - sa) * uView;
    vec2 dir = length(d) > 0.0001 ? normalize(d) : vec2(1.0, 0.0);
    vec2 n = vec2(-dir.y, dir.x);
    vec4 p = mix(pa, pb, aCorner.x);
    p.xy += n * aCorner.y * aWidth / uView * p.w;
    gl_Position = p;
    vColor = aCol; vSide = aCorner.y;
  }`;
  const LINE_FS = `#version 300 es
  precision mediump float;
  in vec4 vColor; in float vSide;
  out vec4 o;
  void main() {
    float a = vColor.a * (1.0 - smoothstep(0.35, 1.0, abs(vSide)));
    o = vec4(vColor.rgb * a, a);
  }`;
  function program(gl, vs, fs) {
    const mk = (type, src) => { const sh = gl.createShader(type); gl.shaderSource(sh, src); gl.compileShader(sh); if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS)) throw new Error(gl.getShaderInfoLog(sh)); return sh; };
    const p = gl.createProgram();
    gl.attachShader(p, mk(gl.VERTEX_SHADER, vs)); gl.attachShader(p, mk(gl.FRAGMENT_SHADER, fs));
    gl.linkProgram(p);
    if (!gl.getProgramParameter(p, gl.LINK_STATUS)) throw new Error(gl.getProgramInfoLog(p));
    return p;
  }

  // ---------- colours: the tokens, through probes, as linear-ish rgb ----------
  const conv = document.createElement('canvas');
  conv.width = conv.height = 1;
  const cctx = conv.getContext('2d', { willReadFrequently: true });
  function rgb(css) {
    cctx.clearRect(0, 0, 1, 1);
    cctx.fillStyle = '#000';
    cctx.fillStyle = css;
    cctx.fillRect(0, 0, 1, 1);
    const d = cctx.getImageData(0, 0, 1, 1).data;
    return [d[0] / 255, d[1] / 255, d[2] / 255];
  }

  function create(stage, opts) {
    if (!supported()) return null;
    const canvas = document.createElement('canvas');
    canvas.className = 'net-canvas';
    canvas.tabIndex = 0;
    canvas.setAttribute('role', 'img');
    canvas.setAttribute('aria-label', 'the net in 3D — drag to orbit, wheel to zoom, click a neuron to focus it; arrow keys turn it, + and - zoom');
    const overlay = document.createElement('canvas');
    overlay.className = 'net-overlay';
    overlay.setAttribute('aria-hidden', 'true');
    const gl = canvas.getContext('webgl2', { antialias: true, premultipliedAlpha: true, alpha: true });
    if (!gl) return null;
    const octx = overlay.getContext('2d');
    let P, LP, quad, pbuf, lbuf, vaoP, vaoL;
    function init() {
      P = program(gl, POINT_VS, POINT_FS);
      LP = program(gl, LINE_VS, LINE_FS);
      pbuf = gl.createBuffer(); lbuf = gl.createBuffer(); quad = gl.createBuffer();
      gl.bindBuffer(gl.ARRAY_BUFFER, quad);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([0, -1, 1, -1, 1, 1, 0, -1, 1, 1, 0, 1]), gl.STATIC_DRAW);
      vaoP = gl.createVertexArray();
      gl.bindVertexArray(vaoP);
      gl.bindBuffer(gl.ARRAY_BUFFER, pbuf);
      const pl = n => gl.getAttribLocation(P, n);
      const PS = 9 * 4; // pos3 col4 size core
      gl.enableVertexAttribArray(pl('aPos')); gl.vertexAttribPointer(pl('aPos'), 3, gl.FLOAT, false, PS, 0);
      gl.enableVertexAttribArray(pl('aColor')); gl.vertexAttribPointer(pl('aColor'), 4, gl.FLOAT, false, PS, 12);
      gl.enableVertexAttribArray(pl('aSize')); gl.vertexAttribPointer(pl('aSize'), 1, gl.FLOAT, false, PS, 28);
      gl.enableVertexAttribArray(pl('aCore')); gl.vertexAttribPointer(pl('aCore'), 1, gl.FLOAT, false, PS, 32);
      vaoL = gl.createVertexArray();
      gl.bindVertexArray(vaoL);
      const ll = n => gl.getAttribLocation(LP, n);
      gl.bindBuffer(gl.ARRAY_BUFFER, quad);
      gl.enableVertexAttribArray(ll('aCorner')); gl.vertexAttribPointer(ll('aCorner'), 2, gl.FLOAT, false, 8, 0);
      gl.bindBuffer(gl.ARRAY_BUFFER, lbuf);
      const LS = 11 * 4; // a3 b3 col4 width
      const inst = (n, size, off) => { const loc = ll(n); gl.enableVertexAttribArray(loc); gl.vertexAttribPointer(loc, size, gl.FLOAT, false, LS, off); gl.vertexAttribDivisor(loc, 1); };
      inst('aA', 3, 0); inst('aB', 3, 12); inst('aCol', 4, 24); inst('aWidth', 1, 40);
      gl.bindVertexArray(null);
    }
    init();

    const st = {
      m: null, nodes: new Map(), edges: [], wears: [], rings: [], focus: '', shown: '', lit: null, live: new Set(), hover: '',
      yaw: 0.7, pitch: 0.42, dist: 10, dist0: 10, target: [0, 0, 0], goal: { target: [0, 0, 0], dist: 10 },
      run: false, raf: 0, last: 0, t: 0, idleAt: 0, pulses: [], flashes: new Map(), pal: null, dark: true, W: 1, H: 1, dpr: 1, VP: null,
      slow: 0,
    };

    // ---------- the scene: a radial tidy tree, one plane per layer ----------
    function scene(m) {
      st.m = m;
      st.nodes = new Map();
      const leaves = [], slot = new Map();
      const place = id => {
        const ks = m.kids.get(id) || [];
        if (!ks.length) { slot.set(id, leaves.length); leaves.push(id); return; }
        ks.forEach(place);
        slot.set(id, ks.reduce((a, k) => a + slot.get(k), 0) / ks.length);
      };
      if (m.throne) place(m.throne.id);
      for (const id of m.order) if (!slot.has(id)) place(id);
      const D = Math.max(1, m.maxDepth), N = Math.max(1, leaves.length);
      const R = clamp(2 + N * 0.11, 2.6, 6.2), DY = 1.35;
      const y0 = (D * DY) / 2 + 0.3;
      st.rings = [];
      for (const id of m.order) {
        const n = m.byId.get(id), d = m.depth.get(id) || 0;
        const a = (slot.get(id) / N) * TAU + 0.5 / N * TAU;
        const r = d === 0 ? 0 : R * Math.pow(d / D, 0.9);
        const size = n.throne ? 0.62 : d === 1 ? 0.42 : d === 2 ? 0.34 : 0.27;
        st.nodes.set(id, { id, n, p: [Math.cos(a) * r, y0 - d * DY, Math.sin(a) * r], a, d, size, rank: n.rank || 'plain' });
      }
      for (let d = 1; d <= m.maxDepth; d++) st.rings.push({ y: y0 - d * DY, r: R * Math.pow(d / D, 0.9) });
      // minds: a halo beneath, each at the middle of its wearers, spread so none overlap
      if (m.minds.length) {
        const my = y0 - (m.maxDepth + 1) * DY, mr = R * 0.78;
        const want = m.minds.map(n => {
          let sx = 0, sz = 0, k = 0;
          for (const w of m.wears) if (w.to === n.id && st.nodes.has(w.from)) { const q = st.nodes.get(w.from); sx += Math.cos(q.a); sz += Math.sin(q.a); k++; }
          return { n, a: k ? Math.atan2(sz, sx) : Math.PI };
        }).sort((x, y) => x.a - y.a);
        const gap = TAU / want.length * 0.7;
        for (let i = 1; i < want.length; i++) if (want[i].a - want[i - 1].a < gap) want[i].a = want[i - 1].a + gap;
        want.forEach(w => st.nodes.set(w.n.id, { id: w.n.id, n: w.n, p: [Math.cos(w.a) * mr, my, Math.sin(w.a) * mr], a: w.a, d: -1, size: 0.2, rank: w.n.lane || 'mind:shared' }));
        st.rings.push({ y: my, r: mr, minds: true });
      }
      const maxW = Math.max(1, ...m.tree.map(t => t.weight));
      st.edges = m.tree.filter(t => st.nodes.has(t.up) && st.nodes.has(t.down)).map((t, i) => ({ ...t, k: t.weight ? Math.log2(1 + t.weight) / Math.log2(1 + maxW) : 0, seed: (i * 0.6180339) % 1 }));
      st.wears = m.wears.filter(w => st.nodes.has(w.from) && st.nodes.has(w.to));
      st.dist0 = R * 2.75 + 2;
      if (!st.focus) { st.goal.dist = st.dist0; st.dist = st.dist0; }
      palette();
      kick();
    }
    // a bond curves: it leaves its parent level, then falls to the child (a bezier, sampled)
    function bez(A, B, t) {
      const C = [A[0] * 0.35 + B[0] * 0.65, A[1] - (A[1] - B[1]) * 0.12, A[2] * 0.35 + B[2] * 0.65];
      const u = 1 - t;
      return [u * u * A[0] + 2 * u * t * C[0] + t * t * B[0], u * u * A[1] + 2 * u * t * C[1] + t * t * B[1], u * u * A[2] + 2 * u * t * C[2] + t * t * B[2]];
    }

    // ---------- the theme ----------
    function palette() {
      const probes = opts.probes;
      const want = new Set(['ink', 'muted', 'line', 'surface', 'ok', 'warn', 'danger', 'accent']);
      const ranks = new Set();
      for (const q of st.nodes.values()) ranks.add(q.rank);
      probes.replaceChildren();
      const read = (attr, v) => { const sp = document.createElement('span'); sp.className = 'net-probe'; sp.setAttribute(attr, v); probes.append(sp); return sp; };
      const els = {};
      for (const t of want) els['t:' + t] = read('data-tone', t);
      for (const r of ranks) els['r:' + r] = read('data-rank', r);
      const pal = { rank: {} };
      for (const t of want) pal[t] = rgb(getComputedStyle(els['t:' + t]).color);
      for (const r of ranks) pal.rank[r] = rgb(getComputedStyle(els['r:' + r]).color);
      const s = pal.surface;
      st.dark = 0.2126 * s[0] + 0.7152 * s[1] + 0.0722 * s[2] < 0.5;
      st.pal = pal;
      kick();
    }
    const mo = new MutationObserver(() => palette());
    mo.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme', 'data-palette'] });
    const mq = matchMedia('(prefers-color-scheme: dark)');
    const onScheme = () => palette();
    mq.addEventListener('change', onScheme);

    // ---------- the camera ----------
    function camera() {
      const t = st.target, cp = Math.cos(st.pitch);
      const eye = [t[0] + st.dist * cp * Math.sin(st.yaw), t[1] + st.dist * Math.sin(st.pitch), t[2] + st.dist * cp * Math.cos(st.yaw)];
      const proj = perspective(0.72, st.W / st.H, 0.1, 200);
      st.VP = mul(proj, lookAt(eye, t, [0, 1, 0]));
      st.scale = st.H * st.dpr / (2 * Math.tan(0.36));
    }
    function project(p) {
      const M = st.VP;
      const x = M[0] * p[0] + M[4] * p[1] + M[8] * p[2] + M[12];
      const y = M[1] * p[0] + M[5] * p[1] + M[9] * p[2] + M[13];
      const w = M[3] * p[0] + M[7] * p[1] + M[11] * p[2] + M[15];
      return { x: (x / w + 1) / 2 * st.W, y: (1 - y / w) / 2 * st.H, w };
    }

    // ---------- a frame ----------
    function frame(now) {
      st.raf = 0;
      if (!st.run) return;
      const reduced = opts.reduced();
      const busy = st.drag || st.pulses.length || st.flashes.size || Math.abs(st.dist - st.goal.dist) > 0.01 || Math.hypot(st.target[0] - st.goal.target[0], st.target[1] - st.goal.target[1], st.target[2] - st.goal.target[2]) > 0.01;
      const ambient = !reduced;
      // the budget: 60 a second while something moves for a reason, 30 while it only breathes
      const budget = busy ? 15 : 32;
      if (st.last && now - st.last < budget) { schedule(); return; }
      const dt = st.last ? Math.min(0.1, (now - st.last) / 1000) : 0.016;
      const spent = performance.now();
      st.last = now;
      st.t += dt;
      if (ambient && !st.drag && !st.focus && now - st.idleAt > 3500) st.yaw += dt * 0.07;
      const k = 1 - Math.pow(0.0015, dt);
      st.dist += (st.goal.dist - st.dist) * k;
      for (let i = 0; i < 3; i++) st.target[i] += (st.goal.target[i] - st.target[i]) * k;
      render(now, reduced);
      st.slow = st.slow * 0.9 + (performance.now() - spent) * 0.1;
      if (ambient || busy) schedule();
    }
    function schedule() { if (st.run && !st.raf) st.raf = requestAnimationFrame(frame); }
    function kick() { if (st.run) { if (!st.raf) st.raf = requestAnimationFrame(frame); } }

    function render(now, reduced) {
      if (!st.pal || !st.m) return;
      camera();
      const pal = st.pal, dark = st.dark, lit = st.lit, focus = st.shown || '';
      const cw = Math.round(st.W * st.dpr), ch = Math.round(st.H * st.dpr);
      gl.viewport(0, 0, cw, ch);
      gl.clearColor(0, 0, 0, 0);
      gl.clear(gl.COLOR_BUFFER_BIT);
      gl.enable(gl.BLEND);
      if (dark) gl.blendFunc(gl.ONE, gl.ONE); else gl.blendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA);
      const dim = id => focus && lit && !lit.nodes.has(id);
      const col = r => pal.rank[r] || pal.muted;
      // ribbons
      const RB = [];
      const seg = (a, b, c, alpha, w) => RB.push(a[0], a[1], a[2], b[0], b[1], b[2], c[0], c[1], c[2], alpha, w * st.dpr);
      for (const r of st.rings) {
        const n = 72;
        for (let i = 0; i < n; i++) {
          const a0 = i / n * TAU, a1 = (i + 1) / n * TAU;
          seg([Math.cos(a0) * r.r, r.y, Math.sin(a0) * r.r], [Math.cos(a1) * r.r, r.y, Math.sin(a1) * r.r], r.minds ? pal.accent : pal.muted, (dark ? 0.16 : 0.22) * (focus ? 0.4 : 1), 1);
        }
      }
      for (const w of st.wears) {
        const on = focus && lit && lit.wears.has(w.from + '>' + w.to);
        seg(st.nodes.get(w.from).p, st.nodes.get(w.to).p, pal.accent, on ? 0.55 : focus ? 0.02 : (dark ? 0.05 : 0.07), on ? 1.4 : 1);
      }
      for (const e of st.edges) {
        const A = st.nodes.get(e.up).p, B = st.nodes.get(e.down).p, c = col(st.nodes.get(e.down).rank);
        const on = focus && lit && lit.tree.has(e.up + '>' + e.down);
        const liveE = st.live.has(e.down);
        const base = (dark ? 0.42 : 0.55) + 0.4 * e.k;
        const alpha = focus ? (on ? 0.95 : 0.05) : liveE ? 0.95 : base * (e.implied ? 0.5 : 1);
        const w = (1.3 + 3 * e.k) * (on || liveE ? 1.5 : 1);
        let prev = A;
        for (let i = 1; i <= 12; i++) { const q = bez(A, B, i / 12); seg(prev, q, c, e.implied && i % 2 ? alpha * 0.3 : alpha, w); prev = q; }
      }
      gl.useProgram(LP);
      gl.uniformMatrix4fv(gl.getUniformLocation(LP, 'uVP'), false, st.VP);
      gl.uniform2f(gl.getUniformLocation(LP, 'uView'), cw, ch);
      gl.bindBuffer(gl.ARRAY_BUFFER, lbuf);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(RB), gl.DYNAMIC_DRAW);
      gl.bindVertexArray(vaoL);
      gl.drawArraysInstanced(gl.TRIANGLES, 0, 6, RB.length / 11);
      // sprites: halos, cores, particles, sparks — back to front
      const S = [];
      const put = (p, c, a, size, core) => S.push(p[0], p[1], p[2], c[0], c[1], c[2], a, size, core);
      const items = [];
      const breath = reduced ? 1 : 0.75 + 0.25 * Math.sin(st.t * 2.4);
      for (const q of st.nodes.values()) {
        const c = col(q.rank), d = dim(q.id), on = q.id === focus;
        const liveN = st.live.has(q.id);
        const a = d ? 0.12 : 1;
        items.push({ z: project(q.p).w, draw: () => {
          put(q.p, c, (dark ? 0.42 : 0.3) * a * (liveN ? 1.6 * breath : on ? 1.4 : 1), q.size * (liveN ? 4.2 : 3.2), 0);
          put(q.p, c, (dark ? 0.95 : 1) * a, q.size, q.n.kind === 'mind' ? 0.5 : 0.62);
        } });
      }
      // particles drift along the bonds, the way work goes: down
      const thin = st.slow > 12 ? 2 : 1;
      if (!reduced) for (const e of st.edges) {
        if (focus && lit && !lit.tree.has(e.up + '>' + e.down)) continue;
        const n = (1 + Math.min(3, Math.floor(Math.log2(1 + e.weight)))) / thin;
        const c = col(st.nodes.get(e.down).rank), A = st.nodes.get(e.up).p, B = st.nodes.get(e.down).p;
        const speed = 0.16 + 0.12 * e.k + (st.live.has(e.down) ? 0.3 : 0);
        for (let j = 0; j < n; j++) {
          const u = (st.t * speed + j / n + e.seed) % 1;
          const fade = Math.sin(u * Math.PI);
          const p = bez(A, B, u);
          items.push({ z: project(p).w, draw: () => put(p, c, (dark ? 0.9 : 0.75) * fade, 0.13, 0.35) });
        }
      }
      // live sparks: a dispatch down, a report up
      st.pulses = st.pulses.filter(pl => {
        const t = (now - pl.at) / (pl.per * pl.hops.length);
        if (t >= 1) { flash(pl.end, pl.tone); return false; }
        const h = Math.min(pl.hops.length - 1, Math.floor(t * pl.hops.length));
        const lt = t * pl.hops.length - h;
        const [up, down] = pl.hops[h];
        const A = st.nodes.get(up).p, B = st.nodes.get(down).p;
        const c = pl.tone === 'down' ? col(st.nodes.get(pl.end).rank) : pal[pl.tone] || pal.ok;
        for (let i = 0; i < 6; i++) {
          let u = lt - i * 0.035;
          if (u < 0) continue;
          u = u < 0.5 ? 2 * u * u : 1 - Math.pow(-2 * u + 2, 2) / 2;
          const p = bez(A, B, pl.dir === 'up' ? 1 - u : u);
          items.push({ z: -1e9, draw: () => { put(p, c, 1 - i * 0.15, i ? 0.22 - i * 0.025 : 0.95, i ? 0.4 : 0); if (!i) put(p, c, 1, 0.24, 0.6); } });
        }
        return true;
      });
      for (const [id, f] of st.flashes) {
        const t = (now - f.at) / 1100;
        if (t >= 1) { st.flashes.delete(id); continue; }
        const q = st.nodes.get(id);
        if (!q) continue;
        items.push({ z: -1e9, draw: () => put(q.p, pal[f.tone] || col(q.rank), (1 - t) * 0.9, q.size * (2 + 5 * t), 0) });
      }
      items.sort((a, b) => b.z - a.z);
      for (const it of items) it.draw();
      gl.useProgram(P);
      gl.uniformMatrix4fv(gl.getUniformLocation(P, 'uVP'), false, st.VP);
      gl.uniform1f(gl.getUniformLocation(P, 'uScale'), st.scale);
      gl.bindBuffer(gl.ARRAY_BUFFER, pbuf);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(S), gl.DYNAMIC_DRAW);
      gl.bindVertexArray(vaoP);
      gl.drawArrays(gl.POINTS, 0, S.length / 9);
      gl.bindVertexArray(null);
      labels(focus, lit);
    }

    // ---------- labels: the upper layers always, the rest on hover, focus or while live ----------
    function labels(focus, lit) {
      const ctx = octx, W = st.W, H = st.H, dpr = st.dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      ctx.clearRect(0, 0, W, H);
      const cs = getComputedStyle(stage);
      const fam = cs.getPropertyValue('--font-body') || 'system-ui';
      const px = parseFloat(getComputedStyle(document.documentElement).fontSize) || 16;
      const ink = 'rgb(' + st.pal.ink.map(v => Math.round(v * 255)).join(',') + ')';
      const halo = 'rgb(' + st.pal.surface.map(v => Math.round(v * 255)).join(',') + ')';
      const many = st.nodes.size > 40;
      const shown = [];
      for (const q of st.nodes.values()) {
        const on = q.id === focus || (focus && lit && lit.nodes.has(q.id));
        const always = q.d === 0 || q.d === 1 || (q.d === 2 && !many);
        if (!(on || st.live.has(q.id) || (!focus && always))) continue;
        const s = project(q.p);
        if (s.w <= 0 || s.x < -40 || s.x > W + 40 || s.y < -20 || s.y > H + 20) continue;
        shown.push({ q, s });
      }
      shown.sort((a, b) => b.s.w - a.s.w);
      const taken = [];
      for (const { q, s } of shown) {
        const n = q.n, big = q.d === 0;
        const text = n.throne ? (n.label || n.name) : n.kind === 'mind' ? n.name : (n.rank && n.name.startsWith(n.rank + '-') ? n.name.slice(n.rank.length + 1) : n.name);
        ctx.font = (big ? '650 ' + px * 1.05 : (q.d === 1 ? '600 ' : '') + px * (q.d < 0 ? 0.72 : 0.8)) + 'px ' + fam;
        const w = ctx.measureText(text).width;
        const r = q.size * st.scale / s.w / dpr * 0.5;
        const x = s.x + r + 6, y = s.y;
        const box = [x - 2, y - px * 0.6, x + w + 2, y + px * 0.6];
        if (q.id !== focus && taken.some(b => !(box[2] < b[0] || box[0] > b[2] || box[3] < b[1] || box[1] > b[3]))) continue;
        taken.push(box);
        const near = clamp(1.4 - (s.w - st.dist * 0.75) / (st.dist0 * 0.9), 0.45, 1);
        ctx.globalAlpha = q.id === focus ? 1 : near;
        ctx.textBaseline = 'middle';
        ctx.lineJoin = 'round';
        ctx.lineWidth = 3.5;
        ctx.strokeStyle = halo;
        ctx.strokeText(text, x, y);
        ctx.fillStyle = ink;
        ctx.fillText(text, x, y);
      }
      ctx.globalAlpha = 1;
    }

    // ---------- picking and the pointer ----------
    function pick(x, y) {
      if (!st.VP) return '';
      let best = '', bd = Infinity;
      for (const q of st.nodes.values()) {
        const s = project(q.p);
        if (s.w <= 0) continue;
        const r = Math.max(10, q.size * st.scale / s.w / st.dpr * 0.7);
        const d = Math.hypot(s.x - x, s.y - y);
        if (d < r && d < bd) { bd = d; best = q.id; }
      }
      return best;
    }
    const ptrs = new Map();
    let down = null, pinch = 0;
    const local = ev => { const r = canvas.getBoundingClientRect(); return [ev.clientX - r.left, ev.clientY - r.top]; };
    canvas.addEventListener('pointerdown', ev => {
      canvas.setPointerCapture(ev.pointerId);
      ptrs.set(ev.pointerId, local(ev));
      down = { at: local(ev), moved: false, yaw: st.yaw, pitch: st.pitch };
      if (ptrs.size === 2) { const [a, b] = [...ptrs.values()]; pinch = Math.hypot(a[0] - b[0], a[1] - b[1]); }
      st.idleAt = performance.now();
    });
    canvas.addEventListener('pointermove', ev => {
      const p = local(ev);
      if (ptrs.has(ev.pointerId)) ptrs.set(ev.pointerId, p);
      if (ptrs.size === 2 && pinch) {
        const [a, b] = [...ptrs.values()], d = Math.hypot(a[0] - b[0], a[1] - b[1]);
        st.goal.dist = clamp(st.goal.dist * pinch / Math.max(1, d), 2, st.dist0 * 2.2); st.dist = st.goal.dist; pinch = d;
        kick(); return;
      }
      if (down && ptrs.size === 1) {
        const dx = p[0] - down.at[0], dy = p[1] - down.at[1];
        if (!down.moved && Math.hypot(dx, dy) > 4) { down.moved = true; st.drag = true; canvas.setAttribute('data-drag', ''); }
        if (down.moved) { st.yaw = down.yaw - dx * 0.007; st.pitch = clamp(down.pitch + dy * 0.005, -1.25, 1.25); st.idleAt = performance.now(); kick(); }
        return;
      }
      const id = pick(p[0], p[1]);
      canvas.toggleAttribute('data-hot', !!id);
      if (id !== st.hover) { st.hover = id; opts.onHover(id); kick(); }
    });
    const end = ev => {
      ptrs.delete(ev.pointerId);
      if (ptrs.size < 2) pinch = 0;
      if (down && !down.moved && ev.type === 'pointerup') opts.onPick(pick(...local(ev)));
      if (!ptrs.size) { down = null; st.drag = false; canvas.removeAttribute('data-drag'); }
      st.idleAt = performance.now();
    };
    canvas.addEventListener('pointerup', end);
    canvas.addEventListener('pointercancel', end);
    canvas.addEventListener('pointerleave', () => { if (st.hover) { st.hover = ''; opts.onHover(''); } canvas.removeAttribute('data-hot'); });
    canvas.addEventListener('wheel', ev => {
      ev.preventDefault();
      st.goal.dist = clamp(st.goal.dist * Math.exp(ev.deltaY * 0.0012), 2, st.dist0 * 2.2);
      st.idleAt = performance.now(); kick();
    }, { passive: false });
    canvas.addEventListener('keydown', ev => {
      const turn = { ArrowLeft: -0.2, ArrowRight: 0.2 }[ev.key], tilt = { ArrowUp: -0.12, ArrowDown: 0.12 }[ev.key];
      if (turn) st.yaw += turn;
      else if (tilt) st.pitch = clamp(st.pitch + tilt, -1.25, 1.25);
      else if (ev.key === '+' || ev.key === '=') st.goal.dist = clamp(st.goal.dist * 0.85, 2, st.dist0 * 2.2);
      else if (ev.key === '-') st.goal.dist = clamp(st.goal.dist / 0.85, 2, st.dist0 * 2.2);
      else return;
      ev.preventDefault(); st.idleAt = performance.now(); kick();
      if (!st.run || opts.reduced()) render(performance.now(), true);
    });
    canvas.addEventListener('webglcontextlost', ev => { ev.preventDefault(); st.run = false; });
    canvas.addEventListener('webglcontextrestored', () => { init(); st.run = true; kick(); });

    function resize() {
      const r = stage.getBoundingClientRect();
      st.W = Math.max(1, r.width); st.H = Math.max(1, r.height);
      st.dpr = Math.min(2, devicePixelRatio || 1);
      for (const c of [canvas, overlay]) { c.width = Math.round(st.W * st.dpr); c.height = Math.round(st.H * st.dpr); }
      if (st.m) { st.last = 0; render(performance.now(), opts.reduced()); }
      kick();
    }
    const ro = new ResizeObserver(resize);
    ro.observe(stage);

    function flash(id, tone) { st.flashes.set(id, { at: performance.now(), tone }); kick(); }

    return {
      elements: [canvas, overlay],
      model(m) { scene(m); resize(); },
      focus(id, lit, pinned) {
        st.lit = lit; st.shown = id || '';
        // the camera eases to a pinned neuron; a hover only lights
        if ((pinned || '') !== st.focus) {
          st.focus = pinned || '';
          const q = st.nodes.get(st.focus);
          st.goal.target = q ? q.p.slice() : [0, 0, 0];
          st.goal.dist = q ? Math.max(3.5, st.dist0 * 0.62) : st.dist0;
        }
        kick();
        if (opts.reduced()) render(performance.now(), true);
      },
      live(ids) { st.live = ids; kick(); },
      signal(kind, hops, id, tone) {
        const pairs = hops.map(h => h.split('>')).filter(([a, b]) => st.nodes.has(a) && st.nodes.has(b));
        const endAt = kind === 'up' ? (pairs.length ? pairs[0][0] : id) : id;
        const t = kind === 'up' ? (tone || 'ok') : 'down';
        if (opts.reduced() || !pairs.length) { flash(endAt, t); return; }
        st.pulses.push({ hops: kind === 'up' ? pairs.slice().reverse() : pairs, dir: kind, at: performance.now(), per: 700, end: endAt, tone: t });
        kick();
      },
      running(v) {
        st.run = v;
        if (v) { st.last = 0; resize(); kick(); }
        else if (st.raf) { cancelAnimationFrame(st.raf); st.raf = 0; }
      },
      resize,
      dispose() {
        st.run = false;
        if (st.raf) cancelAnimationFrame(st.raf);
        ro.disconnect(); mo.disconnect(); mq.removeEventListener('change', onScheme);
        const lose = gl.getExtension('WEBGL_lose_context');
        if (lose) lose.loseContext();
      },
    };
  }

  window.isekaiNet3D = { supported, create };
})();
