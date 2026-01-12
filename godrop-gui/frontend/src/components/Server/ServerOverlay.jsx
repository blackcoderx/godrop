import { RetroProgressBar } from '../Common/RetroProgressBar';
import { StopServer } from '../../../wailsjs/go/main/App';

export const ServerOverlay = ({
    isServerRunning, serverInfo, mode, connectivity, logs, progress, onStop
}) => {
    if (!isServerRunning || !serverInfo) return null;

    return (
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
                <button className="btn-stop" onClick={onStop}>STOP SERVER</button>
            </div>
        </div>
    );
};
