async function refresh() {
  let data;
  try {
    const res = await fetch("/api/stats");
    data = await res.json();
  } catch (e) {
    return;
  }

  console.log(data);

  document.getElementById("total").textContent = data.total;
  document.getElementById("hits").textContent = data.hits;
  document.getElementById("misses").textContent = data.misses;
  document.getElementById("errors").textContent = data.errors;
  document.getElementById("cacheSize").textContent = data.cacheSize;
  document.getElementById("hitRate").textContent =
    data.hitRate.toFixed(1) + "%";
}

refresh();
setInterval(refresh, 1000);
