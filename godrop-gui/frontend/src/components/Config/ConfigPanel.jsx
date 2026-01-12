import { FileSelection } from './FileSelection';
import { SelectDirectory, SetSystemClipboard } from '../../../wailsjs/go/main/App';

export const ConfigPanel = ({
    mode,
    selectedFiles, setSelectedFiles,
    saveLocation, setSaveLocation,
    clipboardText, setClipboardText,
    password, setPassword,
    timeout, setTimeoutVal,
    port, setPort,
    isServerRunning,
    onStartServer
}) => {
    const basename = (path) => path.split(/[/\\]/).pop();

    return (
        <aside className="config-panel">
            <div className="retro-card">
                <div className="section-label">⚙️ SETUP</div>

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
                        <label className="input-label">📋 Local Clipboard</label>
                        <textarea
                            className="input-ui clipboard-area"
                            value={clipboardText}
                            onChange={(e) => setClipboardText(e.target.value)}
                            placeholder="Type or paste here to sync..."
                        />
                        <button className="btn-secondary" onClick={() => SetSystemClipboard(clipboardText)}>
                            Update Local
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

                <button
                    className={`btn-primary ${isServerRunning ? 'pulse' : ''}`}
                    onClick={onStartServer}
                    disabled={isServerRunning || (mode === 'send' && selectedFiles.length === 0)}
                >
                    {isServerRunning ? '🚀 LIVE' : 'START SERVER'}
                </button>
            </div>
        </aside>
    );
};
