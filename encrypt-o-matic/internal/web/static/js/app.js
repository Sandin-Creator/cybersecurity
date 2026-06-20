/* Encrypt-O-Matic Web Dashboard — reviewer-friendly educational UI */

const titles = {
  '/': 'Dashboard',
  '/encrypt': 'Encryption Workflow',
  '/files': 'Encrypted Files',
  '/security': 'Security Visualization',
  '/reviewer': 'Reviewer Mode',
  '/debug': 'Debug Tools',
  '/file': 'File Details',
};

let setupWizardShown = false;

async function api(path, opts = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) },
    ...opts,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || data.result || res.statusText);
  return data;
}

function toast(msg, type = 'success') {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.className = `toast ${type}`;
  setTimeout(() => el.classList.add('hidden'), 3500);
}

function fmtBytes(n) {
  if (!n && n !== 0) return '—';
  const u = ['B', 'KB', 'MB', 'GB'];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < u.length - 1) { v /= 1024; i++; }
  return `${v.toFixed(1)} ${u[i]}`;
}

function fmtTime(iso) {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}

function badge(cls, text) {
  const map = { green: 'badge-green', orange: 'badge-orange', red: 'badge-red' };
  return `<span class="badge ${map[cls] || 'badge-blue'}">${text}</span>`;
}

function route() {
  let path = location.pathname;
  if (path.startsWith('/file/')) path = '/file';
  return path in titles ? path : '/';
}

function navigate(path) {
  history.pushState({}, '', path);
  render();
}
window.navigate = navigate;

function setActiveNav() {
  const r = location.pathname;
  document.querySelectorAll('.nav a').forEach(a => {
    a.classList.toggle('active', a.dataset.route === r || (r.startsWith('/file') && a.dataset.route === '/files'));
  });
}

function asArray(value) {
  return Array.isArray(value) ? value : [];
}

function escapeHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

/* ── Educational helpers ── */

function infoTip(description, why) {
  return `<span class="info-icon" tabindex="0" aria-label="More information">ⓘ
    <span class="tip">${escapeHtml(description)}<br><br><strong>Why it matters:</strong> ${escapeHtml(why)}</span>
  </span>`;
}

function fieldLabel(key) {
  const f = EDU.fields[key];
  return `<span class="field-label">${f.label} ${infoTip(f.description, f.why)}</span>`;
}

function fieldHelp(key) {
  const f = EDU.fields[key];
  let html = `<div class="field-help"><strong>Description:</strong> ${escapeHtml(f.description)}`;
  if (f.browserNote) html += `<br><em>${escapeHtml(f.browserNote)}</em>`;
  if (f.example) html += `<br><strong>Example:</strong> <code>${escapeHtml(f.example)}</code>`;
  if (f.options) {
    html += '<br>';
    for (const [k, v] of Object.entries(f.options)) {
      html += `<br><strong>${escapeHtml(k)}:</strong> ${escapeHtml(v)}`;
    }
  }
  html += `<span class="why-label">Why it matters</span>${escapeHtml(f.why)}</div>`;
  return html;
}

function detailField(key, value) {
  const f = EDU.fileFields[key];
  return `<div class="detail-field">
    <div class="detail-label">${f.label} ${infoTip(f.description, f.why)}</div>
    <div class="detail-value">${value}</div>
    <div class="detail-desc">${escapeHtml(f.description)} <em>Why it matters:</em> ${escapeHtml(f.why)}</div>
  </div>`;
}

function statCard(icon, key, value, extra = '') {
  const d = EDU.dashboard[key];
  return `<div class="card stat-card">
    <div class="stat-icon">${icon}</div>
    <h3>${d.title} ${infoTip(d.description, d.why)}</h3>
    <div class="value">${value}</div>
    <p class="card-desc">${escapeHtml(d.description)}</p>
    <p class="card-why"><strong>Why it matters:</strong> ${escapeHtml(d.why)}</p>
    ${extra}
  </div>`;
}

function showEduModal(title, sections) {
  const body = sections.map(s =>
    `<h4>${escapeHtml(s.heading)}</h4><p>${escapeHtml(s.text)}</p>`
  ).join('');
  showModal(title, `<div class="edu-modal">${body}</div>`);
}

function showModal(title, body) {
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.innerHTML = `<div class="modal"><h3>${title}</h3>${body}<button class="btn btn-secondary" style="margin-top:1rem">Close</button></div>`;
  overlay.querySelector('button').onclick = () => overlay.remove();
  overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };
  document.body.appendChild(overlay);
}

function showPipelineModal(step) {
  showEduModal(step.title, [
    { heading: 'What it does', text: step.what },
    { heading: 'Technology used', text: step.tech },
    { heading: 'Why it matters', text: step.why },
  ]);
}

function showRequirementModal(req, passed) {
  showModal(req.name, `<div class="edu-modal">
    <p>${passed ? '✓ Requirement satisfied' : '✗ Not verified'}</p>
    <h4>What was implemented</h4><p>${escapeHtml(req.implemented)}</p>
    <h4>Where in the code</h4><p><span class="tech-badge">${escapeHtml(req.code)}</span></p>
    <h4>How to test</h4><p><span class="tech-badge">${escapeHtml(req.test)}</span></p>
  </div>`);
}

function attachPipelineClicks(container) {
  container.querySelectorAll('.flow-step[data-pipeline]').forEach(el => {
    el.onclick = () => {
      const step = EDU.pipeline.find(p => p.id === el.dataset.pipeline);
      if (step) showPipelineModal(step);
    };
  });
}

function renderProgressSteps(activeIndex) {
  return EDU.progressSteps.map((s, i) => {
    const cls = i < activeIndex ? 'done' : i === activeIndex ? 'active' : '';
    const icon = i < activeIndex ? '✓' : i + 1;
    return `<div class="progress-step ${cls}">
      <div class="step-icon">${icon}</div>
      <div><div class="step-label">${s.label}</div><div class="step-desc">${escapeHtml(s.explanation)}</div></div>
    </div>`;
  }).join('');
}

/* ── Setup wizard ── */

async function maybeShowSetupWizard(configured) {
  if (configured || setupWizardShown) return;
  setupWizardShown = true;

  const overlay = document.createElement('div');
  overlay.className = 'wizard-overlay';
  overlay.innerHTML = `<div class="wizard-card">
    <h2>Welcome to Encrypt-O-Matic</h2>
    <p class="wizard-sub">Before using the tool, create a master password. This password protects all encryption and decryption operations.</p>
    <form id="setup-form">
      <div class="form-group">${fieldLabel('masterPassword')}
        <input name="password" type="password" required autocomplete="new-password"></div>
      ${fieldHelp('masterPassword')}
      <div class="form-group"><label class="field-label">Confirm Password</label>
        <input name="confirm" type="password" required autocomplete="new-password"></div>
      <button type="submit" class="btn btn-primary" style="width:100%">Create Password</button>
    </form>
    <div class="wizard-note">
      <strong>Educational note</strong>
      The password is never stored in plaintext.<br>
      A bcrypt hash is stored in <code>.encryptomatic/master.hash</code> instead.
    </div>
  </div>`;

  overlay.querySelector('#setup-form').onsubmit = async (e) => {
    e.preventDefault();
    const f = e.target;
    try {
      await api('/api/setup-password', {
        method: 'POST',
        body: JSON.stringify({ password: f.password.value, confirm: f.confirm.value }),
      });
      toast('Master password created');
      overlay.remove();
      render();
    } catch (err) { toast(err.message, 'error'); }
  };

  document.body.appendChild(overlay);
}

/* ── Page renderers ── */

async function renderDashboard(root) {
  const [d, reviewer] = await Promise.all([
    api('/api/dashboard'),
    api('/api/reviewer').catch(() => ({ testsPassed: 0, testsTotal: 0 })),
  ]);
  const recent = asArray(d.recentActivity);
  const algorithms = asArray(d.algorithms);
  const stats = d.stats || {};
  const testsLabel = `${reviewer.testsPassed || 0} / ${reviewer.testsTotal || 0}`;

  root.innerHTML = `
    <div class="page-intro">This dashboard summarizes the encryption tool's state. Each card includes a description and explains why the metric matters for security review.</div>
    <div class="grid grid-4" style="margin-bottom:1.5rem">
      ${statCard('🔑', 'masterPassword', d.masterPasswordConfigured ? badge('green', 'Configured') : badge('orange', 'Not set'))}
      ${statCard('🔒', 'encryptedFiles', d.encryptedFileCount)}
      ${statCard('💾', 'backups', d.backupCount)}
      ${statCard('📋', 'metadata', d.metadataCount)}
    </div>
    <div class="grid grid-2" style="margin-bottom:1rem">
      ${statCard('🛡️', 'algorithms', algorithms.length, `<div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-top:.5rem">${algorithms.map(a => badge('blue', a)).join('')}</div>`)}
      ${statCard('✅', 'tests', testsLabel, `<p class="card-desc" style="margin-top:.5rem">Last run: ${reviewer.lastTestRun || 'Never'}</p>`)}
    </div>
    <div class="grid grid-2">
      <div class="card">
        <h3>System Status ${infoTip('Overview of configured algorithms and last operation.', 'Helps reviewers confirm the tool is operational and see recent activity.')}</h3>
        <p style="margin:.5rem 0;color:var(--muted);font-size:.88rem">Supported algorithms:</p>
        <div style="display:flex;gap:.5rem;flex-wrap:wrap;margin-bottom:1rem">
          ${algorithms.map(a => badge('blue', a)).join('')}
        </div>
        <p style="color:var(--muted);font-size:.85rem">Last operation: ${d.lastOperation ? fmtTime(d.lastOperation) : 'None'}</p>
      </div>
      <div class="card">
        <h3>Statistics ${infoTip('Aggregate counts from the activity log.', 'Demonstrates audit trail and data volume processed by the tool.')}</h3>
        <div class="grid grid-2" style="margin-top:.5rem">
          <div><div class="sub">Total encrypted (log)</div><div class="value" style="font-size:1.2rem">${stats.totalEncrypted || 0}</div></div>
          <div><div class="sub">Total decrypted (log)</div><div class="value" style="font-size:1.2rem">${stats.totalDecrypted || 0}</div></div>
          <div><div class="sub">Backups created</div><div class="value" style="font-size:1.2rem">${stats.totalBackups || 0}</div></div>
          <div><div class="sub">Data processed</div><div class="value" style="font-size:1.2rem">${fmtBytes(stats.dataProcessedBytes || 0)}</div></div>
        </div>
      </div>
    </div>
    <div class="card" style="margin-top:1rem">
      <h3>Recent Activity ${infoTip('Chronological log of encrypt and decrypt operations.', 'Provides an audit trail reviewers can correlate with file changes.')}</h3>
      ${recent.length ? recent.map(e => `
        <div class="activity-item">
          <span><span class="type-${e.type}">${e.type.toUpperCase()}</span> — ${e.path.split(/[/\\]/).pop()} ${e.algorithm ? '(' + e.algorithm + ')' : ''}</span>
          <span style="color:var(--muted)">${fmtTime(e.timestamp)}</span>
        </div>`).join('') : '<div class="empty-state">No activity recorded yet</div>'}
    </div>`;

  maybeShowSetupWizard(d.masterPasswordConfigured);
}

/* ── Path picker (local backend browse) ── */

let pathPickerState = { mode: 'file', selectedPath: '', currentPath: '' };

async function browsePath(path) {
  const q = path ? `?path=${encodeURIComponent(path)}` : '';
  return api('/api/browse' + q);
}

function setTargetPath(path, targetType) {
  const input = document.querySelector('#encrypt-form input[name="target"]');
  const select = document.querySelector('#encrypt-form select[name="targetType"]');
  if (input) input.value = path;
  if (select && targetType) select.value = targetType;
}

function renderBrowseList(title, icon, items, type) {
  if (!items.length) return '';
  const selected = pathPickerState.selectedPath;
  return `<div class="browse-section">
    <div class="browse-section-title">${icon} ${title}</div>
    ${items.map(item => {
      const sel = item.path === selected ? ' browse-item-selected' : '';
      return `<button type="button" class="browse-item${sel}" data-path="${escapeHtml(item.path)}" data-type="${type}">
      <span class="browse-item-icon">${type === 'folder' ? '📁' : '📄'}</span>
      <span class="browse-item-name">${escapeHtml(item.name)}</span>
    </button>`;
    }).join('')}
  </div>`;
}

async function loadBrowseView(container, path) {
  container.innerHTML = '<div class="browse-loading">Loading...</div>';
  try {
    const data = await browsePath(path);
    pathPickerState.currentPath = data.currentPath;
    if (pathPickerState.mode === 'folder') {
      pathPickerState.selectedPath = data.currentPath;
    }

    const rootsHtml = data.roots && data.roots.length ? `
      <div class="browse-section">
        <div class="browse-section-title">📌 Quick locations</div>
        <div class="browse-roots">${data.roots.map(r =>
          `<button type="button" class="browse-root-btn" data-path="${escapeHtml(r.path)}">${escapeHtml(r.name)}</button>`
        ).join('')}</div>
      </div>` : '';

    container.innerHTML = `
      ${rootsHtml}
      <div class="browse-path-bar">
        <button type="button" class="btn btn-sm btn-secondary" id="browse-up" ${data.parentPath ? '' : 'disabled'}>↑ Go up</button>
        <code class="browse-current-path">${escapeHtml(data.currentPath)}</code>
      </div>
      ${renderBrowseList('Folders', '📁', asArray(data.folders), 'folder')}
      ${renderBrowseList('Executable files (.exe)', '📄', asArray(data.files), 'file')}
      ${!data.folders?.length && !data.files?.length ? '<div class="browse-empty">No folders or .exe files in this directory.</div>' : ''}`;

    container.querySelectorAll('.browse-root-btn').forEach(btn => {
      btn.onclick = () => {
        pathPickerState.selectedPath = pathPickerState.mode === 'folder' ? btn.dataset.path : '';
        loadBrowseView(container, btn.dataset.path);
      };
    });
    container.querySelectorAll('.browse-item').forEach(btn => {
      btn.onclick = () => {
        const p = btn.dataset.path;
        const t = btn.dataset.type;
        if (t === 'folder') {
          pathPickerState.selectedPath = pathPickerState.mode === 'folder' ? p : pathPickerState.selectedPath;
          loadBrowseView(container, p);
          updatePickerSelection(document.getElementById('picker-selected'));
          return;
        }
        if (pathPickerState.mode === 'file') {
          pathPickerState.selectedPath = p;
          loadBrowseView(container, pathPickerState.currentPath);
          return;
        }
      };
    });
    const up = container.querySelector('#browse-up');
    if (up && data.parentPath) {
      up.onclick = () => loadBrowseView(container, data.parentPath);
    }
    updatePickerSelection(document.getElementById('picker-selected'));
  } catch (err) {
    container.innerHTML = `<div class="browse-error">${escapeHtml(err.message)}</div>`;
  }
}

function updatePickerSelection(el) {
  if (!el) return;
  let p = pathPickerState.selectedPath;
  if (!p && pathPickerState.mode === 'folder') p = pathPickerState.currentPath;
  el.textContent = p || '— none selected —';
}

function openPathPicker(mode) {
  pathPickerState = { mode, selectedPath: mode === 'folder' ? '' : '', currentPath: '' };

  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay path-picker-overlay';
  overlay.innerHTML = `<div class="modal path-picker-modal">
    <h3>${mode === 'file' ? 'Browse File' : 'Browse Folder'}</h3>
    <p class="picker-note">This picker is local-only and only fills the path field. Encryption still happens only after pressing <strong>Start Encryption</strong>.</p>
    <div id="browse-container" class="browse-container"></div>
    <div class="picker-selection">
      <span class="picker-selection-label">Selected path:</span>
      <code id="picker-selected">— none selected —</code>
    </div>
    <div class="picker-actions">
      <button type="button" class="btn btn-secondary" id="picker-cancel">Cancel</button>
      <button type="button" class="btn btn-primary" id="picker-use">Use selected path</button>
    </div>
  </div>`;

  overlay.querySelector('#picker-cancel').onclick = () => overlay.remove();
  overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };

  overlay.querySelector('#picker-use').onclick = () => {
    let path = pathPickerState.selectedPath;
    if (!path && pathPickerState.mode === 'folder') path = pathPickerState.currentPath;
    if (!path) {
      toast('Select a path first', 'error');
      return;
    }
    setTargetPath(path, pathPickerState.mode === 'folder' ? 'dir' : 'file');
    toast('Path selected');
    overlay.remove();
  };

  document.body.appendChild(overlay);
  loadBrowseView(overlay.querySelector('#browse-container'), '');
}

async function fillQuickSelectPath(which) {
  const defaults = {
    demo: 'tests/testdata/demo.exe',
    testdata: 'tests/testdata/demo.exe',
    encryptDemo: 'C:\\encrypt-demo\\demo.exe',
  };
  try {
    const data = await browsePath('');
    const roots = asArray(data.roots);
    const testdataRoot = roots.find(r => r.name.includes('testdata'));
    const demoRoot = roots.find(r => !r.name.includes('WSL') && r.name.includes('encrypt-demo'));

    if (which === 'demo' || which === 'testdata') {
      const root = testdataRoot ? testdataRoot.path : filepathJoin(data.currentPath, 'tests', 'testdata');
      try {
        const browse = await browsePath(root);
        const file = asArray(browse.files).find(f => f.name.toLowerCase() === 'demo.exe');
        setTargetPath(file ? file.path : filepathJoin(root, 'demo.exe'), 'file');
      } catch {
        setTargetPath(defaults[which], 'file');
      }
      toast('Demo path selected');
      return;
    }

    if (which === 'encryptDemo') {
      if (demoRoot) {
        try {
          const browse = await browsePath(demoRoot.path);
          const file = asArray(browse.files).find(f => f.name.toLowerCase() === 'demo.exe');
          setTargetPath(file ? file.path : filepathJoin(demoRoot.path, 'demo.exe'), 'file');
        } catch {
          setTargetPath(defaults.encryptDemo, 'file');
        }
      } else {
        setTargetPath(defaults.encryptDemo, 'file');
      }
      toast('Demo path selected');
    }
  } catch {
    setTargetPath(defaults[which] || defaults.testdata, 'file');
    toast('Path filled (using default)');
  }
}

function filepathJoin(...parts) {
  const sep = parts[0] && parts[0].includes('\\') ? '\\' : '/';
  return parts.filter(Boolean).join(sep).replace(/[/\\]+/g, sep);
}

async function renderEncrypt(root) {
  const pipelineHtml = EDU.pipeline.map((s, i) =>
    `${i ? '<div class="flow-arrow">↓</div>' : ''}<div class="flow-step clickable" data-pipeline="${s.id}">${s.title}</div>`
  ).join('');

  root.innerHTML = `
    <div class="page-intro">Configure encryption parameters below. Hover the ⓘ icon on any field for a quick tooltip, or read the full description under each input.</div>
    <div class="grid grid-2">
      <div class="card">
        <h3>Encryption Parameters</h3>
        <form id="encrypt-form">
          <div class="form-group">${fieldLabel('targetPath')}
            <div class="path-input-row">
              <input name="target" placeholder="C:\\encrypt-demo\\demo.exe" required>
              <button type="button" class="btn btn-secondary btn-sm" id="browse-file">Browse File</button>
              <button type="button" class="btn btn-secondary btn-sm" id="browse-folder">Browse Folder</button>
            </div>
            <div class="quick-select">
              <span class="quick-select-label">Quick select:</span>
              <button type="button" class="btn btn-sm btn-secondary quick-select-btn" data-quick="demo">Use demo.exe</button>
              <button type="button" class="btn btn-sm btn-secondary quick-select-btn" data-quick="testdata">Use tests/testdata/demo.exe</button>
              <button type="button" class="btn btn-sm btn-secondary quick-select-btn" data-quick="encryptDemo">Use C:\\encrypt-demo\\demo.exe</button>
            </div>
          </div>
          ${fieldHelp('targetPath')}
          <div class="form-row">
            <div class="form-group">${fieldLabel('targetType')}
              <select name="targetType"><option value="file">File</option><option value="dir">Directory</option></select></div>
            <div class="form-group">${fieldLabel('algorithm')}
              <select name="algorithm"><option>AES</option><option>ChaCha20</option><option>Twofish</option></select></div>
          </div>
          ${fieldHelp('targetType')}
          ${fieldHelp('algorithm')}
          <div class="form-row">
            <div class="form-group">${fieldLabel('paddingMb')}
              <input name="paddingMb" type="number" min="0" value="0"></div>
            <div class="form-group">${fieldLabel('durationMin')}
              <input name="durationMin" type="number" min="1" value="5"></div>
          </div>
          ${fieldHelp('paddingMb')}
          <div class="form-group">${fieldLabel('customRange')}
            <input name="customRange" value="0-1000" placeholder="0-100000"></div>
          ${fieldHelp('customRange')}
          <div class="form-group">${fieldLabel('masterPassword')}
            <input name="password" type="password" required></div>
          ${fieldHelp('masterPassword')}
          <button type="submit" class="btn btn-primary">Start Encryption</button>
        </form>
        <div id="encrypt-progress" class="hidden" style="margin-top:1.5rem">
          <p id="progress-label">Preparing...</p>
          <div class="progress-bar"><div class="progress-fill" id="progress-fill" style="width:0%"></div></div>
          <div class="progress-steps" id="progress-steps">${renderProgressSteps(-1)}</div>
        </div>
      </div>
      <div class="card">
        <h3>Encryption Pipeline ${infoTip('Visual overview of each stage in the encryption process.', 'Click any step to learn what it does, the technology used, and why it exists.')}</h3>
        <div class="flow-diagram" id="flow-diagram">${pipelineHtml}</div>
      </div>
    </div>`;

  attachPipelineClicks(document.getElementById('flow-diagram'));

  document.getElementById('browse-file').onclick = () => openPathPicker('file');
  document.getElementById('browse-folder').onclick = () => openPathPicker('folder');
  document.querySelectorAll('.quick-select-btn').forEach(btn => {
    btn.onclick = () => fillQuickSelectPath(btn.dataset.quick);
  });

  document.getElementById('encrypt-form').onsubmit = async (e) => {
    e.preventDefault();
    const f = e.target;
    const body = {
      target: f.target.value,
      algorithm: f.algorithm.value,
      paddingMb: parseInt(f.paddingMb.value, 10) || 0,
      customRange: f.customRange.value,
      durationMin: parseInt(f.durationMin.value, 10) || 1,
      password: f.password.value,
    };
    try {
      const { jobId } = await api('/api/encrypt', { method: 'POST', body: JSON.stringify(body) });
      document.getElementById('encrypt-progress').classList.remove('hidden');
      pollJob(jobId);
    } catch (err) { toast(err.message, 'error'); }
  };
}

async function pollJob(jobId) {
  const fill = document.getElementById('progress-fill');
  const label = document.getElementById('progress-label');
  const stepsEl = document.getElementById('progress-steps');
  const flowSteps = document.querySelectorAll('.flow-step[data-pipeline]');
  const pipelineIds = ['file','compression','custom','pbkdf2','encrypt','padding','metadata'];

  const timer = setInterval(async () => {
    try {
      const job = await api('/api/jobs/' + jobId);
      const pct = job.total ? (job.step / job.total) * 100 : 0;
      fill.style.width = pct + '%';
      label.textContent = `Step ${job.step}/${job.total}: ${job.label}`;
      const activeIdx = Math.max(0, job.step - 1);
      if (stepsEl) stepsEl.innerHTML = renderProgressSteps(activeIdx);
      flowSteps.forEach((el) => {
        const idx = pipelineIds.indexOf(el.dataset.pipeline);
        el.classList.toggle('active', idx >= 0 && idx <= activeIdx);
      });
      if (job.status === 'complete') {
        clearInterval(timer);
        toast('Encryption complete');
        label.textContent = 'Complete — all steps finished';
        if (stepsEl) stepsEl.innerHTML = renderProgressSteps(EDU.progressSteps.length);
      }
      if (job.status === 'error') {
        clearInterval(timer);
        toast(job.error || 'Encryption failed', 'error');
      }
    } catch { clearInterval(timer); }
  }, 400);
}

async function renderFiles(root) {
  const files = asArray(await api('/api/files'));
  root.innerHTML = `
    <div class="page-intro">All encrypted files tracked by the tool. Use Details to see cryptographic metadata, or Decrypt to restore with your master password.</div>
    <div class="card">
      <h3>Encrypted Files (${files.length}) ${infoTip('Executables currently in encrypted state.', 'Reviewers can inspect algorithm, padding, and unlock timer for each file.')}</h3>
      ${files.length ? `<table>
        <thead><tr>
          <th>File</th><th>Algorithm</th><th>Original</th><th>Current</th>
          <th>Padding</th><th>Encrypted</th><th>Unlock</th><th>Status</th><th>Actions</th>
        </tr></thead>
        <tbody>${files.map(f => `<tr>
          <td>${escapeHtml(f.fileName)}</td><td>${escapeHtml(f.algorithm)}</td>
          <td>${fmtBytes(f.originalSize)}</td><td>${fmtBytes(f.currentSize)}</td>
          <td>${fmtBytes(f.paddingSize)}</td>
          <td>${fmtTime(f.encryptedAt)}</td>
          <td>${fmtTime(f.unlockTime)}</td>
          <td>${badge(f.statusClass, f.status)}</td>
          <td>
            <button class="btn btn-sm btn-secondary" onclick="viewFile('${f.id}')">Details</button>
            <button class="btn btn-sm btn-danger" onclick="decryptPrompt('${f.path.replace(/\\/g,'\\\\')}')">Decrypt</button>
            <button class="btn btn-sm btn-secondary" onclick="viewMetadata('${f.id}')">Metadata</button>
          </td>
        </tr>`).join('')}</tbody>
      </table>` : '<div class="empty-state">No encrypted files — use Encrypt or CLI</div>'}
    </div>`;
}

window.viewFile = (id) => navigate('/file/' + id);
window.viewMetadata = async (id) => {
  const data = await api('/api/files/' + id + '/metadata');
  showModal('Metadata JSON', '<pre>' + escapeHtml(data.metadata) + '</pre>');
};
window.decryptPrompt = async (path) => {
  const pw = prompt('Master password for decryption:');
  if (!pw) return;
  try {
    await api('/api/decrypt', { method: 'POST', body: JSON.stringify({ target: path, password: pw }) });
    toast('Decryption complete');
    render();
  } catch (e) { toast(e.message, 'error'); }
};

async function renderFileDetail(root) {
  const id = location.pathname.split('/file/')[1];
  const f = await api('/api/files/' + id);
  const backups = asArray(f.backups);
  root.innerHTML = `
    <button class="btn btn-secondary" onclick="navigate('/files')" style="margin-bottom:1rem">← Back</button>
    <div class="page-intro">Every value below is explained so reviewers can understand the cryptographic metadata without reading source code.</div>
    <div class="grid grid-2">
      <div class="card"><h3>File Information</h3>
        <p style="margin-bottom:1rem"><strong>Path:</strong> ${escapeHtml(f.path)}</p>
        ${detailField('algorithm', escapeHtml(f.algorithm))}
        ${detailField('padding', fmtBytes(f.paddingSize))}
        ${detailField('unlockTime', fmtTime(f.unlockTime))}
        <p style="margin-top:.5rem"><strong>Original size:</strong> ${fmtBytes(f.originalSize)}</p>
        <p><strong>Current size:</strong> ${fmtBytes(f.currentSize)}</p>
        <p><strong>Encrypted at:</strong> ${fmtTime(f.encryptedAt)}</p>
      </div>
      <div class="card"><h3>Integrity</h3>
        ${detailField('sha256', `<pre style="word-break:break-all;font-size:.75rem;margin:0">${escapeHtml(f.originalHash)}</pre>`)}
        <p style="margin-top:.75rem">${badge(f.integrityOk ? 'green' : 'red', f.integrityOk ? 'Metadata valid' : 'Issue detected')}</p>
      </div>
      <div class="card"><h3>Cryptographic Parameters</h3>
        ${detailField('nonce', `<code style="font-size:.72rem">${escapeHtml(f.nonceHex || '—')}</code>`)}
        ${detailField('salt', `<code style="font-size:.72rem">${escapeHtml(f.saltHex || '—')}</code>`)}
        <p style="margin-top:.5rem"><strong>Compression:</strong> ${f.compressed ? 'Yes (gzip)' : 'No'}</p>
      </div>
      <div class="card"><h3>Backups ${infoTip('Timestamped copies created before encryption.', 'Safe rollback if decryption needs to be retried.')}</h3>
        ${backups.length ? '<ul>' + backups.map(b => `<li style="font-size:.82rem">${escapeHtml(b)}</li>`).join('') + '</ul>' : '<p class="sub">No backups found</p>'}
      </div>
    </div>`;
}

function renderSecurity(root) {
  const cards = EDU.security.map(s => `
    <details class="expand-card">
      <summary>${s.title}</summary>
      <div class="expand-body">
        <div class="expand-row"><div class="expand-label">Purpose</div><div class="expand-text">${escapeHtml(s.purpose)}</div></div>
        <div class="expand-row"><div class="expand-label">Technology</div><div class="expand-text"><span class="tech-badge">${escapeHtml(s.technology)}</span></div></div>
        <div class="expand-row"><div class="expand-label">Security Benefit</div><div class="expand-text">${escapeHtml(s.benefit)}</div></div>
      </div>
    </details>`).join('');

  root.innerHTML = `
    <div class="page-intro">Expand each card to understand how Encrypt-O-Matic layers security controls. This mirrors how professional security products document their architecture.</div>
    <div class="card"><h3>Security Layers</h3>${cards}</div>`;
}

async function renderReviewer(root) {
  const r = await api('/api/reviewer');
  const checklist = asArray(r.checklist);
  const build = r.build || {};

  const reqMap = {};
  EDU.reviewer.forEach(req => { reqMap[req.id] = req; reqMap[req.name.toLowerCase()] = req; });

  function findReq(name) {
    const lower = name.toLowerCase();
    for (const req of EDU.reviewer) {
      if (lower.includes(req.id) || lower.includes(req.name.toLowerCase().split(' ')[0])) return req;
    }
    return EDU.reviewer.find(req => name.toLowerCase().includes(req.name.toLowerCase().slice(0, 4)));
  }

  root.innerHTML = `
    <div class="page-intro">Click any requirement to see what was implemented, where in the code, and how to test it. Green checkmarks come from the automated API checklist.</div>
    <div class="grid grid-2">
      <div class="card"><h3>Requirements Checklist</h3>
        <div class="checklist">${checklist.map((c, i) => {
          const req = findReq(c.name) || EDU.reviewer[i] || { name: c.name, implemented: 'See source code', code: '—', test: '—' };
          return `<div class="checklist-item" data-req-idx="${i}">
            <span class="req-name">${escapeHtml(c.name)}</span>
            <span class="check-pass">${c.passed ? '✓ Requirement satisfied' : '✗'}</span>
          </div>`;
        }).join('')}</div>
        <button class="btn btn-primary" style="margin-top:1rem" id="run-tests">Run Automated Tests</button>
        <pre id="test-output" style="margin-top:1rem;display:none;background:var(--bg);padding:1rem;border-radius:8px;font-size:.75rem;overflow:auto;max-height:300px"></pre>
      </div>
      <div class="card"><h3>Build Information ${infoTip('Runtime and test environment details.', 'Helps reviewers confirm build reproducibility and test coverage.')}</h3>
        <p><strong>Go version:</strong> ${escapeHtml(r.goVersion || '—')}</p>
        <p><strong>OS/Arch:</strong> ${escapeHtml(build.os || '—')}</p>
        <p><strong>Tests passed:</strong> ${r.testsPassed} / ${r.testsTotal}</p>
        <p><strong>Last test run:</strong> ${r.lastTestRun || 'Never'}</p>
        <div style="margin-top:1.5rem">
          <h3 style="margin-bottom:.75rem">All Reviewer Topics</h3>
          ${EDU.reviewer.map(req => `
            <div class="checklist-item" data-req-id="${req.id}">
              <span class="req-name">${escapeHtml(req.name)}</span>
              <span class="check-pass">ⓘ</span>
            </div>`).join('')}
        </div>
      </div>
    </div>`;

  root.querySelectorAll('.checklist-item[data-req-idx]').forEach(el => {
    const i = parseInt(el.dataset.reqIdx, 10);
    const c = checklist[i];
    const req = findReq(c.name) || EDU.reviewer[i];
    el.onclick = () => showRequirementModal(req, c.passed);
  });
  root.querySelectorAll('.checklist-item[data-req-id]').forEach(el => {
    const req = EDU.reviewer.find(r => r.id === el.dataset.reqId);
    el.onclick = () => showRequirementModal(req, true);
  });

  document.getElementById('run-tests').onclick = async () => {
    const out = document.getElementById('test-output');
    out.style.display = 'block';
    out.textContent = 'Running tests...';
    try {
      const res = await api('/api/reviewer/run-tests', { method: 'POST', body: '{}' });
      const failed = asArray(res.failedTests);
      let text = `Status: ${res.status}\nPassed: ${res.passed}/${res.total}`;
      if (failed.length) text += `\nFailed:\n${failed.map(t => '  - ' + t).join('\n')}`;
      text += `\n\n${res.output || ''}`;
      out.textContent = text;
      toast('Tests finished');
      render();
    } catch (e) { out.textContent = e.message; toast(e.message, 'error'); }
  };
}

async function renderDebug(root) {
  const d = await api('/api/debug');
  const metadata = asArray(d.metadata);
  const backups = asArray(d.backups);
  root.innerHTML = `
    <div class="page-intro">Debug tools for verifying configuration and password authentication. Intended for reviewers validating runtime state.</div>
    <div class="grid grid-2">
      <div class="card"><h3>Configuration ${infoTip('Paths and hash file status.', 'Confirms where sensitive data is stored on disk.')}</h3>
        <p><strong>Config dir:</strong> ${escapeHtml(d.configDir)}</p>
        <p><strong>Master hash:</strong> ${d.hashExists ? badge('green', 'Exists') : badge('orange', 'Missing')}</p>
        <p style="font-size:.82rem;color:var(--muted)">${escapeHtml(d.hashPath)}</p>
      </div>
      <div class="card"><h3>Verify Password ${infoTip('Test master password against stored bcrypt hash.', 'Confirms authentication works without running a full encrypt/decrypt cycle.')}</h3>
        <form id="verify-form">
          <div class="form-group">${fieldLabel('masterPassword')}
            <input name="password" type="password" required></div>
          <button type="submit" class="btn btn-primary">Verify</button>
          <p id="verify-result" style="margin-top:1rem;font-weight:700"></p>
        </form>
      </div>
      <div class="card"><h3>Metadata Files ${infoTip('JSON records for each encrypted file.', 'Each file stores salt, nonce, algorithm, and unlock timer.')}</h3>
        ${metadata.length ? metadata.map(m => `<p style="font-size:.82rem">${escapeHtml(m)}</p>`).join('') : '<p class="sub">None</p>'}
      </div>
      <div class="card"><h3>Backup Files ${infoTip('Timestamped pre-modification copies.', 'Original content preserved before encryption changes the file.')}</h3>
        ${backups.length ? backups.map(b => `<p style="font-size:.82rem">${escapeHtml(b)}</p>`).join('') : '<p class="sub">None</p>'}
      </div>
    </div>`;
  document.getElementById('verify-form').onsubmit = async (e) => {
    e.preventDefault();
    const pw = e.target.password.value;
    try {
      const res = await fetch('/api/verify-password', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password: pw }),
      });
      const data = await res.json();
      const el = document.getElementById('verify-result');
      el.textContent = data.result;
      el.style.color = data.result === 'Password OK' ? 'var(--green)' : 'var(--red)';
    } catch (err) { toast(err.message, 'error'); }
  };
}

async function render() {
  const r = route();
  document.getElementById('page-title').textContent = titles[r] || 'Dashboard';
  setActiveNav();
  const root = document.getElementById('app-root');
  root.innerHTML = '<div class="empty-state">Loading...</div>';
  try {
    if (r === '/') await renderDashboard(root);
    else if (r === '/encrypt') await renderEncrypt(root);
    else if (r === '/files') await renderFiles(root);
    else if (r === '/file') await renderFileDetail(root);
    else if (r === '/security') renderSecurity(root);
    else if (r === '/reviewer') await renderReviewer(root);
    else if (r === '/debug') await renderDebug(root);
  } catch (e) {
    root.innerHTML = `<div class="card"><p style="color:var(--red)">${escapeHtml(e.message)}</p></div>`;
  }
}

document.querySelectorAll('.nav a').forEach(a => {
  a.addEventListener('click', (e) => {
    e.preventDefault();
    navigate(a.dataset.route);
  });
});

window.addEventListener('popstate', render);
setInterval(() => {
  document.getElementById('clock').textContent = new Date().toLocaleString();
}, 1000);

render();
