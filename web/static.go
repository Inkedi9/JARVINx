package web

const indexHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>JARVINx</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@300;400;500;600&display=swap" rel="stylesheet">
<style>
  :root {
    --bg:       #0D1117;
    --surface:  #161B22;
    --surface2: #1C2128;
    --border:   #21262D;
    --accent:   #00E5FF;
    --accent2:  #005BFF;
    --text:     #E6EDF3;
    --muted:    #8B949E;
    --ok:       #3FB950;
    --warn:     #F0A500;
    --danger:   #F85149;
    --mono:     'JetBrains Mono', monospace;
    --sans:     'Inter', system-ui, sans-serif;
  }

  * { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    background: var(--bg);
    color: var(--text);
    font-family: var(--sans);
    font-size: 13px;
    min-height: 100vh;
  }

  /* ── Topbar ── */
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 24px;
    height: 52px;
    border-bottom: 0.5px solid var(--border);
    background: var(--surface);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .logo-mark {
    width: 28px;
    height: 28px;
    background: var(--accent);
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--mono);
    font-size: 11px;
    font-weight: 600;
    color: #0D1117;
  }

  .logo-name {
    font-family: var(--mono);
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 3px;
    color: var(--text);
    text-shadow:
      0 1px 0 #000,
      0 2px 8px rgba(0,229,255,0.15);
  }

  .logo-tagline {
    font-size: 10px;
    color: var(--accent);
    letter-spacing: 2px;
    text-transform: uppercase;
    margin-top: 1px;
  }

  .topbar-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .pill {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--surface2);
    border: 0.5px solid var(--border);
    border-radius: 20px;
    padding: 4px 12px;
    font-size: 11px;
    color: var(--muted);
    font-family: var(--mono);
  }

  .pulse {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--accent);
    animation: pulse 2s infinite;
  }

  @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.3} }

  /* ── Layout ── */
  .main {
    display: grid;
    grid-template-columns: 1fr 320px;
    gap: 16px;
    padding: 20px 24px;
    max-width: 1200px;
    margin: 0 auto;
  }

  .left { display: flex; flex-direction: column; gap: 16px; }
  .right { display: flex; flex-direction: column; gap: 16px; }

  /* ── Cards ── */
  .card {
    background: var(--surface);
    border: 0.5px solid var(--border);
    border-radius: 10px;
    padding: 16px 18px;
  }

  .card-title {
    font-size: 10px;
    font-family: var(--mono);
    color: var(--muted);
    letter-spacing: 1.5px;
    text-transform: uppercase;
    margin-bottom: 14px;
  }

  /* ── Metrics ── */
  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  .metric {
    background: var(--surface2);
    border: 0.5px solid var(--border);
    border-radius: 8px;
    padding: 12px 14px;
  }

  .metric-label {
    font-size: 10px;
    font-family: var(--mono);
    color: var(--muted);
    letter-spacing: 1px;
    text-transform: uppercase;
    margin-bottom: 6px;
  }

  .metric-value {
    font-family: var(--mono);
    font-size: 28px;
    font-weight: 600;
    color: var(--text);
    line-height: 1;
    margin-bottom: 8px;
    text-shadow: 0 1px 0 #000, 0 0 20px rgba(0,229,255,0.08);
  }

  .metric-value.warn { color: var(--warn); text-shadow: 0 0 20px rgba(240,165,0,0.15); }
  .metric-value.danger { color: var(--danger); }
  .metric-value.ok { color: var(--ok); }

  .bar-track {
    height: 3px;
    background: var(--border);
    border-radius: 2px;
    overflow: hidden;
    margin-bottom: 6px;
  }

  .bar-fill {
    height: 100%;
    border-radius: 2px;
    background: var(--accent);
    transition: width 0.8s ease;
  }

  .bar-fill.warn { background: var(--warn); }
  .bar-fill.danger { background: var(--danger); }
  .bar-fill.ok { background: var(--ok); }
  .metric-meta { font-size: 10px; color: var(--muted); font-family: var(--mono); }

  /* ── Decision ── */
  .decision-action {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 10px;
    border-radius: 4px;
    font-family: var(--mono);
    font-size: 11px;
    font-weight: 500;
    margin-bottom: 10px;
  }

  .action-log     { background: #0D2B14; color: var(--ok);     border: 0.5px solid #3FB95030; }
  .action-suggest { background: #2B1E00; color: var(--warn);   border: 0.5px solid #F0A50030; }
  .action-alert   { background: #2B0D0D; color: var(--danger); border: 0.5px solid #F8514930; }
  .action-execute { background: #001A2B; color: var(--accent);  border: 0.5px solid #00E5FF30; }

  .decision-text {
    font-size: 13px;
    color: var(--text);
    line-height: 1.6;
    margin-bottom: 8px;
  }

  .decision-reason {
    font-size: 11px;
    color: var(--muted);
    line-height: 1.5;
    border-left: 2px solid var(--border);
    padding-left: 10px;
    font-style: italic;
  }

  /* ── macOS Console ── */
  .console-wrap {
    background: #1C2128;
    border: 0.5px solid var(--border);
    border-radius: 10px;
    overflow: hidden;
  }

  .console-topbar {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 9px 12px;
    background: #252B33;
    border-bottom: 0.5px solid var(--border);
  }

  .console-btn {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .btn-red    { background: #FF5F57; }
  .btn-yellow { background: #FFBD2E; }
  .btn-green  { background: #28C840; }

  .console-title {
    flex: 1;
    text-align: center;
    font-size: 11px;
    color: var(--muted);
    font-family: var(--mono);
    margin-left: -36px;
  }

  .console-body {
    padding: 12px 14px;
    font-family: var(--mono);
    font-size: 11px;
    line-height: 1.8;
    color: #8B949E;
    height: 220px;
    overflow-y: auto;
    display: flex;
    flex-direction: column-reverse;
  }

  .console-body::-webkit-scrollbar { width: 4px; }
  .console-body::-webkit-scrollbar-track { background: transparent; }
  .console-body::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }

  .log-line { display: flex; gap: 8px; }
  .log-time { color: #444C56; flex-shrink: 0; }
  .log-tag  { flex-shrink: 0; }
  .tag-obs  { color: var(--accent); }
  .tag-agent{ color: #7C3AED; }
  .tag-exec { color: var(--ok); }
  .tag-err  { color: var(--danger); }
  .tag-state{ color: var(--muted); }
  .log-msg  { color: #C9D1D9; }

  /* ── Agent loop ── */
  .loop-steps { display: flex; flex-direction: column; gap: 4px; }

  .loop-step {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    border-radius: 6px;
    font-family: var(--mono);
    font-size: 11px;
    color: var(--muted);
    transition: all 0.3s;
  }

  .loop-step.done   { color: var(--ok); }
  .loop-step.active { background: #001A2B; color: var(--accent); }
  .loop-step.idle   { color: #444C56; }

  .step-icon { width: 16px; text-align: center; }

  /* ── History table ── */
  .history-list { display: flex; flex-direction: column; gap: 0; }

  .history-row {
    display: grid;
    grid-template-columns: 52px 1fr auto;
    align-items: center;
    gap: 8px;
    padding: 8px 0;
    border-bottom: 0.5px solid var(--border);
    font-family: var(--mono);
    font-size: 11px;
  }

  .history-row:last-child { border-bottom: none; }

  .h-time { color: var(--muted); }

  .h-metrics { color: #C9D1D9; }
  .h-metrics span { color: var(--muted); margin: 0 2px; }

  .h-badge {
    padding: 2px 8px;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 500;
    white-space: nowrap;
  }

  /* ── Footer ── */
  .footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 24px;
    border-top: 0.5px solid var(--border);
    font-family: var(--mono);
    font-size: 10px;
    color: #444C56;
  }

  .footer-accent { color: var(--accent); }

  .uptime-badge {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  /* ── Refresh indicator ── */
  .refresh-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--accent);
    opacity: 0;
    transition: opacity 0.2s;
  }

  .refresh-dot.flash { opacity: 1; }
</style>
</head>
<body>

<div class="topbar">
  <div class="logo">
    <div class="logo-mark">JX</div>
    <div>
      <div class="logo-name">JARVINX</div>
      <div class="logo-tagline">Autonomous agent runtime</div>
    </div>
  </div>
  <div class="topbar-right">
    <div class="pill">
      <div class="pulse"></div>
      <span id="model-pill">—</span>
    </div>
    <div class="pill" id="uptime-pill">—</div>
    <div class="refresh-dot" id="refresh-dot"></div>
  </div>
</div>

<div class="main">
  <div class="left">

    <!-- Métriques -->
    <div class="card">
      <div class="card-title">System state</div>
      <div class="metrics-grid">
        <div class="metric">
          <div class="metric-label">CPU</div>
          <div class="metric-value" id="cpu-val">—</div>
          <div class="bar-track"><div class="bar-fill" id="cpu-bar" style="width:0%"></div></div>
          <div class="metric-meta" id="cpu-meta">—</div>
        </div>
        <div class="metric">
          <div class="metric-label">RAM</div>
          <div class="metric-value" id="ram-val">—</div>
          <div class="bar-track"><div class="bar-fill" id="ram-bar" style="width:0%"></div></div>
          <div class="metric-meta" id="ram-meta">—</div>
        </div>
        <div class="metric">
          <div class="metric-label">Disk</div>
          <div class="metric-value" id="disk-val">—</div>
          <div class="bar-track"><div class="bar-fill" id="disk-bar" style="width:0%"></div></div>
          <div class="metric-meta" id="disk-meta">—</div>
        </div>
      </div>
    </div>

    <!-- Décision agent -->
    <div class="card">
      <div class="card-title">Agent decision</div>
      <div id="decision-action" class="decision-action action-log">log</div>
      <div id="decision-text" class="decision-text">En attente du premier cycle...</div>
      <div id="decision-reason" class="decision-reason" style="display:none"></div>
    </div>

    <!-- Historique -->
    <div class="card">
      <div class="card-title">Cycle history</div>
      <div class="history-list" id="history-list">
        <div style="color:var(--muted);font-family:var(--mono);font-size:11px">Chargement...</div>
      </div>
    </div>

  </div>

  <div class="right">

    <!-- Console macOS -->
    <div class="console-wrap">
      <div class="console-topbar">
        <div class="console-btn btn-red"></div>
        <div class="console-btn btn-yellow"></div>
        <div class="console-btn btn-green"></div>
        <div class="console-title">jarvinx — runtime.log</div>
      </div>
      <div class="console-body" id="console-log">
        <div class="log-line">
          <span class="log-time">--:--:--</span>
          <span class="log-tag tag-state">[ STATE ]</span>
          <span class="log-msg">En attente...</span>
        </div>
      </div>
    </div>

    <!-- Agent loop -->
    <div class="card">
      <div class="card-title">Agent loop</div>
      <div class="loop-steps" id="loop-steps">
        <div class="loop-step idle" id="step-observe">
          <span class="step-icon">○</span>OBSERVE
        </div>
        <div class="loop-step idle" id="step-think">
          <span class="step-icon">○</span>THINK
        </div>
        <div class="loop-step idle" id="step-decide">
          <span class="step-icon">○</span>DECIDE
        </div>
        <div class="loop-step idle" id="step-act">
          <span class="step-icon">○</span>ACT
        </div>
        <div class="loop-step idle" id="step-sleep">
          <span class="step-icon">○</span>SLEEP
        </div>
      </div>
    </div>

    <!-- Infos runtime -->
    <div class="card">
      <div class="card-title">Runtime info</div>
      <table style="width:100%;font-family:var(--mono);font-size:11px;border-collapse:collapse">
        <tr>
          <td style="color:var(--muted);padding:4px 0">Model</td>
          <td style="color:var(--text);text-align:right" id="info-model">—</td>
        </tr>
        <tr>
          <td style="color:var(--muted);padding:4px 0">Interval</td>
          <td style="color:var(--text);text-align:right" id="info-interval">—</td>
        </tr>
        <tr>
          <td style="color:var(--muted);padding:4px 0">Cycle</td>
          <td style="color:var(--accent);text-align:right" id="info-cycle">#0</td>
        </tr>
        <tr>
          <td style="color:var(--muted);padding:4px 0">Uptime</td>
          <td style="color:var(--text);text-align:right" id="info-uptime">—</td>
        </tr>
      </table>
    </div>

  </div>
</div>

<div class="footer">
  <span>JARVINx · Autonomous Agent System</span>
  <span class="footer-accent">localhost:8080</span>
  <span id="footer-cycle">cycle #0 · 0 snapshots</span>
</div>

<script>
const actionClass = {
  log:     'action-log',
  suggest: 'action-suggest',
  alert:   'action-alert',
  execute: 'action-execute',
};

const badgeStyle = {
  log:     'background:#0D2B14;color:#3FB950',
  suggest: 'background:#2B1E00;color:#F0A500',
  alert:   'background:#2B0D0D;color:#F85149',
  execute: 'background:#001A2B;color:#00E5FF',
};

let logLines = [];
let lastCycleNum = 0;

function addLog(tag, tagClass, msg, time) {
  logLines.unshift({ tag, tagClass, msg, time });
  if (logLines.length > 50) logLines.pop();
  renderConsole();
}

function renderConsole() {
  var el = document.getElementById('console-log');
  el.innerHTML = logLines.map(function(l) {
    return '<div class="log-line">' +
      '<span class="log-time">' + l.time + '</span>' +
      '<span class="log-tag ' + l.tagClass + '">' + l.tag + '</span>' +
      '<span class="log-msg">' + l.msg + '</span>' +
      '</div>';
  }).join('');
}

function setBar(id, pct) {
  const el = document.getElementById(id);
  el.style.width = pct + '%';
  el.className = 'bar-fill ' + (pct >= 90 ? 'danger' : pct >= 75 ? 'warn' : 'ok');
}

function setMetricColor(id, pct) {
  const el = document.getElementById(id);
  el.className = 'metric-value ' + (pct >= 90 ? 'danger' : pct >= 75 ? 'warn' : '');
}

function setLoopStep(active) {
  const steps = ['observe','think','decide','act','sleep'];
  steps.forEach(s => {
    const el = document.getElementById('step-' + s);
    if (s === active) {
      el.className = 'loop-step active';
      el.querySelector('.step-icon').textContent = '▸';
    } else if (steps.indexOf(s) < steps.indexOf(active)) {
      el.className = 'loop-step done';
      el.querySelector('.step-icon').textContent = '✓';
    } else {
      el.className = 'loop-step idle';
      el.querySelector('.step-icon').textContent = '○';
    }
  });
}

function fmt(t) {
  return new Date(t).toLocaleTimeString('fr-FR', {hour:'2-digit',minute:'2-digit',second:'2-digit'});
}

async function fetchStatus() {
  try {
    const r = await fetch('/api/status');
    const d = await r.json();

    document.getElementById('model-pill').textContent = d.model;
    document.getElementById('uptime-pill').textContent = d.uptime;
    document.getElementById('info-model').textContent = d.model;
    document.getElementById('info-interval').textContent = d.interval;
    document.getElementById('info-cycle').textContent = '#' + d.cycle_num;
    document.getElementById('info-uptime').textContent = d.uptime;
    document.getElementById('footer-cycle').textContent =
      'cycle #' + d.cycle_num + ' · ' + d.interval;

    if (d.last_cycle) {
      const c = d.last_cycle;
      const s = c.snapshot;
      const t = fmt(c.timestamp);

      document.getElementById('cpu-val').textContent = s.cpu_percent.toFixed(1) + '%';
      document.getElementById('cpu-meta').textContent = 'stable';
      setBar('cpu-bar', s.cpu_percent);
      setMetricColor('cpu-val', s.cpu_percent);

      const ramPct = s.mem_percent;
      document.getElementById('ram-val').textContent = (s.mem_used_mb / 1024).toFixed(1) + ' GB';
      document.getElementById('ram-meta').textContent = ramPct.toFixed(0) + '% · ' + (s.mem_total_mb/1024).toFixed(1) + ' GB total';
      setBar('ram-bar', ramPct);
      setMetricColor('ram-val', ramPct);

      document.getElementById('disk-val').textContent = s.disk_percent.toFixed(1) + '%';
      document.getElementById('disk-meta').textContent = s.disk_used_gb + ' / ' + s.disk_total_gb + ' GB · C:\\';
      setBar('disk-bar', s.disk_percent);
      setMetricColor('disk-val', s.disk_percent);

      const actionEl = document.getElementById('decision-action');
      actionEl.textContent = c.action;
      actionEl.className = 'decision-action ' + (actionClass[c.action] || 'action-log');

      document.getElementById('decision-text').textContent = c.analysis || '—';
      const reasonEl = document.getElementById('decision-reason');
      if (c.reason) {
        reasonEl.textContent = c.reason;
        reasonEl.style.display = 'block';
      }

      if (c.cycle_num !== lastCycleNum) {
        lastCycleNum = c.cycle_num;
        addLog('[ OBSERVE ]', 'tag-obs',
          'CPU ' + s.cpu_percent.toFixed(1) + '% · RAM ' + ramPct.toFixed(0) + '% · Disk ' + s.disk_percent.toFixed(0) + '%', t);
        addLog('[ AGENT ]', 'tag-agent', c.action.toUpperCase() + ' — ' + (c.analysis || '').slice(0, 60), t);
        if (c.command) {
          addLog('[ EXEC ]', 'tag-exec', c.command, t);
        }

        setLoopStep(c.command ? 'act' : 'sleep');
        setTimeout(() => setLoopStep('sleep'), 2000);
        setTimeout(() => setLoopStep('observe'), 5000);

        const dot = document.getElementById('refresh-dot');
        dot.classList.add('flash');
        setTimeout(() => dot.classList.remove('flash'), 400);
      }
    }
  } catch(e) {
    addLog('[ ERR ]', 'tag-err', 'Connexion perdue', '--:--:--');
  }
}

async function fetchHistory() {
  try {
    const r = await fetch('/api/history');
    const d = await r.json();

    const list = document.getElementById('history-list');
    if (!d.cycles || d.cycles.length === 0) {
      list.innerHTML = '<div style="color:var(--muted);font-family:var(--mono);font-size:11px">Aucun cycle enregistré</div>';
      return;
    }

    list.innerHTML = d.cycles.map(function(c) {
  var s = c.snapshot;
  var t = fmt(c.timestamp);
  var style = badgeStyle[c.action] || badgeStyle.log;
  var cmd = c.command ? ' · ' + c.command : '';
  return '<div class="history-row">' +
    '<span class="h-time">' + t + '</span>' +
    '<span class="h-metrics">CPU ' + s.cpu_percent.toFixed(0) + '%' +
    '<span> · </span>RAM ' + s.mem_percent.toFixed(0) + '%' +
    '<span> · </span>Disk ' + s.disk_percent.toFixed(0) + '%</span>' +
    '<span class="h-badge" style="' + style + '">' + c.action + cmd + '</span>' +
    '</div>';
}).join('');
  } catch(e) {}
}

fetchStatus();
fetchHistory();
setInterval(fetchStatus, 5000);
setInterval(fetchHistory, 15000);
</script>
</body>
</html>`
