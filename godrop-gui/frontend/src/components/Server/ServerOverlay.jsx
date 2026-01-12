import { RetroProgressBar } from '../Common/RetroProgressBar';

export const ServerOverlay = ({
    isServerRunning, serverInfo, mode, logs, progress, onStop
}) => {
    if (!isServerRunning || !serverInfo) return null;

    return (
        <div className={`server-overlay ${mode === 'receive' ? 'mini' : ''}`}>
            <div className="server-card vibrant-anim">
                <div className="card-header">
                    <span className="live-badge">● LIVE</span>
                    <span className="mode-badge">{mode.toUpperCase()}</span>
                </div>

                <div className="card-body-mini">
                    <div className="console-section">
                        <div className="console-header">SYSTEM LOGS</div>
                        <div className="log-scroll">
                            {logs.map((log, i) => (
                                <div key={i} className="log-entry">{log}</div>
                            ))}
                        </div>
                    </div>
                </div>

                {progress && (
                    <div className="progress-section">
                        <div className="progress-info">
                            <span>TRANSFERRING...</span>
                            <span>{progress.percent}%</span>
                        </div>
                        <RetroProgressBar percent={progress.percent} />
                        <div className="progress-stats">
                            {Math.round(progress.transferred / 1024 / 1024 * 10) / 10} MB / {Math.round(progress.total / 1024 / 1024 * 10) / 10} MB
                        </div>
                    </div>
                )}

                <button className="btn-primary stop-btn" onClick={onStop}>
                    STOP SERVER
                </button>
            </div>
        </div>
    );
};
