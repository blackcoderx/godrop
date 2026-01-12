import { FileSelection } from './FileSelection';
import { SelectDirectory, SetSystemClipboard } from '../../../wailsjs/go/main/App';

export const ConfigPanel = ({
    mode, setMode,
    connectivity, setConnectivity,
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
                <FileSelection
                    selectedFiles={selectedFiles}
                    onRemove={(path) => setSelectedFiles(selectedFiles.filter(p => p !== path))}
                />
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
                onClick={onStartServer}
                disabled={isServerRunning || (mode === 'send' && selectedFiles.length === 0)}
            >
                {isServerRunning ? 'SERVER LIVE' : 'Start Server'}
            </button>
        </aside>
    );
};
