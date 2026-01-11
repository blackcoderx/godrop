/**
 * GODROP FRONTEND LOGIC
 * This script handles the communication with the Go backend API.
 * It periodically refreshes the page stats and handles the countdown timer.
 */

document.addEventListener('DOMContentLoaded', async () => {
    // UI Elements Selection
    const filenameEl = document.getElementById('filename');
    const filesizeEl = document.getElementById('filesize');
    const downloadsRemainingEl = document.getElementById('downloads-remaining');
    const downloadsTotalEl = document.getElementById('downloads-total');
    const timeLeftEl = document.getElementById('time-left');
    const downloadBtn = document.getElementById('download-btn');
    const securityCheck = document.getElementById('security-check');
    const verifyBtn = document.getElementById('verify-btn');
    const securityCodeInput = document.getElementById('security-code');
    const statusBadge = document.getElementById('status-badge');

    // Store state from server
    let serverStats = {};

    /**
     * Converts seconds into a HH:MM:SS format
     */
    function formatTime(seconds) {
        if (seconds <= 0) return "SHUTTING DOWN...";
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        return [h, m, s].map(v => v < 10 ? "0" + v : v).join(":");
    }

    /**
     * Fetch the latest stats from the server API (/api/stats)
     * and update the UI accordingly.
     */
    async function updateStats() {
        try {
            const resp = await fetch('/api/stats');
            if (!resp.ok) throw new Error('System Offline');
            serverStats = await resp.json();

            // Update file details
            filenameEl.textContent = serverStats.FileName;
            filesizeEl.textContent = `${(serverStats.FileSize / 1024 / 1024).toFixed(2)} MB`;

            // Update download counts
            downloadsRemainingEl.textContent = serverStats.Limit - serverStats.Current;
            downloadsTotalEl.textContent = serverStats.Limit;

            // Security Logic: check if the user needs to enter a code
            // If they already unlocked it, we store it in sessionStorage so they don't have to re-enter
            if (serverStats.HasCode && !sessionStorage.getItem('unlocked')) {
                securityCheck.style.display = 'block';
                downloadBtn.disabled = true;
            } else {
                securityCheck.style.display = 'none';
                downloadBtn.disabled = false;
            }

            // Global Limit Check: if all downloads are used up
            if (serverStats.Current >= serverStats.Limit) {
                downloadBtn.textContent = 'LINK_EXPIRED';
                downloadBtn.disabled = true;
                statusBadge.textContent = 'LINK_EXPIRED';
            }

            updateTimer(); // Update timer immediately after getting stats
        } catch (err) {
            console.error('Failed to fetch stats:', err);
            statusBadge.textContent = 'SYSTEM_OFFLINE';
            statusBadge.style.color = 'red';
            timeLeftEl.textContent = '00:00:00';
        }
    }

    /**
     * Handles the visual countdown timer on the page.
     * This runs every second independently of the stats poll.
     */
    function updateTimer() {
        if (!serverStats.ExpiryTime || serverStats.ExpiryTime === 0) {
            timeLeftEl.textContent = 'INFINITY';
            return;
        }

        const now = Math.floor(Date.now() / 1000);
        const diff = serverStats.ExpiryTime - now;
        timeLeftEl.textContent = formatTime(diff);

        // If time is up, lock the download button
        if (diff <= 0) {
            downloadBtn.disabled = true;
            downloadBtn.textContent = 'TIMEOUT';
        }
    }

    /**
     * Submit the security code to /api/verify
     */
    verifyBtn.addEventListener('click', async () => {
        const code = securityCodeInput.value;
        const resp = await fetch('/api/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code })
        });
        const result = await resp.json();

        if (result.success) {
            // Unlock locally if the server says it's correct
            sessionStorage.setItem('unlocked', 'true');
            securityCheck.style.display = 'none';
            downloadBtn.disabled = false;
        } else {
            // Shake effect or error feedback would go here
            securityCodeInput.value = '';
            securityCodeInput.placeholder = 'INVALID_CODE_TRY_AGAIN...';
        }
    });

    /**
     * Trigger the actual file download via the API
     */
    downloadBtn.addEventListener('click', () => {
        window.location.href = '/api/download';
    });

    // --- INITIALIZATION ---
    updateStats(); // First load

    // Refresh data from server every 5 seconds (SaaS Dashboard style)
    setInterval(updateStats, 5000);

    // Update the visual timer every second for smooth ticking
    setInterval(updateTimer, 1000);
});
