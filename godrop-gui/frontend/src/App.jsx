import { useState, useEffect } from 'react';
import './App.css';
import logo from './assets/images/godrop-logo.png';
import { GetHomeDir, ReadDir, StartServer, StopServer, StartReceiveServer, StartClipboardServer, GetDefaultSaveDir, SelectDirectory, GetSystemClipboard, SetSystemClipboard } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const RetroProgressBar = ({ percent }) => {
    return (
        <div style={{
            width: '100%',
            height: '16px',
            background: 'var(--bg-panel)',
            border: '2px solid var(--border)',
            position: 'relative',
            marginTop: '10px',
            overflow: 'hidden',
            borderRadius: '4px'
        }}>
            <div style={{
                width: `${percent}%`,
                height: '100%',
                background: 'var(--accent)',
                transition: 'width 0.2s cubic-bezier(0.4, 0, 0.2, 1)'
            }} />
        </div>
    );
};

function App() {
    // Explorer State
    const [currentPath, setCurrentPath] = useState("");
    const [files, setFiles] = useState([]);
    const [selectedFiles, setSelectedFiles] = useState([]);

    // Server State
    const [mode, setMode] = useState('send'); // 'send' | 'receive' | 'clipboard'
    const [connectivity, setConnectivity] = useState('local'); // 'local' | 'cloud'
    const [password, setPassword] = useState("");
    const [port, setPort] = useState("8080");
    const [limit, setLimit] = useState(1);
    const [timeout, setTimeoutVal] = useState(10);
    const [saveLocation, setSaveLocation] = useState("");

    // UI State
    const [isServerRunning, setIsServerRunning] = useState(false);
    const [serverInfo, setServerInfo] = useState(null);
    const [progress, setProgress] = useState(null);
    const [logs, setLogs] = useState([]);
    const [clipboardText, setClipboardText] = useState("");

    // Initial Load
    useEffect(() => {
        const init = async () => {
            const home = await GetHomeDir();
            loadDir(home);
            const defaultSave = await GetDefaultSaveDir();
            setSaveLocation(defaultSave);
        };
        init();

        EventsOn("download_started", (data) => addLog(`Download started from ${data.ip}`));
        EventsOn("file-received", (filename) => addLog(`RECEIVED: ${filename}`));
        EventsOn("server_error", (err) => addLog(`ERROR: ${err}`));
        EventsOn("server_stopped", () => {
            addLog("Server stopped.");
            setIsServerRunning(false);
            setServerInfo(null);
            setProgress(null);
        });
        EventsOn("transfer-progress", (data) => setProgress(data));
    }, []);

    // Clipboard Update Loop
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
            const sorted = entries.sort((a, b) => {
                if (a.isDir === b.isDir) return a.name.localeCompare(b.name);
                return a.isDir ? -1 : 1;
            });
            setFiles(sorted);
            setCurrentPath(path);
        } catch (err) {
            addLog(`Error loading dir: ${err}`);
        }
    };

    const handleNavigate = (file) => {
        if (file.isDir) loadDir(file.img);
    };

    const handleUp = () => {
        const separator = currentPath.includes("/") ? "/" : "\\";
        const parts = currentPath.split(separator);
        parts.pop();
        const newPath = parts.join(separator) || "root";
        loadDir(newPath);
    };

    const toggleSelect = (file) => {
        if (file.isDir) return;
        const path = file.img;
        if (selectedFiles.includes(path)) {
            setSelectedFiles(selectedFiles.filter(p => p !== path));
        } else {
            setSelectedFiles([...selectedFiles, path]);
        }
    };

    const handleStartServer = async () => {
        if (mode === 'send' && selectedFiles.length === 0) return;
        if (connectivity === 'cloud') {
            alert("Cloud Tunneling is coming soon!");
            return;
        }

        setLogs([`INITIALIZING ${mode.toUpperCase()} OVER ${connectivity.toUpperCase()}...`]);
        try {
            let info;
            if (mode === 'send') {
                info = await StartServer(port, password, selectedFiles, limit, timeout);
                addLog(`BROADCASTING ${selectedFiles.length} FILES`);
            } else if (mode === 'receive') {
                info = await StartReceiveServer(port, saveLocation);
                addLog(`DROPZONE ACTIVE -> ${saveLocation}`);
            } else {
                info = await StartClipboardServer(port);
                addLog(`CLIPBOARD SYNC ACTIVE`);
            }
            setServerInfo(info);
            setPort(info.port);
            setIsServerRunning(true);
        } catch (err) {
            addLog(`STARTUP FAILED: ${err}`);
        }
    };

    const handleAddLog = (msg) => setLogs(prev => [...prev, `> ${msg}`]);
    const addLog = handleAddLog;

    // Helper to get file basename
    const basename = (path) => path.split(/[/\\]/).pop();

    return (
        <div id="app">
            {/* COLUMN 1: SIDEBAR */}
            <aside className="sidebar">
                <div className="brand">
                    <img src={logo} alt="Godrop" className="brand-logo" />
                    Godrop
                </div>

                <div className={`nav-item ${mode === 'send' || mode === 'receive' ? 'active' : ''}`} onClick={() => setMode('send')}>
                    📁 All Files
                </div>
                <div className="nav-item">
                    💿 Drives
                </div>
                <div className={`nav-item ${mode === 'clipboard' ? 'active' : ''}`} onClick={() => setMode('clipboard')}>
                    📋 Clipboard
                </div>

                <div className="user-account">
                    User account
                </div>
            </aside>

            {/* COLUMN 2: MAIN EXPLORER */}
            <main className="explorer">
                <div className="top-bar">
                    <button className="btn-icon" onClick={handleUp}>↑</button>
                    <div className="breadcrumb">
                        My PC / {currentPath.split(/[/\\]/).slice(-2).map((p, i) => (
                            <span key={i}>{p}{i === 0 ? ' / ' : ''}</span>
                        ))}
                    </div>
                </div>

                <div className="file-grid">
                    {files.map((file) => (
                        <div
                            key={file.name}
                            className={`file-card ${selectedFiles.includes(file.img) ? 'selected' : ''}`}
                            onClick={() => toggleSelect(file)}
                            onDoubleClick={() => handleNavigate(file)}
                        >
                            <div className="file-icon">
                                {file.isDir ? '📁' : file.name.match(/\.(jpg|jpeg|png|gif)$/i) ? '🖼️' : file.name.endsWith('.pdf') ? '📄' : '📦'}
                            </div>
                            <div className="file-name">{file.name}</div>
                        </div>
                    ))}
                </div>
            </main>

            {/* COLUMN 3: CONFIG PANEL */}
            <aside className="config-panel">
                <div className="section-label">Connectivity</div>
                <div className="toggle-group">
                    <div className={`toggle-item ${connectivity === 'local' ? 'active' : ''}`} onClick={() => setConnectivity('local')}>Local</div>
                    <div className={`toggle-item ${connectivity === 'cloud' ? 'active' : ''}`} onClick={() => setConnectivity('cloud')}>Cloud</div>
                </div>

                <div className="toggle-group" style={{ background: 'transparent' }}>
                    <div className={`toggle-item ${mode === 'send' || mode === 'clipboard' ? 'active' : ''}`} onClick={() => setMode('send')}>Send</div>
                    <div className={`toggle-item ${mode === 'receive' ? 'active' : ''}`} onClick={() => setMode('receive')}>Receive</div>
                </div>

                {mode === 'send' && (
                    <>
                        <div className="section-label">
                            <span>Selected Files</span>
                            <span>{selectedFiles.length}</span>
                        </div>
                        <div className="list-box">
                            {selectedFiles.map(path => (
                                <div key={path} className="list-item">
                                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{basename(path)}</span>
                                    <span style={{ opacity: 0.5, cursor: 'pointer' }} onClick={() => setSelectedFiles(selectedFiles.filter(p => p !== path))}>×</span>
                                </div>
                            ))}
                        </div>
                    </>
                )}

                {mode === 'receive' && (
                    <>
                        <div className="section-label">Dropzone Location</div>
                        <div className="list-box" style={{ padding: '20px', textAlign: 'center' }}>
                            <div style={{ fontSize: '0.8rem', opacity: 0.6, marginBottom: '10px' }}>{basename(saveLocation)}</div>
                            <button className="btn-icon" style={{ width: '100%' }} onClick={async () => {
                                const dir = await SelectDirectory();
                                if (dir) setSaveLocation(dir);
                            }}>Change Path</button>
                        </div>
                    </>
                )}

                {mode === 'clipboard' && (
                    <>
                        <div className="section-label">Local Clipboard</div>
                        <textarea
                            className="input-ui"
                            style={{ flex: 1, minHeight: '150px', resize: 'none', marginBottom: '10px' }}
                            value={clipboardText}
                            onChange={(e) => setClipboardText(e.target.value)}
                        />
                        <button className="btn-icon" style={{ width: '100%', marginBottom: '25px' }} onClick={() => SetSystemClipboard(clipboardText)}>Add to Clipboard</button>
                    </>
                )}

                <div className="input-block">
                    <label className="input-label">Password</label>
                    <input type="password" placeholder="optional" className="input-ui" value={password} onChange={e => setPassword(e.target.value)} />
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px' }}>
                    <div className="input-block">
                        <label className="input-label">Time (min)</label>
                        <input type="number" className="input-ui" value={timeout} onChange={e => setTimeoutVal(parseInt(e.target.value) || 0)} />
                    </div>
                    <div className="input-block">
                        <label className="input-label">Port</label>
                        <input type="number" className="input-ui" value={port} onChange={e => setPort(e.target.value)} />
                    </div>
                </div>

                <button
                    className="btn-primary"
                    onClick={handleStartServer}
                    disabled={isServerRunning || (mode === 'send' && selectedFiles.length === 0)}
                >
                    {isServerRunning ? 'SERVER LIVE' : 'Start Server'}
                </button>
            </aside>

            {/* SERVER OVERLAY */}
            {isServerRunning && serverInfo && (
                <div className="server-overlay">
                    <div className="server-card">
                        <div className="card-header">
                            <span>{mode.toUpperCase()} ● LIVE</span>
                            <span>{connectivity.toUpperCase()}</span>
                        </div>
                        <div className="card-body">
                            <div className="qr-box">
                                <img src={serverInfo.qrCode} className="qr-image" alt="QR" />
                                <div className="url-box">{serverInfo.fullUrl}</div>
                                <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', marginTop: '10px' }}>Scan to connect</div>
                            </div>
                            <div className="log-box">
                                {logs.map((log, i) => <div key={i}>{log}</div>)}
                            </div>
                        </div>
                        {progress && (
                            <div className="progress-container">
                                <RetroProgressBar percent={progress.percent} />
                                <div className="progress-text">
                                    {Math.round(progress.transferred / 1024 / 1024 * 10) / 10} MB / {Math.round(progress.total / 1024 / 1024 * 10) / 10} MB
                                </div>
                            </div>
                        )}
                        <button className="btn-stop" onClick={async () => {
                            await StopServer();
                            setIsServerRunning(false);
                            setServerInfo(null);
                        }}>STOP SERVER</button>
                    </div>
                </div>
            )}
        </div>
    );
}

export default App;
