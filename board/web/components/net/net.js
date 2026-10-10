// net.js — the world as a neural net. The hierarchy is drawn as layers (the throne, then each rank
// in the order work flows down, the minds beside), each neuron a rank-coloured tile whose cells are
// its last four weeks of work (TensorFlow-playground style: the neuron shows its own output), each
// tree bond weighted by the commissions it carried. Hover or focus lights a neuron's path to the
// throne and its subtree; a Court's dispatch travels down the tree as a pulse and its report comes
// back up. The 3D stage (net3d.js) draws the same model. No dependencies; colours come from the css
// tokens only (the svg reads them through classes; the canvas resolves them from probes).
(() => {
  'use strict';
  const NS = 'http://www.w3.org/2000/svg';
  function s(tag, attrs, ...kids) {
    const n = document.createElementNS(NS, tag);
    for (const [k, v] of Object.entries(attrs || {})) if (v !== false && v !== null && v !== undefined) n.setAttribute(k, v === true ? '' : v);
    for (const kid of kids.flat()) if (kid !== null && kid !== undefined && kid !== false) n.append(typeof kid === 'string' ? document.createTextNode(kid) : kid);
    return n;
  }
  const PORTRAITS = new Set(['rimuru', 'elf', 'orc', 'slime', 'kijin', 'darkelf', 'highelf', 'highorc']);
  const BRAIN = 'M0 -6 C-3.2 -6 -5.6 -4.4 -5.6 -1.6 C-7.2 -0.6 -7.2 1.8 -5.4 2.8 C-6 4.6 -4.2 6.2 -2.2 6.2 C-1.1 6.2 -0.4 5.4 0 4.4 C0.4 5.4 1.1 6.2 2.2 6.2 C4.2 6.2 6 4.6 5.4 2.8 C7.2 1.8 7.2 -0.6 5.6 -1.6 C5.6 -4.4 3.2 -6 0 -6 Z';
  const UP_BONDS = ['truth', 'verdict', 'reports', 'above'];
  const RANK_ORDER = ['rimuru', 'elf', 'highelf', 'darkelf', 'kijin', 'orc', 'highorc', 'slime'];
  const clamp = (v, a, b) => Math.max(a, Math.min(b, v));
  const reducedQ = matchMedia('(prefers-reduced-motion: reduce)');

  // ---------- the model: one tree under the throne, the minds beside it ----------
  // colony: the /api/colony payload as dash.js merged it (bodies and skills joined). activity: the
  // /api/activity payload, or null when the instruments are silent (tiles stay blank, honestly).
  function build(colony, activity, word) {
    const all = colony.nodes || [];
    const byId = new Map(all.map(n => [n.id, n]));
    let throne = byId.get(colony.throne) || all.find(n => n.throne) || null;
    const isMind = n => n.kind === 'mind';
    const creatures = all.filter(n => !isMind(n));
    const minds = all.filter(isMind);
    const kids = new Map(), parent = new Map();
    const edgeOf = new Map(); // child id → the colony edge that ties it to its parent
    for (const e of colony.edges || []) if (UP_BONDS.includes(e.bond) && byId.has(e.to)) {
      const have = edgeOf.get(e.from);
      if (!have || UP_BONDS.indexOf(e.bond) < UP_BONDS.indexOf(have.bond)) edgeOf.set(e.from, e);
    }
    for (const n of creatures) {
      if (n === throne) continue;
      let p = n.parent && byId.has(n.parent) && n.parent !== n.id ? n.parent : (edgeOf.get(n.id) || {}).to;
      if (!p || !byId.has(p)) p = throne ? throne.id : '';
      if (!p) continue;
      parent.set(n.id, p);
      if (!kids.has(p)) kids.set(p, []);
      kids.get(p).push(n.id);
    }
    const rankKey = n => { const i = RANK_ORDER.indexOf(n.rank); return i < 0 ? RANK_ORDER.length : i; };
    for (const list of kids.values()) list.sort((a, b) => rankKey(byId.get(a)) - rankKey(byId.get(b)) || byId.get(a).name.localeCompare(byId.get(b).name));
    // depth by walking down from the throne; anything the walk misses (a loop) hangs at depth 1
    const depth = new Map();
    const order = []; // creatures in preorder
    const walk = (id, d, seen) => {
      if (seen.has(id)) return;
      seen.add(id); depth.set(id, d); order.push(id);
      for (const k of kids.get(id) || []) walk(k, d + 1, seen);
    };
    const seen = new Set();
    if (throne) walk(throne.id, 0, seen);
    for (const n of creatures) if (!seen.has(n.id)) walk(n.id, throne ? 1 : 0, seen);
    const act = (activity && activity.bodies) || {};
    const tree = [];
    for (const [c, p] of parent) {
      const e = edgeOf.get(c);
      const bond = e && e.to === p ? e.bond : 'reports';
      const a = act[byId.get(c).name];
      tree.push({ up: p, down: c, bond, implied: !(e && e.to === p) || !!(e && e.implied), weight: a ? a.total : 0 });
    }
    const wears = [];
    for (const e of colony.edges || []) if (e.bond === 'wears' && byId.has(e.from) && byId.has(e.to) && isMind(byId.get(e.to))) wears.push({ from: e.from, to: e.to });
    let maxDepth = 0;
    for (const d of depth.values()) maxDepth = Math.max(maxDepth, d);
    // per-node activity: one value a day (tokens when the usage journal spoke, else runs)
    let peak = 0;
    const cells = new Map();
    for (const n of all) {
      const a = act[n.name];
      if (!a) continue;
      const v = (a.tokens || []).some(x => x > 0) ? a.tokens : a.runs || [];
      cells.set(n.id, v);
      for (const x of v) peak = Math.max(peak, x);
    }
    return { byId, throne, creatures, minds, kids, parent, depth, order, tree, wears, maxDepth, cells, peak, act, days: (activity && activity.days) || [], word };
  }

  // ancestors (the path to the throne) and descendants (the subtree) of a node
  function lineage(m, id) {
    const up = [], down = [];
    for (let p = m.parent.get(id), guard = 0; p && guard < 64; p = m.parent.get(p), guard++) up.push(p);
    const stack = [...(m.kids.get(id) || [])];
    while (stack.length) { const k = stack.pop(); down.push(k); stack.push(...(m.kids.get(k) || [])); }
    return { up, down };
  }
  // the lit set of a focus: path to the throne, subtree, worn minds (or a mind's wearers)
  function litSet(m, id) {
    const nodes = new Set([id]), tree = new Set(), wears = new Set();
    const n = m.byId.get(id);
    if (!n) return { nodes, tree, wears };
    if (n.kind === 'mind') {
      for (const w of m.wears) if (w.to === id) { nodes.add(w.from); wears.add(w.from + '>' + w.to); }
      return { nodes, tree, wears };
    }
    const { up, down } = lineage(m, id);
    let c = id;
    for (const p of up) { nodes.add(p); tree.add(p + '>' + c); c = p; }
    for (const d of down) { nodes.add(d); tree.add(m.parent.get(d) + '>' + d); }
    for (const w of m.wears) if (w.from === id) { nodes.add(w.to); wears.add(w.from + '>' + w.to); }
    return { nodes, tree, wears };
  }
  // the hops a dispatch travels: from the nearest running ancestor (else the throne) down to id
  function route(m, id, live) {
    const { up } = lineage(m, id);
    let stop = up.length;
    for (let i = 0; i < up.length; i++) if (live.has(m.byId.get(up[i]).name)) { stop = i + 1; break; }
    const chain = [id, ...up.slice(0, stop)].reverse(); // dispatcher … id
    const hops = [];
    for (let i = 0; i + 1 < chain.length; i++) hops.push(chain[i] + '>' + chain[i + 1]);
    return hops;
  }

  // ---------- text measure (labels decide the column widths) ----------
  const measureCtx = document.createElement('canvas').getContext('2d');
  function measure(text, font) { measureCtx.font = font; return measureCtx.measureText(text).width; }
  function fit(text, font, max) {
    if (measure(text, font) <= max) return text;
    let lo = 1, hi = text.length;
    while (lo < hi) { const mid = (lo + hi + 1) >> 1; if (measure(text.slice(0, mid) + '…', font) <= max) lo = mid; else hi = mid - 1; }
    return text.slice(0, lo) + '…';
  }

  // ---------- the 2D layout ----------
  // wide: TF-playground columns, one per layer, the tree tidy (a parent sits at the middle of its
  // children); narrow: an outline, one row per neuron, indented by depth — the tree still reads.
  function layout2d(m, W, availH, fonts) {
    const narrow = W < 640;
    const label = n => n.throne ? (n.label || n.name) : (n.rank && n.name.startsWith(n.rank + '-') ? n.name.slice(n.rank.length + 1) : n.name);
    const pos = new Map();
    const cols = [];
    const T = narrow ? 24 : 26, TT = narrow ? 32 : 38, TM = 18;
    if (narrow) {
      const G = 40, indent = clamp((W - 190) / Math.max(1, m.maxDepth), 18, 34), top = 16;
      let y = top + TT / 2;
      for (const id of m.order) {
        const n = m.byId.get(id), d = m.depth.get(id) || 0, size = n.throne ? TT : T;
        const x = 12 + size / 2 + d * indent;
        const room = W - (x + size / 2 + 8) - 8;
        pos.set(id, { x, y, size, label: fit(label(n), fonts.label, room), lx: x + size / 2 + 8, d });
        y += n.throne ? G + 6 : G;
      }
      let H = y;
      if (m.minds.length) {
        cols.push({ kind: 'minds', x: 4, y: H - 8, w: W - 8, h: 0, title: m.word('minds', 'minds'), count: m.minds.length, rank: 'mind:shared' });
        let my = H + 24;
        for (const n of mindOrder(m, pos)) {
          const x = 12 + TM / 2;
          pos.set(n.id, { x, y: my, size: TM, label: fit(n.name, fonts.small, W - x - 40), lx: x + TM / 2 + 8, d: -1 });
          my += 30;
        }
        cols[0].h = my - cols[0].y;
        H = my;
      }
      return { pos, cols, W, H: H + 8, narrow };
    }
    // columns: the tidy tree's leaves take consecutive slots; a parent sits at its children's middle
    const leaves = [];
    const slot = new Map();
    const place = id => {
      const ks = m.kids.get(id) || [];
      if (!ks.length) { slot.set(id, leaves.length); leaves.push(id); return; }
      ks.forEach(place);
      slot.set(id, ks.reduce((a, k) => a + slot.get(k), 0) / ks.length);
    };
    if (m.throne) place(m.throne.id);
    for (const id of m.order) if (!slot.has(id)) place(id);
    const D = m.maxDepth + 1;
    const top = 44;
    // the deepest layer wraps into k sub-columns when one column would not fit the height given
    const room = Math.max(160, availH - top - 16), Gmin = leaves.length > 40 ? 24 : 30;
    const k = clamp(Math.ceil(leaves.length * Gmin / room), 1, 4);
    const rows = Math.max(1, Math.ceil(leaves.length / k), Math.ceil(m.minds.length * 22 / Math.max(Gmin, 1) / 1.4));
    const G = clamp(room / rows, Gmin, 42);
    const H = top + rows * G + 16;
    const deep = d => k > 1 && d === m.maxDepth;
    // column widths: the tile, its label, and room for the bonds between
    const want = [];
    for (let d = 0; d < D; d++) want.push(0);
    for (const id of m.order) {
      const n = m.byId.get(id), d = m.depth.get(id) || 0;
      want[d] = Math.max(want[d], (n.throne ? TT : T) + 8 + measure(label(n), n.throne ? fonts.bold : fonts.label));
    }
    const sub = want[D - 1], subGap = 22;
    if (k > 1) want[D - 1] = sub * k + subGap * (k - 1);
    let mindW = 0;
    for (const n of m.minds) mindW = Math.max(mindW, TM + 8 + measure(n.name, fonts.small) + 26);
    const nCols = D + (m.minds.length ? 1 : 0);
    const minGap = k > 1 ? 30 : 44;
    const widths = [...want, ...(m.minds.length ? [mindW] : [])];
    let total = widths.reduce((a, b) => a + b, 0) + minGap * (nCols - 1) + 24;
    // too wide: the minds' labels give way first (they are the side layer), then the widest
    if (m.minds.length) { const cut = Math.min(Math.max(0, total - W), Math.max(0, widths[D] - 130)); widths[D] -= cut; total -= cut; }
    let shave = 0;
    while (total > W && shave < 400) { const i = widths.indexOf(Math.max(...widths)); widths[i] -= 8; total -= 8; shave += 8; }
    const subW = k > 1 ? (widths[D - 1] - subGap * (k - 1)) / k : widths[D - 1];
    const gap = minGap + Math.max(0, (W - total) / Math.max(1, nCols - 1)) * 0.92;
    let x = 12;
    const colX = widths.map(w => { const at = x; x += w + gap; return at; });
    // leaves by slot (row-major across the sub-columns), parents at the middle of their children
    const ys = new Map();
    const offY = (rows - Math.ceil(leaves.length / k)) * G / 2;
    // staggered: sub-column j sits j/k of a row lower, so a bond to it passes between the labels
    leaves.forEach((id, i) => ys.set(id, top + offY + G / 2 + i * G / k));
    const yOf = id => {
      if (ys.has(id)) return ys.get(id);
      const ks = m.kids.get(id) || [];
      const y = ks.length ? ks.reduce((a, c) => a + yOf(c), 0) / ks.length : top + G / 2;
      ys.set(id, y);
      return y;
    };
    for (let d = 0; d < D; d++) {
      const ids = m.order.filter(id => (m.depth.get(id) || 0) === d).sort((a, b) => slot.get(a) - slot.get(b));
      const ranks = [];
      for (const id of ids) { const r = m.byId.get(id).rank || ''; if (!ranks.includes(r)) ranks.push(r); }
      const title = ranks.map(r => m.word('rank.' + r, r)).join(' · ');
      cols.push({ kind: 'rank', x: colX[d] - 6, y: 8, w: widths[d] + 12, h: H - 12, title, count: ids.length, rank: ranks[0] || '' });
      let last = -Infinity;
      for (const id of ids) {
        const n = m.byId.get(id), size = n.throne ? TT : T;
        let y = yOf(id), x0 = colX[d] + size / 2, roomL = widths[d] - size - 8;
        if (deep(d) && slot.has(id) && Number.isInteger(slot.get(id))) {
          x0 = colX[d] + (leaves.indexOf(id) % k) * (subW + subGap) + size / 2;
          roomL = subW - size - 8;
        } else {
          if (y - last < size + 6) y = last + size + 6;
          last = y;
        }
        pos.set(id, { x: x0, y, size, label: fit(label(n), n.throne ? fonts.bold : fonts.label, roomL), lx: x0 + size / 2 + 8, d });
      }
    }
    if (m.minds.length) {
      const cx = colX[D];
      cols.push({ kind: 'minds', x: cx - 6, y: 8, w: widths[D] + 12, h: H - 12, title: m.word('minds', 'minds'), count: m.minds.length, rank: 'mind:shared' });
      const list = mindOrder(m, pos);
      const MG = clamp((H - top - 16) / list.length, 22, 40);
      list.forEach((n, i) => {
        const y = top + (i + 0.5) * MG + (H - top - 16 - MG * list.length) / 2;
        pos.set(n.id, { x: cx + TM / 2, y, size: TM, label: fit(n.name, fonts.small, widths[D] - TM - 34), lx: cx + TM + 8, d: -1 });
      });
    }
    // where a neuron's bonds leave it: past its label, so the label rides the axon
    for (const [id, p] of pos) {
      const n = m.byId.get(id);
      p.out = p.lx + measure(p.label, n.throne ? fonts.bold : (n.kind === 'mind' ? fonts.small : fonts.label)) + 6;
    }
    return { pos, cols, W, H, narrow };
  }
  // minds in the order of their wearers (the barycentre), so anima-threads cross as little as they can
  function mindOrder(m, pos) {
    const at = n => {
      const ys = m.wears.filter(w => w.to === n.id && pos.has(w.from)).map(w => pos.get(w.from).y);
      return ys.length ? ys.reduce((a, b) => a + b, 0) / ys.length : Infinity;
    };
    return [...m.minds].map(n => [n, at(n)]).sort((a, b) => a[1] - b[1] || a[0].name.localeCompare(b[0].name)).map(x => x[0]);
  }

  // ---------- the view ----------
  function create(host, opts) {
    const word = opts.word || ((k, d) => d || k);
    const st = {
      m: null, L: null, focus: '', hover: '', live: new Set(), mode: 'auto', visible: true, three: null,
      pulses: [], raf: 0, edgeEls: new Map(), nodeEls: new Map(), current: '',
    };
    host.classList.add('net');
    const stage = document.createElement('div');
    stage.className = 'net-stage';
    stage.dataset.mode = '2d';
    host.replaceChildren(stage);
    // probes: the canvas reads the theme's colours through these (never a literal)
    const probes = document.createElement('div');
    probes.setAttribute('aria-hidden', 'true');
    host.append(probes);
    const hint = document.createElement('p');
    hint.className = 'visually-hidden';
    hint.id = 'net-keys';
    hint.textContent = 'Arrow keys walk the net: left to the parent, right to the first child, up and down within a layer. Enter pins a neuron, Escape clears.';
    host.append(hint);

    // the label fonts as the css resolves them (the svg text and the measure must agree)
    const fonts = () => {
      const probe = document.createElement('span');
      probe.className = 'net-probe';
      host.append(probe);
      const at = size => { probe.style.font = size + ' var(--font-body)'; const cs = getComputedStyle(probe); return cs.fontSize + ' ' + cs.fontFamily; };
      const out = { label: at('var(--step--1)'), small: at('var(--step--2)'), bold: '650 ' + at('var(--step--1)') };
      probe.remove();
      return out;
    };
    function draw() {
      if (!st.m || !host.isConnected || host.offsetParent === null) return;
      const fit3 = opts.fit && opts.fit();
      if (fit3) stage.style.setProperty('--net-h', Math.max(320, fit3) + 'px'); else stage.style.removeProperty('--net-h');
      if (st.mode === '3d' && st.three) { stage.dataset.mode = '3d'; st.three.resize(); return; }
      stage.dataset.mode = '2d';
      const W = Math.max(300, stage.clientWidth || host.clientWidth);
      const fitH = opts.fit && opts.fit();
      const availH = fitH ? Math.max(360, fitH) : Math.max(560, innerHeight - 190);
      const f = fonts();
      const L = layout2d(st.m, W, availH, f);
      st.L = L;
      const m = st.m;
      const g = s('svg', { viewBox: `0 0 ${W} ${L.H}`, role: 'group', 'aria-label': 'the ' + word('world', 'world') + ' as a neural net — ' + m.creatures.length + ' ' + word('creatures', 'creatures') + ', ' + m.minds.length + ' ' + word('minds', 'minds'), 'aria-describedby': 'net-keys' });
      const defs = s('defs');
      g.append(defs);
      for (const c of L.cols) {
        if (c.kind === 'minds' && L.narrow) {
          g.append(s('g', { 'data-rank': c.rank }, s('text', { class: 'net-col-label', x: 12, y: c.y + 14 }, c.title + ' ', s('tspan', { class: 'net-col-count' }, String(c.count)))));
          continue;
        }
        if (L.narrow) continue;
        g.append(s('g', { 'data-rank': c.rank },
          s('rect', { class: 'net-col', x: c.x, y: c.y, width: c.w, height: c.h, rx: 12 }),
          s('text', { class: 'net-col-label', x: c.x + 10, y: c.y + 20 }, c.title + ' ', s('tspan', { class: 'net-col-count' }, String(c.count)))));
      }
      const edges = s('g', { 'data-layer': 'edges' }), sparks = s('g', { 'data-layer': 'sparks' }), nodes = s('g', { 'data-layer': 'nodes' });
      g.append(edges, nodes, sparks);
      st.edgeEls = new Map(); st.nodeEls = new Map();
      // bonds
      const maxW = Math.max(1, ...m.tree.map(t => t.weight));
      const curve = (A, B) => {
        if (L.narrow) {
          const x0 = A.x, y0 = A.y + A.size / 2, x1 = B.x - B.size / 2, y1 = B.y, r = Math.min(8, Math.abs(y1 - y0) / 2, Math.abs(x1 - x0));
          return `M${x0} ${y0} V${y1 - r} Q${x0} ${y1} ${x0 + r} ${y1} H${x1}`;
        }
        const x0 = A.out, x1 = B.x - B.size / 2 - 2, k = Math.max(24, (x1 - x0) * 0.5);
        return `M${x0.toFixed(1)} ${A.y.toFixed(1)} C${(x0 + k).toFixed(1)} ${A.y.toFixed(1)} ${(x1 - k).toFixed(1)} ${B.y.toFixed(1)} ${x1.toFixed(1)} ${B.y.toFixed(1)}`;
      };
      if (!L.narrow) for (const w of m.wears) {
        const A = L.pos.get(w.from), B = L.pos.get(w.to);
        if (!A || !B) continue;
        const p = s('path', { class: 'net-edge', 'data-bond': 'wears', d: curve(A, B) });
        edges.append(p);
        st.edgeEls.set('w:' + w.from + '>' + w.to, p);
      }
      for (const t of m.tree) {
        const A = L.pos.get(t.up), B = L.pos.get(t.down);
        if (!A || !B) continue;
        const k = t.weight ? Math.log2(1 + t.weight) / Math.log2(1 + maxW) : 0;
        const p = s('path', { class: 'net-edge', 'data-bond': t.bond, 'data-implied': t.implied || false, d: curve(A, B),
          style: `stroke-width:${(1.4 + 3.2 * k).toFixed(2)};--net-edge-o:${(0.5 + 0.45 * k).toFixed(2)}` });
        p.append(s('title', {}, nameOf(t.up) + ' → ' + nameOf(t.down) + ' · ' + word('bond.' + t.bond, t.bond) + (t.implied ? ' (implied)' : '') + (t.weight ? ' · ' + t.weight + ' commission' + (t.weight === 1 ? '' : 's') + ' in 4 weeks' : '')));
        edges.append(p);
        st.edgeEls.set('t:' + t.up + '>' + t.down, p);
      }
      // neurons
      let k = 0;
      const cellsOf = id => m.cells.get(id);
      for (const id of [...m.order, ...m.minds.map(n => n.id)]) {
        const p = L.pos.get(id);
        if (!p) continue;
        const n = m.byId.get(id), mind = n.kind === 'mind', S = p.size, x0 = p.x - S / 2, y0 = p.y - S / 2;
        const a = m.act[n.name];
        const rank = mind ? n.lane : (n.rank || 'plain');
        const node = s('g', { class: 'net-node', 'data-id': id, 'data-rank': rank, 'data-kind': mind ? 'mind' : 'body', 'data-throne': n.throne || false,
          'data-stress': n.desk && n.desk.stress && n.desk.stress !== 'ok' ? n.desk.stress : false, tabindex: -1, role: 'button', 'aria-label': aria(n, a) });
        node.append(s('rect', { class: 'net-halo', x: x0 - 6, y: y0 - 6, width: S + 12, height: S + 12, rx: S * 0.3 + 4 }));
        node.append(s('rect', { class: 'net-tile', x: x0, y: y0, width: S, height: S, rx: S * 0.22 }));
        if (mind) {
          node.append(s('g', { transform: `translate(${p.x.toFixed(1)} ${p.y.toFixed(1)}) scale(${(S * 0.06).toFixed(2)})` }, s('path', { class: 'net-brain', d: BRAIN }), s('path', { class: 'net-brain-fold', d: 'M0 -6 L0 4.4' })));
        } else {
          // the activity tile: 4 weeks × 7 days, oldest top-left; blank when no instrument saw it
          const pad = S * 0.14, cw = (S - 2 * pad) / 7, ch = (S - 2 * pad) / 4;
          const v = cellsOf(id) || [];
          const days = v.slice(-28), off = 28 - days.length;
          for (let j = 0; j < 28; j++) {
            const val = j >= off ? days[j - off] : 0;
            const o = val && m.peak > 0 ? 0.3 + 0.7 * Math.log1p(val) / Math.log1p(m.peak) : 0.13;
            node.append(s('rect', { class: 'net-cell', 'data-on': val ? true : false, x: (x0 + pad + (j % 7) * cw + 0.35).toFixed(2), y: (y0 + pad + Math.floor(j / 7) * ch + 0.35).toFixed(2), width: (cw - 0.7).toFixed(2), height: (ch - 0.7).toFixed(2), rx: 0.6, opacity: o.toFixed(2) }));
          }
          // the rank's portrait as a badge on the tile's corner
          if (PORTRAITS.has(n.rank) && (p.d <= 2 || L.narrow)) {
            const r = n.throne ? 11 : 8, cx = x0 + 1, cy = y0 + 1, cid = 'netp-' + (k++);
            defs.append(s('clipPath', { id: cid }, s('circle', { cx, cy, r: r - 1.2 })));
            node.append(s('circle', { class: 'net-ring', cx, cy, r }));
            const img = s('image', { class: 'net-portrait', href: 'web/portraits/' + n.rank + '.png', x: cx - r, y: cy - r, width: r * 2, height: r * 2, preserveAspectRatio: 'xMidYMid slice', 'clip-path': `url(#${cid})` });
            img.addEventListener('error', () => img.remove());
            node.append(img);
          }
        }
        node.append(s('text', { class: 'net-label', x: p.lx, y: p.y }, p.label));
        if (n.desk && mind && n.desk.thoughts) node.append(s('text', { class: 'net-sub', x: p.out + 2, y: p.y }, n.desk.thoughts + '/' + n.desk.limit));
        node.append(s('title', {}, n.name));
        node.addEventListener('click', () => pin(id));
        node.addEventListener('pointerenter', () => preview(id));
        node.addEventListener('pointerleave', () => preview(''));
        node.addEventListener('focus', () => { st.current = id; preview(id); });
        node.addEventListener('blur', () => preview(''));
        node.addEventListener('keydown', ev => keys(ev, id));
        nodes.append(node);
        st.nodeEls.set(id, node);
      }
      // one tab stop: the current neuron (else the throne)
      const first = st.nodeEls.get(st.current) || st.nodeEls.get(m.throne && m.throne.id) || nodes.firstElementChild;
      if (first) first.setAttribute('tabindex', '0');
      st.sparks = sparks;
      stage.replaceChildren(g);
      applyLive(); applyFocus();
    }
    const nameOf = id => { const n = st.m && st.m.byId.get(id); return n ? (n.throne ? n.label || n.name : n.name) : id; };
    function aria(n, a) {
      const m = st.m;
      const bits = [n.throne ? (n.label || n.name) : n.name];
      if (n.kind === 'mind') bits.push(word('mind', 'mind'));
      else {
        bits.push(word('rank.' + n.rank, n.rank || 'body'));
        const p = m.parent.get(n.id);
        if (p) bits.push('answers to ' + nameOf(p));
        const kids = (m.kids.get(n.id) || []).length;
        if (kids) bits.push(kids + ' below');
      }
      bits.push(a && a.total ? a.total + ' run' + (a.total === 1 ? '' : 's') + ' in 4 weeks' : 'no journaled work in 4 weeks');
      return bits.join(', ');
    }

    // ---------- focus: hover previews, click or Enter pins ----------
    function pin(id) { st.focus = st.focus === id ? '' : id; if (id) st.current = id; applyFocus(); opts.onFocus && opts.onFocus(st.focus); }
    function preview(id) { st.hover = id; applyFocus(); }
    function applyFocus() {
      const id = st.hover || st.focus;
      host.toggleAttribute('data-focus', !!id);
      const lit = id && st.m ? litSet(st.m, id) : null;
      for (const [nid, el] of st.nodeEls) {
        el.toggleAttribute('data-near', !!lit && lit.nodes.has(nid));
        el.toggleAttribute('data-self', nid === id);
        el.setAttribute('aria-pressed', String(nid === st.focus));
      }
      for (const [k, el] of st.edgeEls) {
        const key = k.slice(2);
        el.toggleAttribute('data-lit', !!lit && (k[0] === 't' ? lit.tree.has(key) : lit.wears.has(key)));
      }
      if (st.three) st.three.focus(id, lit, st.focus);
    }
    function keys(ev, id) {
      const m = st.m, L = st.L;
      if (!m || !L) return;
      let to = '';
      const p = L.pos.get(id);
      const peers = [...L.pos.keys()].filter(x => L.pos.get(x).d === p.d && (L.narrow ? true : true)).sort((a, b) => L.pos.get(a).y - L.pos.get(b).y);
      const i = peers.indexOf(id);
      switch (ev.key) {
        case 'ArrowLeft': to = m.parent.get(id) || ''; break;
        case 'ArrowRight': to = (m.kids.get(id) || []).slice().sort((a, b) => L.pos.get(a).y - L.pos.get(b).y)[0] || (m.wears.find(w => w.from === id) || {}).to || ''; break;
        case 'ArrowUp': to = peers[i - 1] || ''; break;
        case 'ArrowDown': to = peers[i + 1] || ''; break;
        case 'Home': to = m.throne ? m.throne.id : ''; break;
        case 'Enter': case ' ': ev.preventDefault(); pin(id); return;
        case 'Escape': if (st.focus) { ev.preventDefault(); ev.stopPropagation(); pin(st.focus); } return;
        default: return;
      }
      ev.preventDefault();
      if (!to || !st.nodeEls.has(to)) return;
      st.nodeEls.get(id).setAttribute('tabindex', '-1');
      const el = st.nodeEls.get(to);
      el.setAttribute('tabindex', '0');
      el.focus();
    }

    // ---------- live: running bodies glow, their bond to the dispatcher flows ----------
    function applyLive() {
      const m = st.m;
      if (!m) return;
      const liveIds = new Set();
      for (const [id, n] of m.byId) if (st.live.has(n.name)) liveIds.add(id);
      for (const [id, el] of st.nodeEls) el.toggleAttribute('data-live', liveIds.has(id));
      for (const [k, el] of st.edgeEls) {
        if (k[0] !== 't') continue;
        const [up, down] = k.slice(2).split('>');
        el.toggleAttribute('data-live', liveIds.has(down) && (liveIds.has(up) || up === (m.throne && m.throne.id)));
      }
      if (st.three) st.three.live(liveIds);
    }

    // ---------- signals: a pulse down on a dispatch, back up on a report ----------
    function signal(kind, name, tone) {
      const m = st.m;
      if (!m) return;
      let id = '';
      for (const [nid, n] of m.byId) if (n.name === name && n.kind !== 'mind') id = nid;
      if (!id || (m.throne && id === m.throne.id)) return;
      const hops = route(m, id, st.live);
      if (st.three) st.three.signal(kind, hops, id, tone);
      if (st.mode === '3d') return;
      if (reducedQ.matches || !st.visible || !hops.length) { ping(kind === 'up' ? hops.length ? hops[0].split('>')[0] : id : id, kind === 'up' ? tone : 'down'); return; }
      const els = hops.map(h => st.edgeEls.get('t:' + h)).filter(Boolean);
      if (!els.length) { ping(id, kind === 'up' ? tone : 'down'); return; }
      if (kind === 'up') els.reverse();
      const t = kind === 'up' ? (tone || 'ok') : '';
      const glow = s('circle', { class: 'net-spark', 'data-glow': true, 'data-tone': t || false, r: 9 });
      const trail = [0, 1, 2].map(i => s('circle', { class: 'net-spark', 'data-tone': t || false, r: 3.6 - i * 0.9, opacity: (1 - i * 0.3).toFixed(2) }));
      st.sparks.append(glow, ...trail);
      st.pulses.push({ els, dir: kind, at: performance.now(), per: 620, glow, trail, end: kind === 'up' ? hops[0].split('>')[0] : id, tone: kind === 'up' ? (tone || 'ok') : 'down', lens: els.map(e => e.getTotalLength()) });
      tick();
    }
    function ping(id, tone, kind) {
      const el = st.nodeEls.get(id);
      if (!el) return;
      el.setAttribute('data-ping', tone || 'down');
      clearTimeout(el._ping);
      el._ping = setTimeout(() => el.removeAttribute('data-ping'), kind === 'read' ? 700 : 1400);
      if (!reducedQ.matches && st.L) {
        const p = st.L.pos.get(id);
        if (!p) return;
        const r = s('circle', { class: 'net-ripple', 'data-tone': tone === 'down' ? false : tone, 'data-kind': kind || false, cx: p.x, cy: p.y, r: p.size * 0.7 });
        el.append(r);
        setTimeout(() => r.remove(), 1300);
      }
    }
    // an action on a neuron: a tool step, a gate verdict — each class reads its own way, calmly
    // (read: a soft accent ripple · write: the rank's own pulse · outward/destructive and a waiting
    // gate: an amber ring · pass: a green flare · fail or denied: a red one). Bursts coalesce: one
    // ripple per neuron per quarter second unless a louder one arrives.
    const LOUD = { read: 0, write: 1, gate: 2, outward: 2, destructive: 3, pass: 3, fail: 4 };
    const TONE = { read: 'info', write: 'rank', gate: 'warn', outward: 'warn', destructive: 'warn', pass: 'ok', fail: 'danger' };
    const lastAct = new Map();
    function act(name, kind) {
      const m = st.m;
      if (!m || !(kind in LOUD)) return;
      let id = m.throne ? m.throne.id : '';
      if (name) for (const [nid, n] of m.byId) if (n.name === name && n.kind !== 'mind') { id = nid; break; }
      if (!id) return;
      const now = performance.now(), was = lastAct.get(id);
      if (was && now - was.at < 250 && LOUD[kind] <= was.loud) return;
      lastAct.set(id, { at: now, loud: LOUD[kind] });
      if (st.three) st.three.act(id, TONE[kind], kind);
      if (st.mode !== '3d' && st.visible) ping(id, TONE[kind], kind);
    }
    function tick() {
      if (st.raf || !st.pulses.length || !st.visible) return;
      st.raf = requestAnimationFrame(frame);
    }
    function frame(now) {
      st.raf = 0;
      st.pulses = st.pulses.filter(p => {
        const total = p.per * p.els.length;
        const t = Math.min(1, (now - p.at) / total);
        const hop = Math.min(p.els.length - 1, Math.floor(t * p.els.length));
        const local = t * p.els.length - hop;
        const place = (c, lag) => {
          let lt = local - lag, h = hop;
          if (lt < 0) { h--; lt += 1; }
          if (h < 0) { c.setAttribute('opacity', '0'); return; }
          const len = p.lens[h], u = p.dir === 'up' ? 1 - lt : lt;
          const e = ease(u), pt = p.els[h].getPointAtLength(len * (p.dir === 'up' ? 1 - ease(1 - u) : e));
          c.setAttribute('cx', pt.x.toFixed(1)); c.setAttribute('cy', pt.y.toFixed(1));
        };
        place(p.glow, 0);
        p.trail.forEach((c, i) => place(c, i * 0.05));
        if (t >= 1) { p.glow.remove(); p.trail.forEach(c => c.remove()); ping(p.end, p.tone); return false; }
        return true;
      });
      if (st.pulses.length) tick();
    }
    const ease = u => u < 0.5 ? 2 * u * u : 1 - Math.pow(-2 * u + 2, 2) / 2;

    // ---------- calm: animate only while seen ----------
    const io = new IntersectionObserver(es => { for (const e of es) setVisible(e.isIntersecting && !document.hidden); });
    io.observe(stage);
    document.addEventListener('visibilitychange', () => setVisible(!document.hidden && stage.offsetParent !== null));
    function setVisible(v) {
      st.visible = v;
      host.toggleAttribute('data-asleep', !v);
      if (st.three) st.three.running(v && st.mode === '3d');
      if (v) tick();
    }

    // ---------- the 3D stage ----------
    const can3d = () => !!(window.isekaiNet3D && window.isekaiNet3D.supported());
    function mode(v) {
      if (v === '3d' && !can3d()) v = '2d';
      st.mode = v;
      if (v === '3d') {
        if (!st.three) {
          st.three = window.isekaiNet3D.create(stage, { probes, word, onPick: id => pin(id || st.focus), onHover: id => preview(id), reduced: () => reducedQ.matches });
          if (!st.three) { st.mode = '2d'; return mode('2d'); }
        }
        stage.dataset.mode = '3d';
        stage.replaceChildren(...st.three.elements);
        if (st.m) st.three.model(st.m);
        st.edgeEls = new Map(); st.nodeEls = new Map();
        applyLive(); applyFocus();
        st.three.running(st.visible);
      } else {
        if (st.three) { st.three.running(false); st.three.dispose(); st.three = null; }
        stage.dataset.mode = '2d';
        draw();
      }
      return st.mode;
    }

    let resizeT = 0;
    new ResizeObserver(() => { clearTimeout(resizeT); resizeT = setTimeout(draw, 100); }).observe(host);
    return {
      set(colony, activity) {
        st.m = build(colony, activity, word);
        if (st.focus && !st.m.byId.has(st.focus)) st.focus = '';
        if (st.three) st.three.model(st.m);
        draw();
        return st.m;
      },
      live(names) { st.live = new Set(names); applyLive(); },
      signal,
      act,
      hover(id) { preview(id || ''); },
      focus(id) { st.focus = id || ''; applyFocus(); },
      mode,
      can3d,
      redraw: draw,
      model: () => st.m,
      route: id => st.m ? route(st.m, id, st.live) : [],
    };
  }

  window.isekaiNet = { create, build, litSet, route };
})();
