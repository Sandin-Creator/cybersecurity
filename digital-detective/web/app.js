let currentTab = 'username';

function switchTab(tab) {
    currentTab = tab;

    // Update buttons
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');

    // Update inputs
    document.querySelectorAll('.form-group').forEach(grp => grp.classList.remove('active'));
    document.getElementById(`${tab}-form`).classList.add('active');
}

function handleEnter(e) {
    if (e.key === 'Enter') {
        performSearch();
    }
}

async function performSearch() {
    const terminal = document.getElementById('terminal-content');
    let query = '';

    if (currentTab === 'username') query = document.getElementById('username-input').value;
    else if (currentTab === 'name') query = document.getElementById('name-input').value;
    else if (currentTab === 'ip') query = document.getElementById('ip-input').value;
    else if (currentTab === 'email') query = document.getElementById('email-input').value;
    else if (currentTab === 'dating') query = document.getElementById('dating-input').value;

    if (!query) return;

    // Show loading
    terminal.innerHTML = `
        <span class="system-msg">> INITIALIZING SCAN PROTOCOLS...</span><br>
        <span class="system-msg">> TARGET: ${query}</span><br>
        <div class="loading-bar"></div>
    `;

    try {
        const response = await fetch('/api/search', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ type: currentTab, query: query })
        });

        const data = await response.json();



        // Separate header/footer from results to avoid messing up banner
        // But for simplicity, we'll split by line and process
        const lines = data.output.split('\n');
        let formattedHTML = '';

        lines.forEach(line => {
            // Basic sanitation
            let safeLine = line.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

            // Highlight keywords
            safeLine = safeLine.replace(/\[\+\] FOUND:/g, '<span class="found">[+] FOUND:</span>');
            safeLine = safeLine.replace(/\[-\] NOT FOUND:/g, '<span class="not-found">[-] NOT FOUND:</span>');
            safeLine = safeLine.replace(/CONFIDENCE:/g, '<span class="highlight">CONFIDENCE:</span>');

            // Linkify URLs
            const urlRegex = /(https?:\/\/[^\s]+)/g;
            safeLine = safeLine.replace(urlRegex, (url) => {
                // Check if we have an image for this URL
                if (data.images && data.images[url]) {
                    const imgSrc = data.images[url];
                    return `<br><a href="${url}" target="_blank" class="profile-link">
                              <img src="${imgSrc}" class="profile-thumb" alt="Profile Image">
                              <span class="link-text">${url}</span>
                           </a>`;
                }
                return `<a href="${url}" target="_blank">${url}</a>`;
            });

            formattedHTML += safeLine + '<br>';
        });

        terminal.innerHTML = formattedHTML + '<br><span class="cursor">_</span>';

    } catch (e) {
        terminal.innerHTML = `<span class="error">CRITICAL ERROR: CONNECTION FAILED</span><br><span class="cursor">_</span>`;
    }
}
