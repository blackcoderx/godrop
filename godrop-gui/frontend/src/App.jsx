import { useState, useEffect } from 'react';
import './App.css';
import { GetHomeDir, ReadDir, StartServer, StopServer, StartReceiveServer, StartClipboardServer, GetDefaultSaveDir, SelectDirectory, GetSystemClipboard, SetSystemClipboard } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

function App() {
    const [currentPath, setCurrentPath] = useState("");
    const [files, setFiles] = useState([]);
    const [selectedFiles, setSelectedFiles] = useState([]);
    const [serverInfo, setServerInfo] = useState(null); // { ip, port, fullUrl, qrCode }
    const [logs, setLogs] = useState([]);
    const [password, setPassword] = useState("");
    const [port, setPort] = useState("8080");
    const [limit, setLimit] = useState(1);
    const [timeout, setTimeoutVal] = useState(10);
    const [isServerRunning, setIsServerRunning] = useState(false);

    const [mode, setMode] = useState('send'); // 'send' | 'receive' | 'clipboard'
    const [saveLocation, setSaveLocation] = useState("");
    const [clipboardText, setClipboardText] = useState("");

    // Initial Load
    useEffect(() => {
        const init = async () => {
            const home = await GetHomeDir();
            setCurrentPath(home);
            loadDir(home);

            const defaultSave = await GetDefaultSaveDir();
            setSaveLocation(defaultSave);

            const clip = await GetSystemClipboard();
            setClipboardText(clip);
        };
        init();

        // Listen for server events
        EventsOn("download_started", (data) => {
            addLog(`Download started from ${data.ip}`);
        });

        EventsOn("file-received", (filename) => {
            addLog(`RECEIVED FILE: ${filename}`);
        });

        EventsOn("server_error", (err) => {
            addLog(`Error: ${err}`);
        });

        EventsOn("server_stopped", () => {
            addLog("Server stopped.");
            setIsServerRunning(false);
            setServerInfo(null);
        });

    }, []);

    // Polling for Clipboard in Clipboard Mode
    useEffect(() => {
        if (mode !== 'clipboard' || isServerRunning) return;
        const interval = setInterval(async () => {
            const text = await GetSystemClipboard();
            setClipboardText(text);
        }, 2000);
        return () => clearInterval(interval);
    }, [mode, isServerRunning]);

    const loadDir = async (path) => {
        try {
            const entries = await ReadDir(path);
            // Sort: folders first
            const sorted = entries.sort((a, b) => {
                if (a.isDir === b.isDir) return a.name.localeCompare(b.name);
                return a.isDir ? -1 : 1;
            });
            setFiles(sorted);
            setCurrentPath(path);
        } catch (err) {
            console.error(err);
        }
    };

    const handleNavigate = (file) => {
        if (file.isDir) {
            // Basic path joining. Only supports Windows for now based on context, but cleaner is to let backend handle joining if possible or simple string concat
            // Wails sends paths as strings.
            loadDir(file.img); // .img holds full path from backend
        }
    };

    const handleUp = () => {
        // Simple string manipulation to go up one level
        // Supports both forward and backslashes
        const separator = currentPath.includes("/") ? "/" : "\\";
        const parts = currentPath.split(separator);
        parts.pop();
        const newPath = parts.join(separator) || "root"; // Fallback to root if empty
        loadDir(newPath);
    };

    const toggleSelect = (file) => {
        if (file.isDir) return; // Only select files
        const path = file.img;
        if (selectedFiles.includes(path)) {
            setSelectedFiles(selectedFiles.filter(p => p !== path));
        } else {
            setSelectedFiles([...selectedFiles, path]);
        }
    };

    const handleStartServer = async () => {
        if (mode === 'send' && selectedFiles.length === 0) return;

        setLogs([`INITIALIZING ${mode.toUpperCase()} SERVER ON PORT ${port}...`]);
        try {
            let info;
            if (mode === 'send') {
                info = await StartServer(port, password, selectedFiles, limit, timeout);
                addLog(`HOSTING ${selectedFiles.length} FILE(S)`);
            } else if (mode === 'receive') {
                info = await StartReceiveServer(port, saveLocation);
                addLog(`DROPZONE ACTIVE. Saving to: ${saveLocation}`);
            } else {
                info = await StartClipboardServer(port);
                addLog(`CLIPBOARD SERVER ACTIVE.`);
            }

            setServerInfo(info);
            setIsServerRunning(true);
            addLog(`LIVE AT: ${info.fullUrl}`);
        } catch (err) {
            addLog(`FAILED: ${err}`);
        }
    };

    const handleStopServer = async () => {
        await StopServer();
        setIsServerRunning(false);
        setServerInfo(null);
    };

    const addLog = (msg) => {
        setLogs(prev => [...prev, `> ${msg}`]);
    };

    return (
        <div id="app" className="app-container">
            {/* Sidebar */}
            <aside className="sidebar">
                <div className="brand">⚡ GODROP</div>
                <div className="nav-item active" onClick={() => GetHomeDir().then(loadDir)}>💻 My PC</div>
                <div className="nav-item" onClick={() => loadDir("root")}>💿 Drives</div>
            </aside>

            {/* Explorer */}
            <main className="explorer">
                <div className="address-bar">
                    <button onClick={handleUp} style={{ cursor: 'pointer', background: 'transparent', border: 'none', fontSize: '1.2rem' }}>⬆</button>
                    <input type="text" className="path-input" value={currentPath} readOnly />
                </div>
                <div className="file-grid">
                    {files.map((file) => (
                        <div
                            key={file.name}
                            className={`file-item ${selectedFiles.includes(file.img) ? 'selected' : ''}`}
                            onClick={() => toggleSelect(file)}
                            onDoubleClick={() => handleNavigate(file)}
                        >
                            <div className="file-icon">
                                {file.isDir ? '📁' :
                                    file.name.endsWith('.zip') ? '📦' : '📄'}
                            </div>
                            <div className="file-name">{file.name}</div>
                        </div>
                    ))}
                </div>
            </main>

            {/* Config Panel */}
            <aside className="config-panel">
                <div className="panel-title">Mode: {mode.toUpperCase()}</div>

                <div className="mode-toggle">
                    <button className={mode === 'send' ? 'active' : ''} onClick={() => setMode('send')}>SEND</button>
                    <button className={mode === 'receive' ? 'active' : ''} onClick={() => setMode('receive')}>RECEIVE</button>
                    <button className={mode === 'clipboard' ? 'active' : ''} onClick={() => setMode('clipboard')}>CLIPBOARD</button>
                </div>

                {mode === 'send' ? (
                    <>
                        <label className="form-label">Selected Files ({selectedFiles.length})</label>
                        <div className="selection-list">
                            {selectedFiles.length === 0 && <div style={{ textAlign: 'center', color: '#999', marginTop: 10 }}>Select files...</div>}
                            {selectedFiles.map(path => (
                                <div key={path} className="selected-tag">
                                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '170px' }}>
                                        {path.split(/[/\\]/).pop()}
                                    </span>
                                    <span style={{ color: 'red', cursor: 'pointer' }} onClick={() => setSelectedFiles(selectedFiles.filter(p => p !== path))}>×</span>
                                </div>
                            ))}
                        </div>
                    </>
                ) : mode === 'receive' ? (
                    <>
                        <label className="form-label">Save Location</label>
                        <div className="dropzone-info">
                            <div style={{ fontSize: '0.8rem', wordBreak: 'break-all', marginBottom: 10 }}>{saveLocation || "Not set"}</div>
                            <button className="btn-small" onClick={async () => {
                                const dir = await SelectDirectory();
                                if (dir) setSaveLocation(dir);
                            }}>Change...</button>
                        </div>
                    </>
                ) : (
                    <>
                        <label className="form-label">System Clipboard</label>
                        <textarea
                            className="form-input"
                            style={{ height: '150px', resize: 'none', fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}
                            value={clipboardText}
                            onChange={(e) => setClipboardText(e.target.value)}
                        />
                        <button className="btn-small" style={{ marginTop: 10, width: '100%' }} onClick={() => SetSystemClipboard(clipboardText)}>
                            UPDATE PC CLIPBOARD
                        </button>
                    </>
                )}

                <hr style={{ width: '100%', borderColor: 'var(--border)', opacity: 0.3, margin: '20px 0' }} />

                <div className="form-group">
                    <label className="form-label">Password (Optional)</label>
                    <input
                        type="password"
                        className="form-input"
                        placeholder="••••••"
                        value={password}
                        onChange={e => setPassword(e.target.value)}
                    />
                </div>

                {mode === 'send' && (
                    <div className="toggle-row" style={{ display: 'flex', gap: 10 }}>
                        <div className="form-group" style={{ flex: 1 }}>
                            <label className="form-label">Limit (0=Unlim)</label>
                            <input type="number" className="form-input" min="0" value={limit} onChange={e => setLimit(parseInt(e.target.value) || 0)} />
                        </div>
                        <div className="form-group" style={{ flex: 1 }}>
                            <label className="form-label">Time (Min)</label>
                            <input type="number" className="form-input" min="0" value={timeout} onChange={e => setTimeoutVal(parseInt(e.target.value) || 0)} />
                        </div>
                    </div>
                )}

                <div className="form-group">
                    <label className="form-label">Port</label>
                    <input
                        type="number"
                        className="form-input"
                        value={port}
                        onChange={e => setPort(e.target.value)}
                    />
                </div>

                <button
                    className="btn-start"
                    onClick={handleStartServer}
                    disabled={(mode === 'send' && selectedFiles.length === 0) || isServerRunning}
                >
                    {mode === 'send' ? 'START SENDING 🚀' :
                        mode === 'receive' ? 'OPEN DROPZONE 📥' :
                            'START CLIPBOARD 📋'}
                </button>
            </aside>

            {/* Server Overlay */}
            {isServerRunning && serverInfo && (
                <div className="server-overlay">
                    <div className="server-card">
                        <div className="card-header">
                            <span>GODROP SERVER RUNNING</span>
                            <span>● LIVE</span>
                        </div>
                        <div className="card-body">
                            <div className="qr-section">
                                <img src={serverInfo.qrCode} alt="QR Code" style={{ width: 150, height: 150 }} />
                                <div className="url-box">{serverInfo.fullUrl}</div>
                                <div style={{ fontSize: '0.7rem', marginTop: 5, color: '#999' }}>Scan to Download</div>
                            </div>
                            <div className="log-section">
                                {logs.map((log, i) => <div key={i}>{log}</div>)}
                            </div>
                        </div>
                        <button className="btn-stop" onClick={handleStopServer}>STOP SERVER</button>
                    </div>
                </div>
            )}
        </div>
    );
}

export default App;
