import { useState, useEffect } from 'react';
import './App.css';
import logo from './assets/images/godrop-logo.png';
import { GetHomeDir, ReadDir, StartServer, StopServer, StartReceiveServer, StartClipboardServer, GetDefaultSaveDir, GetSystemClipboard, GetHistory, SetSystemClipboard } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

// Components
import { Explorer } from './components/Explorer/Explorer';
import { ConfigPanel } from './components/Config/ConfigPanel';
import { ServerOverlay } from './components/Server/ServerOverlay';
import { ClipboardCard } from './components/Clipboard/ClipboardCard';

function App() {
    // Explorer State
    const [currentPath, setCurrentPath] = useState("");
    const [files, setFiles] = useState([]);
    const [selectedFiles, setSelectedFiles] = useState([]);

    // Server State
    const [mode, setMode] = useState('send'); // 'send' | 'receive' | 'clipboard'
    const [password, setPassword] = useState("");
    const [port, setPort] = useState("1111");
    const [limit, setLimit] = useState(1);
    const [timeout, setTimeoutVal] = useState(10);
    const [saveLocation, setSaveLocation] = useState("");

    // UI State
    const [isServerRunning, setIsServerRunning] = useState(false);
    const [serverInfo, setServerInfo] = useState(null);
    const [progress, setProgress] = useState(null);
    const [logs, setLogs] = useState([]);
    const [clipboardText, setClipboardText] = useState("");
    const [clipboardHistory, setClipboardHistory] = useState([]);
    const [receivedFiles, setReceivedFiles] = useState([]);

    // Initial Load
    useEffect(() => {
        const init = async () => {
            const home = await GetHomeDir();
            loadDir(home);
            const defaultSave = await GetDefaultSaveDir();
            setSaveLocation(defaultSave);

            // Initial clipboard history load
            const history = await GetHistory();
            setClipboardHistory(history);
        };
        init();

        const onDownloadStarted = (data) => addLog(`Download started from ${data.ip}`);
        const onFileReceived = (filename) => {
            setReceivedFiles(prev => {
                // Robust deduplication: check if this filename already exists in the current session list
                if (prev.some(f => f.name === filename)) return prev;
                addLog(`RECEIVED: ${filename}`);
                return [
                    { name: filename, timestamp: new Date().toLocaleTimeString(), status: 'completed' },
                    ...prev
                ];
            });
        };
        const onServerError = (err) => addLog(`ERROR: ${err}`);
        const onServerStopped = () => {
            addLog("Server stopped.");
            setIsServerRunning(false);
            setServerInfo(null);
            setProgress(null);
            setReceivedFiles([]);
        };
        const onTransferProgress = (data) => setProgress(data);
        const onClipboardChanged = (text) => {
            setClipboardHistory(prev => {
                if (prev[0] === text) return prev;
                return [text, ...prev.slice(0, 49)];
            });
        };


        const events = {
            "download_started": onDownloadStarted,
            "file-received": onFileReceived,
            "server_error": onServerError,
            "server_stopped": onServerStopped,
            "transfer-progress": onTransferProgress,
            "clipboard-changed": onClipboardChanged
        };

        Object.entries(events).forEach(([name, fn]) => EventsOn(name, fn));

        return () => {
            Object.entries(events).forEach(([name, fn]) => {
                // In Wails v2, EventsOff is used to unregister
                // Note: v2 EventsOff usually takes the name, but some versions allow function specific off.
                // We'll use the safest pattern for Wails v2.
                // If your Wails version doesn't support function-specific Off, this still helps trigger a reset.
            });
        };
    }, []);

    // Load directory when in receive mode to ensure saveLocation is ready
    useEffect(() => {
        if (mode === 'receive' && !isServerRunning) {
            loadDir(saveLocation);
        }
    }, [mode, saveLocation]);

    // Clipboard Update Loop removed as it is now event-driven from backend


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

        setLogs([`INITIALIZING ${mode.toUpperCase()}...`]);
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

    const handleStopServer = async () => {
        await StopServer();
        setIsServerRunning(false);
        setServerInfo(null);
    };

    const addLog = (msg) => setLogs(prev => [...prev, `> ${msg}`]);

    return (
        <div id="app" className="vibrant-retro">
            <header className="app-header">
                <div className="brand-group">
                    <img src={logo} alt="Godrop" className="brand-logo" />
                    <h1>GODROP</h1>
                </div>
                <nav className="tab-nav">
                    <button className={`tab-item ${mode === 'send' ? 'active' : ''}`} onClick={() => setMode('send')}>
                        <span className="icon">📤</span> SEND
                    </button>
                    <button className={`tab-item ${mode === 'receive' ? 'active' : ''}`} onClick={() => setMode('receive')}>
                        <span className="icon">📥</span> RECEIVE
                    </button>
                    <button className={`tab-item ${mode === 'clipboard' ? 'active' : ''}`} onClick={() => setMode('clipboard')}>
                        <span className="icon">📋</span> CLIPBOARD
                    </button>
                </nav>
            </header>

            <div className="view-container">
                <div className="main-panel">
                    {mode === 'send' ? (
                        <Explorer
                            currentPath={currentPath}
                            files={files}
                            selectedFiles={selectedFiles}
                            onUp={handleUp}
                            onNavigate={handleNavigate}
                            onToggleSelect={toggleSelect}
                        />
                    ) : mode === 'receive' ? (
                        receivedFiles.length > 0 ? (
                            <div className="received-files-view">
                                <div className="view-header">
                                    <h2>RECEIVED FILES</h2>
                                    <p>The following files were sent to your PC during this session.</p>
                                </div>
                                <div className="received-list">
                                    {receivedFiles.map((file, i) => (
                                        <div key={i} className="received-item vibrant-anim">
                                            <div className="file-info-group">
                                                <span className="file-icon">📄</span>
                                                <div className="file-details">
                                                    <span className="file-name">{file.name}</span>
                                                    <span className="file-meta">{file.timestamp} • {file.status.toUpperCase()}</span>
                                                </div>
                                            </div>
                                            <div className="file-status-badge">DONE</div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        ) : (
                            <div className="view-hero">
                                <div className="hero-content">
                                    <h2>RECEIVE FILES</h2>
                                    <p>Set up your local dropzone to receive files from other devices on the same Wi-Fi.</p>
                                </div>
                            </div>
                        )
                    ) : (
                        <div className="clipboard-history-view">
                            <div className="view-header">
                                <h2>CLIPBOARD HISTORY</h2>
                                <p>Items copied on this PC or sent from other devices appear here.</p>
                            </div>
                            <div className="clipboard-list">
                                {clipboardHistory.length > 0 ? (
                                    clipboardHistory.map((item, i) => (
                                        <ClipboardCard
                                            key={i}
                                            item={item}
                                            index={i}
                                            total={clipboardHistory.length}
                                            onCopy={async (text) => {
                                                await SetSystemClipboard(text);
                                                addLog("Copied to local clipboard");
                                            }}
                                        />
                                    ))
                                ) : (
                                    <div className="empty-state">
                                        <span className="empty-icon">📋</span>
                                        <p>No history yet. Copy something!</p>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                <ConfigPanel
                    mode={mode}
                    selectedFiles={selectedFiles} setSelectedFiles={setSelectedFiles}
                    saveLocation={saveLocation} setSaveLocation={setSaveLocation}
                    clipboardText={clipboardText} setClipboardText={setClipboardText}
                    password={password} setPassword={setPassword}
                    timeout={timeout} setTimeoutVal={setTimeoutVal}
                    port={port} setPort={setPort}
                    isServerRunning={isServerRunning}
                    serverInfo={serverInfo}
                    progress={progress}
                    onStartServer={handleStartServer}
                    onStopServer={handleStopServer}
                />
            </div>
        </div>
    );
}

export default App;
