const actionClass = {
  log: "action-log",
  suggest: "action-suggest",
  alert: "action-alert",
  execute: "action-execute",
};

const badgeStyle = {
  log: "background:#0D2B14;color:#3FB950",
  suggest: "background:#2B1E00;color:#F0A500",
  alert: "background:#2B0D0D;color:#F85149",
  execute: "background:#001A2B;color:#00E5FF",
};

let logLines = [];
let lastCycleNum = 0;

function addLog(tag, tagClass, msg, time) {
  logLines.unshift({ tag, tagClass, msg, time });
  if (logLines.length > 50) logLines.pop();
  renderConsole();
}

function renderConsole() {
  var el = document.getElementById("console-log");
  el.innerHTML = logLines
    .map(function (l) {
      return (
        '<div class="log-line">' +
        '<span class="log-time">' +
        l.time +
        "</span>" +
        '<span class="log-tag ' +
        l.tagClass +
        '">' +
        l.tag +
        "</span>" +
        '<span class="log-msg">' +
        l.msg +
        "</span>" +
        "</div>"
      );
    })
    .join("");
}

function setBar(id, pct) {
  var el = document.getElementById(id);
  el.style.width = pct + "%";
  el.className =
    "bar-fill " + (pct >= 90 ? "danger" : pct >= 75 ? "warn" : "ok");
}

function setMetricColor(id, pct) {
  var el = document.getElementById(id);
  el.className =
    "metric-value " + (pct >= 90 ? "danger" : pct >= 75 ? "warn" : "");
}

function setLoopStep(active) {
  var steps = ["observe", "think", "decide", "act", "sleep"];
  steps.forEach(function (s) {
    var el = document.getElementById("step-" + s);
    if (s === active) {
      el.className = "loop-step active";
      el.querySelector(".step-icon").textContent = "▸";
    } else if (steps.indexOf(s) < steps.indexOf(active)) {
      el.className = "loop-step done";
      el.querySelector(".step-icon").textContent = "✓";
    } else {
      el.className = "loop-step idle";
      el.querySelector(".step-icon").textContent = "○";
    }
  });
}

function fmt(t) {
  return new Date(t).toLocaleTimeString("fr-FR", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

async function fetchStatus() {
  try {
    var r = await fetch("/api/status");
    var d = await r.json();

    document.getElementById("model-pill").textContent = d.model;
    document.getElementById("uptime-pill").textContent = d.uptime;
    document.getElementById("info-model").textContent = d.model;
    document.getElementById("info-interval").textContent = d.interval;
    document.getElementById("info-cycle").textContent = "#" + d.cycle_num;
    document.getElementById("info-uptime").textContent = d.uptime;
    document.getElementById("footer-cycle").textContent =
      "cycle #" + d.cycle_num + " · " + d.interval;

    if (d.last_cycle) {
      var c = d.last_cycle;
      var s = c.snapshot;
      var t = fmt(c.timestamp);

      document.getElementById("cpu-val").textContent =
        s.cpu_percent.toFixed(1) + "%";
      document.getElementById("cpu-meta").textContent = "stable";
      setBar("cpu-bar", s.cpu_percent);
      setMetricColor("cpu-val", s.cpu_percent);

      var ramPct = s.mem_percent;
      document.getElementById("ram-val").textContent =
        (s.mem_used_mb / 1024).toFixed(1) + " GB";
      document.getElementById("ram-meta").textContent =
        ramPct.toFixed(0) +
        "% · " +
        (s.mem_total_mb / 1024).toFixed(1) +
        " GB total";
      setBar("ram-bar", ramPct);
      setMetricColor("ram-val", ramPct);

      document.getElementById("disk-val").textContent =
        s.disk_percent.toFixed(1) + "%";
      document.getElementById("disk-meta").textContent =
        s.disk_used_gb + " / " + s.disk_total_gb + " GB · C:\\";
      setBar("disk-bar", s.disk_percent);
      setMetricColor("disk-val", s.disk_percent);

      var actionEl = document.getElementById("decision-action");
      actionEl.textContent = c.action;
      actionEl.className =
        "decision-action " + (actionClass[c.action] || "action-log");

      document.getElementById("decision-text").textContent = c.analysis || "—";
      var reasonEl = document.getElementById("decision-reason");
      if (c.reason) {
        reasonEl.textContent = c.reason;
        reasonEl.style.display = "block";
      }

      if (c.cycle_num !== lastCycleNum) {
        lastCycleNum = c.cycle_num;
        addLog(
          "[ OBSERVE ]",
          "tag-obs",
          "CPU " +
            s.cpu_percent.toFixed(1) +
            "% · RAM " +
            ramPct.toFixed(0) +
            "% · Disk " +
            s.disk_percent.toFixed(0) +
            "%",
          t,
        );
        addLog(
          "[ AGENT ]",
          "tag-agent",
          c.action.toUpperCase() + " — " + (c.analysis || "").slice(0, 60),
          t,
        );
        if (c.command) {
          addLog("[ EXEC ]", "tag-exec", c.command, t);
        }

        setLoopStep(c.command ? "act" : "sleep");
        setTimeout(function () {
          setLoopStep("sleep");
        }, 2000);
        setTimeout(function () {
          setLoopStep("observe");
        }, 5000);

        var dot = document.getElementById("refresh-dot");
        dot.classList.add("flash");
        setTimeout(function () {
          dot.classList.remove("flash");
        }, 400);
      }
    }
  } catch (e) {
    addLog("[ ERR ]", "tag-err", "Connexion perdue", "--:--:--");
  }
}

async function fetchHistory() {
  try {
    var r = await fetch("/api/history");
    var d = await r.json();

    var list = document.getElementById("history-list");
    if (!d.cycles || d.cycles.length === 0) {
      list.innerHTML =
        '<div style="color:var(--muted);font-family:var(--mono);font-size:11px">Aucun cycle</div>';
      return;
    }

    list.innerHTML = d.cycles
      .map(function (c) {
        var s = c.snapshot;
        var t = fmt(c.timestamp);
        var style = badgeStyle[c.action] || badgeStyle.log;
        var cmd = c.command ? " · " + c.command : "";
        return (
          '<div class="history-row">' +
          '<span class="h-time">' +
          t +
          "</span>" +
          '<span class="h-metrics">CPU ' +
          s.cpu_percent.toFixed(0) +
          "%" +
          "<span>·</span>RAM " +
          s.mem_percent.toFixed(0) +
          "%" +
          "<span>·</span>Disk " +
          s.disk_percent.toFixed(0) +
          "%</span>" +
          '<span class="h-badge" style="' +
          style +
          '">' +
          c.action +
          cmd +
          "</span>" +
          "</div>"
        );
      })
      .join("");
  } catch (e) {}
}

fetchStatus();
fetchHistory();
setInterval(fetchStatus, 5000);
setInterval(fetchHistory, 15000);
