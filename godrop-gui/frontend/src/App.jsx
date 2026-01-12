import { useState, useEffect } from 'react';
import './App.css';
import logo from './assets/images/godrop-logo.png';
import { GetHomeDir, ReadDir, StartServer, StopServer, StartReceiveServer, StartClipboardServer, GetDefaultSaveDir, GetSystemClipboard } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

// Components
import { Explorer } from './components/Explorer/Explorer';
import { ConfigPanel } from './components/Config/ConfigPanel';
import { ServerOverlay } from './components/Server/ServerOverlay';

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
                    ) : (
                        <div className="view-hero">
                            <div className="hero-content">
                                <h2>{mode === 'receive' ? 'RECEIVE FILES' : 'CLIPBOARD SYNC'}</h2>
                                <p>
                                    {mode === 'receive'
                                        ? 'Set up your local dropzone to receive files from other devices on the same Wi-Fi.'
                                        : 'Automatically sync your clipboard with other devices. Minimal effort, maximum speed.'}
                                </p>
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
                    onStartServer={handleStartServer}
                />
            </div>

            <ServerOverlay
                isServerRunning={isServerRunning}
                serverInfo={serverInfo}
                mode={mode}
                connectivity="local"
                logs={logs}
                progress={progress}
                onStop={handleStopServer}
            />
        </div>
    );
}

export default App;
