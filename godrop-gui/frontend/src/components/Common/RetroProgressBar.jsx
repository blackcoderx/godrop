export const RetroProgressBar = ({ percent }) => {
    return (
        <div style={{
            width: '100%',
            height: '12px',
            background: '#eee',
            border: '2px solid var(--text-main)',
            position: 'relative',
            overflow: 'hidden',
            borderRadius: '6px'
        }}>
            <div style={{
                width: `${percent}%`,
                height: '100%',
                background: 'var(--accent)',
                transition: 'width 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                boxShadow: 'inset 0 2px 4px rgba(255,255,255,0.3)'
            }} />
        </div>
    );
};
