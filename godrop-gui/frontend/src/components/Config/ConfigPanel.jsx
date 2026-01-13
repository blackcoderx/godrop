import { FileSelection } from './FileSelection';
import { SelectDirectory, SetSystemClipboard } from '../../../wailsjs/go/main/App';
import { RetroProgressBar } from '../Common/RetroProgressBar';

export const ConfigPanel = ({
    mode,
    selectedFiles, setSelectedFiles,
    saveLocation, setSaveLocation,
    clipboardText, setClipboardText,
    password, setPassword,
    timeout, setTimeoutVal,
    port, setPort,
    isServerRunning,
    serverInfo,
    progress,
    onStartServer,
    onStopServer
}) => {
    const basename = (path) => path.split(/[/\\]/).pop();

    return (
        <aside className="config-panel">
            <div className="retro-card">
                <div className="section-label">⚙️ SETUP</div>

                {!isServerRunning ? (
                    <>
                        {mode === 'send' && (
                            <FileSelection
                                selectedFiles={selectedFiles}
                                onRemove={(path) => setSelectedFiles(selectedFiles.filter(p => p !== path))}
                            />
                        )}

                        {mode === 'receive' && (
                            <div className="config-group">
                                <label className="input-label">📁 Dropzone</label>
                                <div className="path-display" onClick={async () => {
                                    const dir = await SelectDirectory();
                                    if (dir) setSaveLocation(dir);
                                }}>
                                    {basename(saveLocation) || "Select Path..."}
                                </div>
                            </div>
                        )}

                        {mode === 'clipboard' && (
                            <div className="config-group" style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
                                <label className="input-label">📝 Add to History</label>
                                <textarea
                                    className="input-ui clipboard-area"
                                    value={clipboardText}
                                    onChange={(e) => setClipboardText(e.target.value)}
                                    placeholder="Type here to add a manual entry..."
                                />
                                <button className="btn-secondary" onClick={async () => {
                                    if (!clipboardText.trim()) return;
                                    await SetSystemClipboard(clipboardText);
                                    setClipboardText("");
                                }}>
                                    Add to History
                                </button>
                            </div>
                        )}

                        <div className="config-grid">
                            <div className="input-block">
                                <label className="input-label">🔒 Pass</label>
                                <input type="password" placeholder="none" className="input-ui" value={password} onChange={e => setPassword(e.target.value)} />
                            </div>
                            <div className="input-block">
                                <label className="input-label">🔌 Port</label>
                                <input type="number" className="input-ui" value={port} onChange={e => setPort(e.target.value)} />
                            </div>
                        </div>

                        <div className="input-block">
                            <label className="input-label">⏳ Timeout (min)</label>
                            <div className="timeout-input-wrapper">
                                <input
                                    type="number"
                                    className="input-ui"
                                    value={timeout}
                                    onChange={e => setTimeoutVal(parseInt(e.target.value) || 0)}
                                />
                                {timeout === 0 && <span className="infinity-badge">∞ INFINITY</span>}
                            </div>
                        </div>
                    </>
                ) : (
                    serverInfo && (
                        <div className="sidebar-session-info animated slideInUp">
                            <div className="sidebar-qr-container">
                                <div className="sidebar-qr-frame">
                                    <img src={serverInfo.qrCode} alt="QR Code" className="sidebar-qr-image" />
                                </div>
                                <div className="sidebar-url-box">
                                    <code>{serverInfo.fullUrl}</code>
                                </div>
                            </div>

                            {progress && (
                                <div className="sidebar-progress-section">
                                    <div className="progress-header">
                                        <span>TRANSFERRING...</span>
                                        <span>{progress.percent}%</span>
                                    </div>
                                    <RetroProgressBar percent={progress.percent} />
                                </div>
                            )}
                        </div>
                    )
                )}

                <button
                    className={`btn-primary ${isServerRunning ? 'stop-btn pulse' : ''}`}
                    onClick={isServerRunning ? onStopServer : onStartServer}
                    disabled={!isServerRunning && (mode === 'send' && selectedFiles.length === 0)}
                    style={{ marginTop: isServerRunning ? '20px' : 'auto' }}
                >
                    {isServerRunning ? 'STOP SERVER' : 'START SERVER'}
                </button>
            </div>
        </aside>
    );
};
