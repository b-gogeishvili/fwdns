// sanitize DNS query names to avoid any HTML injection
const escapes = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
  "'": "&#39;",
};
const escapeHtml = (s) => s.replace(/[&<>"']/g, (c) => escapes[c]);

const set = (id, value) => (document.getElementById(id).textContent = value);

async function refresh() {
  let data;
  try {
    data = await (await fetch("/api/stats")).json();
  } catch (e) {
    return;
  }

  set("total", data.total);
  set("hits", data.hits);
  set("misses", data.misses);
  set("errors", data.errors);
  set("cacheSize", data.cacheSize);
  set("hitRate", data.hitRate.toFixed(1) + "%");
  document.getElementById("hitBar").style.width = data.hitRate + "%";

  // The server sends oldest queries first. Reverse so the newest are on top.
  document.getElementById("recent").innerHTML = (data.recent || [])
    .slice()
    .reverse()
    .map((q) => {
      const time = new Date(q.time).toLocaleTimeString();
      const tag = q.cached
        ? '<span class="tag hit">CACHE</span>'
        : '<span class="tag miss">UPSTREAM</span>';
      return `<tr>
      <td>${time}</td>
      <td>${escapeHtml(q.name)}</td>
      <td>${q.type}</td>
      <td>${tag}</td>
      <td class="num">${q.latency.toFixed(2)} ms</td>
    </tr>`;
    })
    .join("");
}

refresh();
setInterval(refresh, 500);
