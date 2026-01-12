export const RetroProgressBar = ({ percent }) => {
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
