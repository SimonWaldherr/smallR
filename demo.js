/* smallR interactive demos
   - Multiple demos showcasing different R features
   - All computation runs in WASM (no server)
   - Visualizations with D3
*/

// --- Tab switching ---
document.querySelectorAll('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    const demoId = tab.dataset.demo;
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.demo-content').forEach(d => d.classList.remove('active'));
    tab.classList.add('active');
    document.getElementById(`demo-${demoId}`).classList.add('active');
  });
});

const statusEl = document.getElementById("status");
const codeEl = document.getElementById("code");
const outEl = document.getElementById("out");

const nEl = document.getElementById("n");
const noiseEl = document.getElementById("noise");
const trueSlopeEl = document.getElementById("trueSlope");
const trueInterceptEl = document.getElementById("trueIntercept");

const nVal = document.getElementById("nVal");
const noiseVal = document.getElementById("noiseVal");
const trueSlopeVal = document.getElementById("trueSlopeVal");
const trueInterceptVal = document.getElementById("trueInterceptVal");

const estSlopeEl = document.getElementById("estSlope");
const estInterceptEl = document.getElementById("estIntercept");
const r2El = document.getElementById("r2");

const regenBtn = document.getElementById("regen");
const runBtn = document.getElementById("run");

codeEl.value = `# You already have vectors x and y (generated in JS).
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

# smallR returns a named list → JSON object (great for JS)
list(
  intercept = intercept,
  slope = slope,
  r2 = r2,
  yhat = yhat
)
`; // replaced at build-time in JS below

// The line above is replaced; keep a fallback:
if (!codeEl.value.trim()) {
  codeEl.value = `mx <- mean(x)
my <- mean(y)
list(mx=mx,my=my)`;
}

function escapeBackticks(s) {
  return s.replaceAll("`", "\`");
}

// --- Data generation (browser-native) ---

function randn() {
  // Box-Muller
  let u = 0, v = 0;
  while (u === 0) u = Math.random();
  while (v === 0) v = Math.random();
  return Math.sqrt(-2.0 * Math.log(u)) * Math.cos(2.0 * Math.PI * v);
}

function generateData(n, slope, intercept, noise) {
  const x = new Array(n);
  const y = new Array(n);
  for (let i = 0; i < n; i++) {
    const xv = Math.random() * 10; // wider x-range for nicer plots
    const eps = randn() * noise;
    x[i] = xv;
    y[i] = intercept + slope * xv + eps;
  }
  return { x, y };
}

// --- smallR bridge ---

function toRVec(arr) {
  // keep it compact; R-like numeric literal
  const parts = arr.map(v => (Number.isFinite(v) ? v.toFixed(6).replace(/0+$/,'').replace(/\.$/,'') : "NA"));
  return `c(${parts.join(",")})`;
}

function runSmallR(x, y, userCode) {
  if (typeof window.smallrEval !== "function") {
    throw new Error("smallrEval not available (WASM not loaded yet).");
  }
  const prelude = `x <- ${toRVec(x)}\ny <- ${toRVec(y)}\n`;
  const code = prelude + userCode;
  const res = window.smallrEval(code);
  if (res && res.error) {
    throw new Error(res.error + (res.output ? `\n\nOutput:\n${res.output}` : ""));
  }
  const json = res && res.json ? JSON.parse(res.json) : null;
  return { json, output: (res && res.output) || "", value: (res && res.value) || "" };
}

// --- D3 chart ---

const svg = d3.select("#chart");
const W = parseFloat(svg.attr("width"));
const H = parseFloat(svg.attr("height"));
const margin = { top: 22, right: 22, bottom: 44, left: 56 };
const innerW = W - margin.left - margin.right;
const innerH = H - margin.top - margin.bottom;

const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

const xAxisG = g.append("g").attr("transform", `translate(0,${innerH})`);
const yAxisG = g.append("g");

const pointsG = g.append("g");
const lineG = g.append("g");

const tooltip = d3.select("body")
  .append("div")
  .style("position", "fixed")
  .style("pointer-events", "none")
  .style("opacity", 0)
  .style("padding", "8px 10px")
  .style("border-radius", "12px")
  .style("border", "1px solid rgba(255,255,255,.12)")
  .style("background", "rgba(0,0,0,.6)")
  .style("backdrop-filter", "blur(8px)")
  .style("font", "12px ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace")
  .style("color", "white");

function render({ x, y, intercept, slope }) {
  const xExt = d3.extent(x);
  const yExt = d3.extent(y);
  const padY = (yExt[1] - yExt[0]) * 0.08 || 1;

  const xScale = d3.scaleLinear().domain(xExt).range([0, innerW]).nice();
  const yScale = d3.scaleLinear().domain([yExt[0] - padY, yExt[1] + padY]).range([innerH, 0]).nice();

  xAxisG.call(d3.axisBottom(xScale));
  yAxisG.call(d3.axisLeft(yScale));

  // Points
  const data = x.map((xv, i) => ({ x: xv, y: y[i] }));
  const sel = pointsG.selectAll("circle").data(data);
  sel.enter().append("circle")
    .attr("r", 3)
    .attr("cx", d => xScale(d.x))
    .attr("cy", d => yScale(d.y))
    .style("opacity", 0.8)
    .on("mousemove", (event, d) => {
      tooltip
        .style("opacity", 1)
        .style("left", (event.clientX + 12) + "px")
        .style("top", (event.clientY + 12) + "px")
        .html(`x=${d.x.toFixed(3)}<br/>y=${d.y.toFixed(3)}<br/>ŷ=${(intercept + slope * d.x).toFixed(3)}`);
    })
    .on("mouseleave", () => tooltip.style("opacity", 0));
  sel
    .attr("cx", d => xScale(d.x))
    .attr("cy", d => yScale(d.y));
  sel.exit().remove();

  // Regression line from fitted coeffs
  const x0 = xScale.domain()[0];
  const x1 = xScale.domain()[1];
  const y0 = intercept + slope * x0;
  const y1 = intercept + slope * x1;

  const line = [{ x: x0, y: y0 }, { x: x1, y: y1 }];

  const lineSel = lineG.selectAll("path").data([line]);
  lineSel.enter().append("path")
    .attr("fill", "none")
    .attr("stroke-width", 2)
    .attr("stroke", "rgba(255,255,255,.95)")
    .attr("d", d3.line().x(d => xScale(d.x)).y(d => yScale(d.y)));
  lineSel
    .attr("d", d3.line().x(d => xScale(d.x)).y(d => yScale(d.y)));
  lineSel.exit().remove();

  // Axes labels
  svg.selectAll(".xlabel").data([0]).join("text")
    .attr("class","xlabel")
    .attr("x", margin.left + innerW/2)
    .attr("y", H - 10)
    .attr("text-anchor","middle")
    .style("fill","rgba(255,255,255,.7)")
    .style("font-size","12px")
    .text("x");

  svg.selectAll(".ylabel").data([0]).join("text")
    .attr("class","ylabel")
    .attr("transform", `translate(14,${margin.top + innerH/2}) rotate(-90)`)
    .attr("text-anchor","middle")
    .style("fill","rgba(255,255,255,.7)")
    .style("font-size","12px")
    .text("y");
}

// --- App wiring ---

let current = generateData(+nEl.value, +trueSlopeEl.value, +trueInterceptEl.value, +noiseEl.value);

function updateLabels() {
  nVal.textContent = nEl.value;
  noiseVal.textContent = noiseEl.value;
  trueSlopeVal.textContent = trueSlopeEl.value;
  trueInterceptVal.textContent = trueInterceptEl.value;
}

async function recompute() {
  updateLabels();
  const userCode = codeEl.value;
  const { json, output } = runSmallR(current.x, current.y, userCode);

  outEl.textContent = (output || "").trim();

  if (!json || typeof json !== "object") {
    throw new Error("smallR returned no JSON object (did you return a named list?)");
  }

  const intercept = +json.intercept;
  const slope = +json.slope;
  const r2 = +json.r2;

  estInterceptEl.textContent = Number.isFinite(intercept) ? intercept.toFixed(4) : "NA";
  estSlopeEl.textContent = Number.isFinite(slope) ? slope.toFixed(4) : "NA";
  r2El.textContent = Number.isFinite(r2) ? r2.toFixed(4) : "NA";

  render({ x: current.x, y: current.y, intercept, slope });
}

function regenerate() {
  current = generateData(+nEl.value, +trueSlopeEl.value, +trueInterceptEl.value, +noiseEl.value);
}

// Debounce slider-driven changes
let t = null;
function debounceRecompute() {
  clearTimeout(t);
  t = setTimeout(() => {
    try {
      regenerate();
      recompute();
    } catch (e) {
      statusEl.textContent = String(e.message || e);
      statusEl.style.color = "rgba(255,120,120,.95)";
    }
  }, 80);
}

regenBtn.addEventListener("click", () => {
  try {
    regenerate();
    recompute();
  } catch (e) {
    statusEl.textContent = String(e.message || e);
    statusEl.style.color = "rgba(255,120,120,.95)";
  }
});

runBtn.addEventListener("click", () => {
  try {
    recompute();
  } catch (e) {
    statusEl.textContent = String(e.message || e);
    statusEl.style.color = "rgba(255,120,120,.95)";
  }
});

[nEl, noiseEl, trueSlopeEl, trueInterceptEl].forEach(el => {
  el.addEventListener("input", debounceRecompute);
});

codeEl.addEventListener("keydown", (e) => {
  // Cmd/Ctrl+Enter to run
  if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
    e.preventDefault();
    runBtn.click();
  }
});

// --- Demo 2: Statistics & Vectors ---
document.getElementById('runStats')?.addEventListener('click', () => {
  try {
    const code = document.getElementById('codeStats').value;
    const res = window.smallrEval(code);
    if (res && res.error) {
      throw new Error(res.error);
    }
    
    const out1 = document.getElementById('statsOut1');
    const out2 = document.getElementById('statsOut2');
    
    out1.textContent = res.output || res.value || '';
    
    if (res.json) {
      const data = JSON.parse(res.json);
      out2.textContent = `Mean: ${data.mean}\nSum: ${data.sum}\nLength: ${data.length}`;
    }
    
    statusEl.textContent = "Statistics demo complete ✓";
    statusEl.style.color = "rgba(110,231,255,.95)";
  } catch (e) {
    statusEl.textContent = String(e.message || e);
    statusEl.style.color = "rgba(255,120,120,.95)";
  }
});

// --- Demo 3: Data Frames ---
document.getElementById('runDataFrame')?.addEventListener('click', () => {
  try {
    const code = document.getElementById('codeDataFrame').value;
    const res = window.smallrEval(code);
    if (res && res.error) {
      throw new Error(res.error);
    }
    
    const out1 = document.getElementById('dfOut1');
    const out2 = document.getElementById('dfOut2');
    
    out1.textContent = res.output || '';
    
    if (res.json) {
      const data = JSON.parse(res.json);
      out2.textContent = `Dimensions: ${data.dimensions?.[0]} x ${data.dimensions?.[1]}\nMean Score: ${data.mean_score}\nMean Age: ${data.mean_age}`;
    }
    
    statusEl.textContent = "DataFrame demo complete ✓";
    statusEl.style.color = "rgba(110,231,255,.95)";
  } catch (e) {
    statusEl.textContent = String(e.message || e);
    statusEl.style.color = "rgba(255,120,120,.95)";
  }
});

// --- Demo 4: Time Series ---
const tsPointsEl = document.getElementById('tsPoints');
const tsWindowEl = document.getElementById('tsWindow');
const tsPointsVal = document.getElementById('tsPointsVal');
const tsWindowVal = document.getElementById('tsWindowVal');

function updateTSLabels() {
  tsPointsVal.textContent = tsPointsEl.value;
  tsWindowVal.textContent = tsWindowEl.value;
}

tsPointsEl?.addEventListener('input', updateTSLabels);
tsWindowEl?.addEventListener('input', updateTSLabels);

document.getElementById('runTimeSeries')?.addEventListener('click', () => {
  try {
    updateTSLabels();
    const n = +tsPointsEl.value;
    const windowSize = +tsWindowEl.value;
    
    // Generate synthetic time series data
    const data = [];
    let val = 50;
    for (let i = 0; i < n; i++) {
      val += (Math.random() - 0.5) * 5 + Math.sin(i / 10) * 3;
      data.push(val);
    }
    
    const codeBase = document.getElementById('codeTimeSeries').value;
    const code = codeBase
      .replace('# Data generated in JavaScript', `data <- ${toRVec(data)}`)
      .replace('moving_avg <- ma(data, window)', `moving_avg <- ma(data, ${windowSize})`);
    
    const res = window.smallrEval(code);
    if (res && res.error) {
      throw new Error(res.error);
    }
    
    if (res.json) {
      const result = JSON.parse(res.json);
      renderTimeSeries(result.original, result.ma, windowSize);
    }
    
    statusEl.textContent = "Time series analysis complete ✓";
    statusEl.style.color = "rgba(110,231,255,.95)";
  } catch (e) {
    statusEl.textContent = String(e.message || e);
    statusEl.style.color = "rgba(255,120,120,.95)";
  }
});

function renderTimeSeries(original, ma, windowSize) {
  const svg = d3.select("#tsChart");
  svg.selectAll("*").remove();
  
  const W = 900;
  const H = 400;
  const margin = { top: 20, right: 20, bottom: 40, left: 50 };
  const innerW = W - margin.left - margin.right;
  const innerH = H - margin.top - margin.bottom;
  
  const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);
  
  const xScale = d3.scaleLinear()
    .domain([0, original.length - 1])
    .range([0, innerW]);
  
  const allValues = [...original, ...ma];
  const yScale = d3.scaleLinear()
    .domain([d3.min(allValues) - 2, d3.max(allValues) + 2])
    .range([innerH, 0]);
  
  // Axes
  g.append("g")
    .attr("transform", `translate(0,${innerH})`)
    .call(d3.axisBottom(xScale));
  
  g.append("g")
    .call(d3.axisLeft(yScale));
  
  // Original data line
  const line = d3.line()
    .x((d, i) => xScale(i))
    .y(d => yScale(d));
  
  g.append("path")
    .datum(original)
    .attr("fill", "none")
    .attr("stroke", "rgba(110,231,255,.4)")
    .attr("stroke-width", 1.5)
    .attr("d", line);
  
  // Moving average line
  const maLine = d3.line()
    .x((d, i) => xScale(i + windowSize - 1))
    .y(d => yScale(d));
  
  g.append("path")
    .datum(ma)
    .attr("fill", "none")
    .attr("stroke", "rgba(255,231,110,.95)")
    .attr("stroke-width", 2)
    .attr("d", maLine);
  
  // Legend
  const legend = g.append("g")
    .attr("transform", `translate(${innerW - 150}, 10)`);
  
  legend.append("line")
    .attr("x1", 0).attr("x2", 30)
    .attr("stroke", "rgba(110,231,255,.4)")
    .attr("stroke-width", 1.5);
  legend.append("text")
    .attr("x", 35).attr("y", 5)
    .style("font-size", "12px")
    .style("fill", "rgba(255,255,255,.7)")
    .text("Original");
  
  legend.append("line")
    .attr("x1", 0).attr("x2", 30)
    .attr("y1", 20).attr("y2", 20)
    .attr("stroke", "rgba(255,231,110,.95)")
    .attr("stroke-width", 2);
  legend.append("text")
    .attr("x", 35).attr("y", 25)
    .style("font-size", "12px")
    .style("fill", "rgba(255,255,255,.7)")
    .text(`MA(${windowSize})`);
}

// --- WASM boot ---

async function boot() {
  statusEl.textContent = "Loading WASM…";
  statusEl.style.color = "rgba(255,255,255,.75)";

  const go = new Go();

  // NOTE: `smallr.wasm` must be present next to this file (or adjust the path).
  const resp = await fetch("./smallr.wasm");
  if (!resp.ok) {
    throw new Error("Failed to fetch smallr.wasm. Did you build it? See demo/README.md");
  }
  const bytes = await resp.arrayBuffer();
  const { instance } = await WebAssembly.instantiate(bytes, go.importObject);

  // go.run installs globalThis.smallrEval
  go.run(instance);

  statusEl.textContent = "WASM ready ✓";
  statusEl.style.color = "rgba(110,231,255,.95)";

  updateLabels();
  regenerate();
  recompute();
}

boot().catch(err => {
  statusEl.textContent = String(err.message || err);
  statusEl.style.color = "rgba(255,120,120,.95)";
});

// --- embed default code without breaking JS template strings ---
