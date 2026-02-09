/* smallR Interactive Demos
   - Playground with full R editor
   - Regression, Statistics, Data Frames, Strings & Functional, Time Series
   - All computation runs in WASM (no server)
   - Visualizations with D3
*/

// ════════════════════════════════════════════
// Default code snippets for each demo
// ════════════════════════════════════════════

const DEFAULT_PLAYGROUND = `# smallR Playground — write any R code here!
# Try the new features: pipes, %in%, sapply, etc.

# --- Vectors & Math ---
x <- c(3, 1, 4, 1, 5, 9, 2, 6)
cat("sorted:", sort(x), "\\n")
cat("unique:", unique(x), "\\n")
cat("cumsum:", cumsum(x), "\\n")

# --- Pipe operator ---
result <- c(5, 3, 1, 4, 2) |> sort() |> rev()
cat("piped: ", result, "\\n")

# --- %in% operator ---
cat("2 %in% 1:5 =", 2 %in% 1:5, "\\n")

# --- String operations ---
words <- c("hello", "world", "foo")
cat("upper:", toupper(words), "\\n")
cat("nchar:", sapply(words, nchar), "\\n")

# --- Apply family ---
squares <- sapply(1:10, function(x) x^2)
cat("squares:", squares, "\\n")

total <- Reduce(function(a, b) a + b, 1:10, 0)
cat("sum 1:10 =", total, "\\n")

# --- Statistics ---
data <- c(23, 45, 12, 67, 34, 89, 56, 78)
cat("\\nmean:", round(mean(data), 2), "\\n")
cat("sd:  ", round(sd(data), 2), "\\n")
cat("range:", range(data), "\\n")

# Return a named list to see JSON output
list(
  sorted = sort(data),
  mean = round(mean(data), 2),
  sd = round(sd(data), 2),
  n = length(data)
)`;

const DEFAULT_REGRESSION = `# You already have vectors x and y (generated in JS).
# Modify this code and click "Run smallR".

mx <- mean(x)
my <- mean(y)

sxy <- sum((x - mx) * (y - my))
sxx <- sum((x - mx) * (x - mx))

slope <- sxy / sxx
intercept <- my - slope * mx

yhat <- intercept + slope * x

ss_tot <- sum((y - my) * (y - my))
ss_res <- sum((y - yhat) * (y - yhat))
r2 <- 1 - ss_res / ss_tot

# smallR returns a named list -> JSON object
list(
  intercept = intercept,
  slope = slope,
  r2 = r2,
  yhat = yhat
)`;

const DEFAULT_STATS = `# Vector operations and statistics
x <- c(23, 45, 12, 67, 34, 89, 56, 78, 42, 31)

# Basic stats
cat("mean:", round(mean(x), 2), "\\n")
cat("sd:  ", round(sd(x), 2), "\\n")
cat("sum: ", sum(x), "\\n")

# Sorting and ranking
cat("sorted:    ", sort(x), "\\n")
cat("order:     ", order(x), "\\n")
cat("cumsum:    ", cumsum(sort(x)), "\\n")

# Set operations
a <- c(1, 2, 3, 4, 5)
b <- c(3, 4, 5, 6, 7)
cat("union:     ", union(a, b), "\\n")
cat("intersect: ", intersect(a, b), "\\n")
cat("setdiff:   ", setdiff(a, b), "\\n")

# NA handling
y <- c(10, 20, NA, 40, 50, NA, 70)
cat("has NAs:   ", any(is.na(y)), "\\n")
cat("mean(na.rm):", round(mean(y, na.rm=TRUE), 2), "\\n")

# Return data for visualization
list(
  values = x,
  sorted = sort(x),
  labels = paste0("x", seq_along(x)),
  mean = mean(x),
  sd = sd(x)
)`;

const DEFAULT_DATAFRAME = `# Create a data frame
ages <- c(25, 30, 35, 28, 42, 31, 27, 38)
scores <- c(85, 92, 78, 88, 95, 73, 91, 82)
names_vec <- c("Alice", "Bob", "Carol", "Dan",
               "Eve", "Frank", "Grace", "Hank")

df <- data.frame(
  name = names_vec,
  age = ages,
  score = scores
)

cat("Rows:", nrow(df), " Cols:", ncol(df), "\\n")
cat("Names:", names(df), "\\n\\n")

# Column access
cat("Scores:", df$score, "\\n")
cat("Mean score:", round(mean(df$score), 1), "\\n")
cat("Mean age:  ", round(mean(df$age), 1), "\\n\\n")

# Add computed column
df$pass <- df$score >= 85
cat("Passing:", df$pass, "\\n")
cat("Pass rate:", round(sum(df$pass) / length(df$pass) * 100, 1), "%\\n")

# Return for table rendering
list(
  name = df$name,
  age = df$age,
  score = df$score,
  pass = df$pass,
  mean_score = round(mean(df$score), 1),
  mean_age = round(mean(df$age), 1)
)`;

const DEFAULT_STRINGS = `# === String Operations ===
cat("--- Strings ---\\n")
cat(paste("Hello", "World"), "\\n")
cat(paste("a", "b", "c", sep="-"), "\\n")
cat("nchar:", nchar("hello world"), "\\n")
cat("upper:", toupper("hello"), "\\n")
cat("trim:  ", trimws("  spaces  "), "\\n")
cat("gsub:  ", gsub("o", "0", "hello world"), "\\n")
cat("repeat:", strrep("ab", 4), "\\n")

words <- c("apple", "banana", "cherry", "avocado")
cat("starts 'a':", startsWith(words, "a"), "\\n")

# Split and rejoin
parts <- strsplit("one,two,three", ",")
cat("split:", parts[[1]], "\\n")
cat(sprintf("Pi is approximately %.4f\\n", pi))

# === Pipe Operator ===
cat("\\n--- Pipes ---\\n")
result <- c(5, 3, 8, 1, 9, 2, 7) |> sort() |> rev()
cat("sorted desc:", result, "\\n")

# === %in% Operator ===
cat("\\n--- Membership ---\\n")
fruits <- c("apple", "banana", "cherry")
test <- c("banana", "grape", "apple", "kiwi")
cat("test:  ", test, "\\n")
cat("fruits:", fruits, "\\n")
cat("%in%:  ", test %in% fruits, "\\n")

# === Apply Family ===
cat("\\n--- Apply ---\\n")
nums <- list(1, 4, 9, 16, 25)
roots <- sapply(nums, sqrt)
cat("sqrt:", roots, "\\n")

factorial_5 <- Reduce(function(a, b) a * b, 1:5, 1)
cat("5! =", factorial_5, "\\n")

big <- Filter(function(x) x > 10, c(3, 15, 7, 22, 9, 18))
cat("Filter >10:")
for (i in seq_along(big)) cat("", big[[i]])
cat("\\n")

# Vectorized ifelse
x <- c(1, -2, 3, -4, 5)
cat("abs via ifelse:", ifelse(x > 0, x, -x), "\\n")

# === Type Checking ===
cat("\\n--- Types ---\\n")
cat("is.numeric(42):  ", is.numeric(42), "\\n")
cat("is.character('x'):", is.character("x"), "\\n")
cat("is.logical(TRUE):", is.logical(TRUE), "\\n")
cat("identical(1, 1): ", identical(1, 1), "\\n")

list(result = "done", features_shown = 13)`;

const DEFAULT_TIMESERIES = `# Time series with moving average
# (data and window are injected from JS)

# Calculate moving average
ma <- function(x, n) {
  result <- c()
  len <- length(x)
  for (i in n:len) {
    w <- x[(i-n+1):i]
    result <- c(result, mean(w))
  }
  result
}

moving_avg <- ma(data, window)

cat("Points:", length(data), "\\n")
cat("MA window:", window, "\\n")
cat("Mean:", round(mean(data), 2), "\\n")
cat("SD:  ", round(sd(data), 2), "\\n")
cat("Range:", range(data), "\\n")

list(
  original = data,
  ma = moving_avg,
  mean = round(mean(data), 4),
  sd = round(sd(data), 4),
  min = round(min(data), 4),
  max = round(max(data), 4)
)`;

// ════════════════════════════════════════════
// DOM references
// ════════════════════════════════════════════

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

const statusEl = $("#status");

// ════════════════════════════════════════════
// Tab switching
// ════════════════════════════════════════════

$$('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    const demoId = tab.dataset.demo;
    $$('.tab').forEach(t => t.classList.remove('active'));
    $$('.demo-content').forEach(d => d.classList.remove('active'));
    tab.classList.add('active');
    $(`#demo-${demoId}`).classList.add('active');
  });
});

// ════════════════════════════════════════════
// Helpers
// ════════════════════════════════════════════

function randn() {
  let u = 0, v = 0;
  while (u === 0) u = Math.random();
  while (v === 0) v = Math.random();
  return Math.sqrt(-2.0 * Math.log(u)) * Math.cos(2.0 * Math.PI * v);
}

function toRVec(arr) {
  const parts = arr.map(v =>
    Number.isFinite(v) ? v.toFixed(6).replace(/0+$/, '').replace(/\.$/, '') : "NA"
  );
  return `c(${parts.join(",")})`;
}

function smallrEval(code) {
  if (typeof window.smallrEval !== "function") {
    throw new Error("WASM not loaded yet.");
  }
  const res = window.smallrEval(code);
  if (res && res.error) {
    throw new Error(res.error + (res.output ? `\n\nOutput:\n${res.output}` : ""));
  }
  return {
    json: res?.json ? JSON.parse(res.json) : null,
    output: res?.output || "",
    value: res?.value || "",
    rawJson: res?.json || ""
  };
}

function setStatus(msg, type = "info") {
  statusEl.textContent = msg;
  const colors = {
    info: "var(--muted)",
    ok: "rgba(110,231,255,.95)",
    error: "rgba(255,120,120,.95)",
    running: "rgba(251,191,36,.95)"
  };
  statusEl.style.color = colors[type] || colors.info;
}

// Tab support in textareas
document.addEventListener('keydown', (e) => {
  if (e.target.tagName === 'TEXTAREA' && e.key === 'Tab') {
    e.preventDefault();
    const ta = e.target;
    const start = ta.selectionStart;
    const end = ta.selectionEnd;
    ta.value = ta.value.substring(0, start) + "  " + ta.value.substring(end);
    ta.selectionStart = ta.selectionEnd = start + 2;
  }
});

// Ctrl/Cmd+Enter to run in any textarea
document.addEventListener('keydown', (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter' && e.target.tagName === 'TEXTAREA') {
    e.preventDefault();
    // Find the active demo's run button
    const active = $('.demo-content.active');
    if (active) {
      const btn = active.querySelector('.btn-run');
      if (btn) btn.click();
    }
  }
});

// ════════════════════════════════════════════
// D3 Tooltip
// ════════════════════════════════════════════

const tooltip = d3.select("body")
  .append("div")
  .attr("class", "tooltip");

// ════════════════════════════════════════════
// 1. PLAYGROUND
// ════════════════════════════════════════════

const playCodeEl = $("#playCode");
const playRunBtn = $("#playRun");
const playResetBtn = $("#playReset");
const playConsoleEl = $("#playConsole");
const playJSONEl = $("#playJSON");
const playValueEl = $("#playValue");
const playTimeEl = $("#playTime");

playCodeEl.value = DEFAULT_PLAYGROUND;

// Output tab switching
$$('.output-tab').forEach(tab => {
  tab.addEventListener('click', () => {
    const which = tab.dataset.out;
    $$('.output-tab').forEach(t => t.classList.remove('active'));
    tab.classList.add('active');
    playConsoleEl.classList.toggle('active', which === 'console');
    playJSONEl.classList.toggle('active', which === 'json');
    playValueEl.classList.toggle('active', which === 'value');
  });
});

playRunBtn.addEventListener('click', () => {
  try {
    setStatus("Running...", "running");
    const t0 = performance.now();
    const { json, output, value, rawJson } = smallrEval(playCodeEl.value);
    const dt = performance.now() - t0;

    playConsoleEl.textContent = output || "(no output)";
    playConsoleEl.classList.remove('error');
    playJSONEl.textContent = rawJson ? JSON.stringify(json, null, 2) : "(no JSON — return a named list)";
    playValueEl.textContent = value || "(no value)";
    playTimeEl.textContent = `${dt.toFixed(1)}ms`;

    setStatus("Done", "ok");
  } catch (e) {
    playConsoleEl.textContent = String(e.message || e);
    playConsoleEl.classList.add('error');
    playTimeEl.textContent = "";
    setStatus("Error", "error");
  }
});

playResetBtn.addEventListener('click', () => {
  playCodeEl.value = DEFAULT_PLAYGROUND;
  playConsoleEl.textContent = "";
  playJSONEl.textContent = "";
  playValueEl.textContent = "";
  playTimeEl.textContent = "";
  playConsoleEl.classList.remove('error');
});

// ════════════════════════════════════════════
// 2. REGRESSION
// ════════════════════════════════════════════

const codeEl = $("#code");
const outEl = $("#out");
const nEl = $("#n"); const noiseEl = $("#noise");
const trueSlopeEl = $("#trueSlope"); const trueInterceptEl = $("#trueIntercept");
const nVal = $("#nVal"); const noiseVal = $("#noiseVal");
const trueSlopeVal = $("#trueSlopeVal"); const trueInterceptVal = $("#trueInterceptVal");
const estSlopeEl = $("#estSlope"); const estInterceptEl = $("#estIntercept"); const r2El = $("#r2");
const regenBtn = $("#regen"); const runBtn = $("#run");

codeEl.value = DEFAULT_REGRESSION;

function generateData(n, slope, intercept, noise) {
  const x = [], y = [];
  for (let i = 0; i < n; i++) {
    const xv = Math.random() * 10;
    x.push(xv);
    y.push(intercept + slope * xv + randn() * noise);
  }
  return { x, y };
}

let currentReg = generateData(+nEl.value, +trueSlopeEl.value, +trueInterceptEl.value, +noiseEl.value);

function updateRegLabels() {
  nVal.textContent = nEl.value;
  noiseVal.textContent = noiseEl.value;
  trueSlopeVal.textContent = trueSlopeEl.value;
  trueInterceptVal.textContent = trueInterceptEl.value;
}

// D3 regression chart
const regSvg = d3.select("#chart");
const regVB = regSvg.attr("viewBox").split(" ").map(Number);
const regW = regVB[2], regH = regVB[3];
const regM = { top: 24, right: 24, bottom: 44, left: 56 };
const regIW = regW - regM.left - regM.right;
const regIH = regH - regM.top - regM.bottom;

const regG = regSvg.append("g").attr("transform", `translate(${regM.left},${regM.top})`);
const regXAxisG = regG.append("g").attr("transform", `translate(0,${regIH})`);
const regYAxisG = regG.append("g");
const regPointsG = regG.append("g");
const regLineG = regG.append("g");

function renderReg({ x, y, intercept, slope }) {
  const xExt = d3.extent(x);
  const yExt = d3.extent(y);
  const padY = (yExt[1] - yExt[0]) * 0.08 || 1;
  const xScale = d3.scaleLinear().domain(xExt).range([0, regIW]).nice();
  const yScale = d3.scaleLinear().domain([yExt[0] - padY, yExt[1] + padY]).range([regIH, 0]).nice();

  regXAxisG.transition().duration(300).call(d3.axisBottom(xScale));
  regYAxisG.transition().duration(300).call(d3.axisLeft(yScale));

  // Style axes
  regSvg.selectAll(".tick text").style("fill", "var(--muted)").style("font-size", "11px");
  regSvg.selectAll(".tick line").style("stroke", "rgba(255,255,255,0.08)");
  regSvg.selectAll(".domain").style("stroke", "rgba(255,255,255,0.1)");

  const data = x.map((xv, i) => ({ x: xv, y: y[i] }));

  const circles = regPointsG.selectAll("circle").data(data);
  circles.join(
    enter => enter.append("circle")
      .attr("r", 0)
      .attr("cx", d => xScale(d.x))
      .attr("cy", d => yScale(d.y))
      .style("fill", "var(--accent)")
      .style("opacity", 0.6)
      .on("mousemove", (event, d) => {
        tooltip.style("opacity", 1)
          .style("left", (event.clientX + 14) + "px")
          .style("top", (event.clientY - 10) + "px")
          .html(`x = ${d.x.toFixed(3)}<br>y = ${d.y.toFixed(3)}<br>ŷ = ${(intercept + slope * d.x).toFixed(3)}`);
      })
      .on("mouseleave", () => tooltip.style("opacity", 0))
      .transition().duration(400)
      .attr("r", 3),
    update => update.transition().duration(300)
      .attr("cx", d => xScale(d.x))
      .attr("cy", d => yScale(d.y)),
    exit => exit.transition().duration(200).attr("r", 0).remove()
  );

  // Regression line
  const x0 = xScale.domain()[0], x1 = xScale.domain()[1];
  const lineData = [
    { x: x0, y: intercept + slope * x0 },
    { x: x1, y: intercept + slope * x1 }
  ];
  const linePath = d3.line().x(d => xScale(d.x)).y(d => yScale(d.y));

  const lineSel = regLineG.selectAll("path").data([lineData]);
  lineSel.join(
    enter => enter.append("path")
      .attr("fill", "none")
      .attr("stroke", "rgba(255,255,255,0.9)")
      .attr("stroke-width", 2.5)
      .attr("stroke-dasharray", function() { return this.getTotalLength(); })
      .attr("stroke-dashoffset", function() { return this.getTotalLength(); })
      .attr("d", linePath)
      .transition().duration(600)
      .attr("stroke-dashoffset", 0),
    update => update.transition().duration(300).attr("d", linePath)
      .attr("stroke-dasharray", "none")
      .attr("stroke-dashoffset", 0)
  );

  // Axis labels
  regSvg.selectAll(".xlabel").data([0]).join("text")
    .attr("class", "xlabel")
    .attr("x", regM.left + regIW / 2).attr("y", regH - 8)
    .attr("text-anchor", "middle")
    .style("fill", "var(--muted)").style("font-size", "12px")
    .text("x");
  regSvg.selectAll(".ylabel").data([0]).join("text")
    .attr("class", "ylabel")
    .attr("transform", `translate(16,${regM.top + regIH / 2}) rotate(-90)`)
    .attr("text-anchor", "middle")
    .style("fill", "var(--muted)").style("font-size", "12px")
    .text("y");
}

function recomputeReg() {
  setStatus("Running...", "running");
  updateRegLabels();
  const prelude = `x <- ${toRVec(currentReg.x)}\ny <- ${toRVec(currentReg.y)}\n`;
  const { json, output } = smallrEval(prelude + codeEl.value);

  outEl.textContent = (output || "").trim();

  if (!json || typeof json !== "object") {
    throw new Error("Return a named list with intercept, slope, r2.");
  }

  const intercept = +json.intercept;
  const slope = +json.slope;
  const r2 = +json.r2;

  estInterceptEl.textContent = Number.isFinite(intercept) ? intercept.toFixed(4) : "NA";
  estSlopeEl.textContent = Number.isFinite(slope) ? slope.toFixed(4) : "NA";
  r2El.textContent = Number.isFinite(r2) ? r2.toFixed(4) : "NA";

  renderReg({ x: currentReg.x, y: currentReg.y, intercept, slope });
  setStatus("Done", "ok");
}

regenBtn.addEventListener("click", () => {
  try {
    currentReg = generateData(+nEl.value, +trueSlopeEl.value, +trueInterceptEl.value, +noiseEl.value);
    recomputeReg();
  } catch (e) { setStatus(e.message, "error"); }
});

runBtn.addEventListener("click", () => {
  try { recomputeReg(); } catch (e) { setStatus(e.message, "error"); }
});

let regTimer = null;
[nEl, noiseEl, trueSlopeEl, trueInterceptEl].forEach(el => {
  el.addEventListener("input", () => {
    clearTimeout(regTimer);
    regTimer = setTimeout(() => {
      try {
        currentReg = generateData(+nEl.value, +trueSlopeEl.value, +trueInterceptEl.value, +noiseEl.value);
        recomputeReg();
      } catch (e) { setStatus(e.message, "error"); }
    }, 100);
  });
});

// ════════════════════════════════════════════
// 3. STATISTICS
// ════════════════════════════════════════════

const codeStatsEl = $("#codeStats");
codeStatsEl.value = DEFAULT_STATS;

$("#runStats").addEventListener("click", () => {
  try {
    setStatus("Running...", "running");
    const { json, output } = smallrEval(codeStatsEl.value);
    $("#statsOut1").textContent = output || "(no output)";

    if (json && json.values) {
      renderStatsChart(json.values, json.labels, json.mean, json.sd);
    }
    setStatus("Statistics done", "ok");
  } catch (e) {
    $("#statsOut1").textContent = String(e.message || e);
    setStatus("Error", "error");
  }
});

function renderStatsChart(values, labels, mean, sd) {
  const svg = d3.select("#statsChart");
  svg.selectAll("*").remove();

  const vb = svg.attr("viewBox").split(" ").map(Number);
  const W = vb[2], H = vb[3];
  const m = { top: 20, right: 16, bottom: 36, left: 40 };
  const iW = W - m.left - m.right;
  const iH = H - m.top - m.bottom;

  const g = svg.append("g").attr("transform", `translate(${m.left},${m.top})`);

  const xScale = d3.scaleBand()
    .domain(labels || values.map((_, i) => `x${i + 1}`))
    .range([0, iW])
    .padding(0.25);

  const yScale = d3.scaleLinear()
    .domain([0, d3.max(values) * 1.15])
    .range([iH, 0]).nice();

  // Grid
  g.selectAll(".grid-line").data(yScale.ticks(5)).join("line")
    .attr("class", "grid-line")
    .attr("x1", 0).attr("x2", iW)
    .attr("y1", d => yScale(d)).attr("y2", d => yScale(d))
    .style("stroke", "rgba(255,255,255,0.05)");

  // Mean line
  g.append("line")
    .attr("x1", 0).attr("x2", iW)
    .attr("y1", yScale(mean)).attr("y2", yScale(mean))
    .style("stroke", "var(--orange)")
    .style("stroke-width", 1.5)
    .style("stroke-dasharray", "6,4");

  g.append("text")
    .attr("x", iW - 4).attr("y", yScale(mean) - 6)
    .attr("text-anchor", "end")
    .style("fill", "var(--orange)").style("font-size", "10px").style("font-weight", "600")
    .text(`μ = ${mean.toFixed(1)}`);

  // Bars
  g.selectAll("rect").data(values).join("rect")
    .attr("x", (_, i) => xScale(labels ? labels[i] : `x${i + 1}`))
    .attr("width", xScale.bandwidth())
    .attr("y", iH)
    .attr("height", 0)
    .attr("rx", 4)
    .style("fill", (d) => d >= mean ? "var(--accent)" : "var(--accent2)")
    .style("opacity", 0.75)
    .on("mousemove", (event, d) => {
      tooltip.style("opacity", 1)
        .style("left", (event.clientX + 14) + "px")
        .style("top", (event.clientY - 10) + "px")
        .html(`Value: <b>${d}</b><br>vs mean: ${(d - mean).toFixed(1)}`);
    })
    .on("mouseleave", () => tooltip.style("opacity", 0))
    .transition().duration(500).delay((_, i) => i * 40)
    .attr("y", d => yScale(d))
    .attr("height", d => iH - yScale(d));

  // Axes
  g.append("g").attr("transform", `translate(0,${iH})`)
    .call(d3.axisBottom(xScale).tickSize(0))
    .selectAll("text").style("fill", "var(--muted)").style("font-size", "10px");
  g.append("g")
    .call(d3.axisLeft(yScale).ticks(5).tickSize(-4))
    .selectAll("text").style("fill", "var(--muted)").style("font-size", "10px");
  g.selectAll(".domain").style("stroke", "rgba(255,255,255,0.1)");
  g.selectAll(".tick line").style("stroke", "rgba(255,255,255,0.08)");
}

// ════════════════════════════════════════════
// 4. DATA FRAMES
// ════════════════════════════════════════════

const codeDFEl = $("#codeDataFrame");
codeDFEl.value = DEFAULT_DATAFRAME;

$("#runDataFrame").addEventListener("click", () => {
  try {
    setStatus("Running...", "running");
    const { json, output } = smallrEval(codeDFEl.value);
    $("#dfOut1").textContent = output || "(no output)";

    if (json && json.name) {
      renderDFTable(json);
    }
    setStatus("DataFrame done", "ok");
  } catch (e) {
    $("#dfOut1").textContent = String(e.message || e);
    setStatus("Error", "error");
  }
});

function renderDFTable(data) {
  const wrap = $("#dfTable");
  wrap.innerHTML = "";

  const names = Array.isArray(data.name) ? data.name : [data.name];
  const ages = Array.isArray(data.age) ? data.age : [data.age];
  const scores = Array.isArray(data.score) ? data.score : [data.score];
  const pass = Array.isArray(data.pass) ? data.pass : [data.pass];
  const n = names.length;

  let html = '<table><thead><tr>';
  html += '<th>#</th><th>Name</th><th>Age</th><th>Score</th><th>Pass</th>';
  html += '</tr></thead><tbody>';

  for (let i = 0; i < n; i++) {
    const p = pass[i];
    html += '<tr>';
    html += `<td class="num">${i + 1}</td>`;
    html += `<td>${names[i]}</td>`;
    html += `<td class="num">${ages[i]}</td>`;
    html += `<td class="num">${scores[i]}</td>`;
    html += `<td class="${p ? 'bool-true' : 'bool-false'}">${p ? 'TRUE' : 'FALSE'}</td>`;
    html += '</tr>';
  }

  html += '</tbody></table>';

  // Summary row
  html += `<div style="padding:8px 12px;font-size:12px;color:var(--muted)">`;
  html += `Mean Score: <b style="color:var(--accent)">${data.mean_score}</b> &middot; `;
  html += `Mean Age: <b style="color:var(--accent)">${data.mean_age}</b>`;
  html += '</div>';

  wrap.innerHTML = html;
}

// ════════════════════════════════════════════
// 5. STRINGS & FUNCTIONAL
// ════════════════════════════════════════════

const codeStrEl = $("#codeStrings");
codeStrEl.value = DEFAULT_STRINGS;

$("#runStrings").addEventListener("click", () => {
  try {
    setStatus("Running...", "running");
    const { output } = smallrEval(codeStrEl.value);
    $("#stringsOut").textContent = output || "(no output)";
    setStatus("Strings & Functional done", "ok");
  } catch (e) {
    $("#stringsOut").textContent = String(e.message || e);
    setStatus("Error", "error");
  }
});

// ════════════════════════════════════════════
// 6. TIME SERIES
// ════════════════════════════════════════════

const codeTSEl = $("#codeTimeSeries");
codeTSEl.value = DEFAULT_TIMESERIES;

const tsPointsEl = $("#tsPoints");
const tsWindowEl = $("#tsWindow");
const tsPointsVal = $("#tsPointsVal");
const tsWindowVal = $("#tsWindowVal");

function updateTSLabels() {
  tsPointsVal.textContent = tsPointsEl.value;
  tsWindowVal.textContent = tsWindowEl.value;
}
tsPointsEl?.addEventListener('input', updateTSLabels);
tsWindowEl?.addEventListener('input', updateTSLabels);

$("#runTimeSeries").addEventListener("click", () => {
  try {
    updateTSLabels();
    setStatus("Running...", "running");
    const n = +tsPointsEl.value;
    const win = +tsWindowEl.value;

    // Generate synthetic time series
    const data = [];
    let val = 50;
    for (let i = 0; i < n; i++) {
      val += (Math.random() - 0.5) * 5 + Math.sin(i / 10) * 3;
      data.push(val);
    }

    const prelude = `data <- ${toRVec(data)}\nwindow <- ${win}\n`;
    const code = prelude + codeTSEl.value;
    const { json } = smallrEval(code);

    if (json && json.original) {
      renderTimeSeries(json.original, json.ma, win);
      renderTSStats(json);
    }
    setStatus("Time series done", "ok");
  } catch (e) {
    setStatus(e.message, "error");
  }
});

function renderTSStats(data) {
  const el = $("#tsStats");
  const items = [
    { label: "Mean", value: data.mean, color: "var(--accent)" },
    { label: "Std Dev", value: data.sd, color: "var(--accent2)" },
    { label: "Min", value: data.min, color: "var(--green)" },
    { label: "Max", value: data.max, color: "var(--orange)" },
  ];
  el.innerHTML = items.map(item => `
    <div class="ts-stat-card">
      <div class="k">${item.label}</div>
      <div class="v" style="color:${item.color}">${
        typeof item.value === 'number' ? item.value.toFixed(2) : item.value
      }</div>
    </div>
  `).join('');
}

function renderTimeSeries(original, ma, windowSize) {
  const svg = d3.select("#tsChart");
  svg.selectAll("*").remove();

  const vb = svg.attr("viewBox").split(" ").map(Number);
  const W = vb[2], H = vb[3];
  const m = { top: 20, right: 20, bottom: 40, left: 50 };
  const iW = W - m.left - m.right;
  const iH = H - m.top - m.bottom;

  const g = svg.append("g").attr("transform", `translate(${m.left},${m.top})`);

  const xScale = d3.scaleLinear().domain([0, original.length - 1]).range([0, iW]);
  const allValues = [...original, ...ma];
  const yScale = d3.scaleLinear()
    .domain([d3.min(allValues) - 2, d3.max(allValues) + 2])
    .range([iH, 0]).nice();

  // Grid
  g.selectAll(".grid").data(yScale.ticks(6)).join("line")
    .attr("x1", 0).attr("x2", iW)
    .attr("y1", d => yScale(d)).attr("y2", d => yScale(d))
    .style("stroke", "rgba(255,255,255,0.04)");

  // Axes
  g.append("g").attr("transform", `translate(0,${iH})`)
    .call(d3.axisBottom(xScale).ticks(8))
    .selectAll("text").style("fill", "var(--muted)").style("font-size", "11px");
  g.append("g")
    .call(d3.axisLeft(yScale).ticks(6))
    .selectAll("text").style("fill", "var(--muted)").style("font-size", "11px");
  g.selectAll(".domain").style("stroke", "rgba(255,255,255,0.1)");
  g.selectAll(".tick line").style("stroke", "rgba(255,255,255,0.06)");

  const line = d3.line().x((d, i) => xScale(i)).y(d => yScale(d)).curve(d3.curveMonotoneX);

  // Area fill
  const area = d3.area()
    .x((d, i) => xScale(i))
    .y0(iH)
    .y1(d => yScale(d))
    .curve(d3.curveMonotoneX);

  g.append("path")
    .datum(original)
    .attr("fill", "rgba(110,231,255,0.06)")
    .attr("d", area);

  // Original line
  const origPath = g.append("path")
    .datum(original)
    .attr("fill", "none")
    .attr("stroke", "rgba(110,231,255,0.4)")
    .attr("stroke-width", 1.5)
    .attr("d", line);

  // Animate original
  const origLen = origPath.node().getTotalLength();
  origPath
    .attr("stroke-dasharray", origLen)
    .attr("stroke-dashoffset", origLen)
    .transition().duration(800).ease(d3.easeQuadOut)
    .attr("stroke-dashoffset", 0);

  // MA line
  const maLine = d3.line()
    .x((d, i) => xScale(i + windowSize - 1))
    .y(d => yScale(d))
    .curve(d3.curveMonotoneX);

  const maPath = g.append("path")
    .datum(ma)
    .attr("fill", "none")
    .attr("stroke", "var(--orange)")
    .attr("stroke-width", 2.5)
    .attr("d", maLine);

  const maLen = maPath.node().getTotalLength();
  maPath
    .attr("stroke-dasharray", maLen)
    .attr("stroke-dashoffset", maLen)
    .transition().duration(800).delay(300).ease(d3.easeQuadOut)
    .attr("stroke-dashoffset", 0);

  // Hover dots
  const hoverG = g.append("g");
  const bisect = d3.bisector(d => d).left;
  const overlay = g.append("rect")
    .attr("width", iW).attr("height", iH)
    .style("fill", "none").style("pointer-events", "all");

  overlay.on("mousemove", (event) => {
    const [mx] = d3.pointer(event);
    const i = Math.round(xScale.invert(mx));
    if (i < 0 || i >= original.length) return;

    hoverG.selectAll("*").remove();
    hoverG.append("circle")
      .attr("cx", xScale(i)).attr("cy", yScale(original[i]))
      .attr("r", 4).style("fill", "var(--accent)");

    let html = `<b>t=${i}</b><br>Value: ${original[i].toFixed(2)}`;
    if (i >= windowSize - 1) {
      const maIdx = i - windowSize + 1;
      if (maIdx < ma.length) {
        hoverG.append("circle")
          .attr("cx", xScale(i)).attr("cy", yScale(ma[maIdx]))
          .attr("r", 4).style("fill", "var(--orange)");
        html += `<br>MA: ${ma[maIdx].toFixed(2)}`;
      }
    }

    tooltip.style("opacity", 1)
      .style("left", (event.clientX + 14) + "px")
      .style("top", (event.clientY - 10) + "px")
      .html(html);
  })
  .on("mouseleave", () => {
    hoverG.selectAll("*").remove();
    tooltip.style("opacity", 0);
  });

  // Legend
  const leg = g.append("g").attr("transform", `translate(${iW - 160}, 8)`);

  leg.append("rect").attr("width", 155).attr("height", 46).attr("rx", 8)
    .style("fill", "rgba(10,14,22,0.7)").style("stroke", "rgba(255,255,255,0.08)");

  leg.append("line").attr("x1", 10).attr("x2", 35).attr("y1", 16).attr("y2", 16)
    .style("stroke", "rgba(110,231,255,0.5)").style("stroke-width", 1.5);
  leg.append("text").attr("x", 42).attr("y", 20)
    .style("font-size", "11px").style("fill", "var(--muted)").text("Original");

  leg.append("line").attr("x1", 10).attr("x2", 35).attr("y1", 34).attr("y2", 34)
    .style("stroke", "var(--orange)").style("stroke-width", 2.5);
  leg.append("text").attr("x", 42).attr("y", 38)
    .style("font-size", "11px").style("fill", "var(--muted)").text(`MA(${windowSize})`);
}

// ════════════════════════════════════════════
// WASM Boot
// ════════════════════════════════════════════

async function boot() {
  setStatus("Loading WASM...", "info");

  const go = new Go();
  const resp = await fetch("./smallr.wasm");
  if (!resp.ok) {
    throw new Error("Failed to fetch smallr.wasm. Run `make` first.");
  }
  const bytes = await resp.arrayBuffer();
  const { instance } = await WebAssembly.instantiate(bytes, go.importObject);

  go.run(instance);

  setStatus("Ready", "ok");

  // Auto-run the regression demo
  updateRegLabels();
  recomputeReg();
}

boot().catch(err => {
  setStatus(err.message || String(err), "error");
});
