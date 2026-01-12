import { useState, useEffect } from 'react';
import './App.css';
import { GetHomeDir, ReadDir, StartServer, StopServer } from '../wailsjs/go/main/App';
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

    // Initial Load
    useEffect(() => {
        const init = async () => {
            const home = await GetHomeDir();
            setCurrentPath(home);
            loadDir(home);
        };
        init();

        // Listen for server events
        EventsOn("download_started", (data) => {
            addLog(`Download started from ${data.ip}`);
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
        if (selectedFiles.length === 0) return;
        setLogs([`INITIALIZING SERVER ON PORT ${port}...`]);
        try {
            const info = await StartServer(port, password, selectedFiles, limit, timeout);
            setServerInfo(info);
            setIsServerRunning(true);
            addLog(`HOSTING ${selectedFiles.length} FILE(S)`);
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
                <div className="panel-title">Preparation</div>
                <label className="form-label">Selected Files ({selectedFiles.length})</label>
                <div className="selection-list">
                    {selectedFiles.length === 0 && <div style={{ textAlign: 'center', color: '#999', marginTop: 10 }}>Select files...</div>}
                    {selectedFiles.map(path => (
                        <div key={path} className="selected-tag">
                            <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '200px' }}>
                                {path.split(/[/\\]/).pop()}
                            </span>
                            <span style={{ color: 'red', cursor: 'pointer' }} onClick={() => setSelectedFiles(selectedFiles.filter(p => p !== path))}>×</span>
                        </div>
                    ))}
                </div>

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
                    disabled={selectedFiles.length === 0 || isServerRunning}
                >
                    START SERVER 🚀
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
