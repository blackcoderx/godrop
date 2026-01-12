import logo from '../../assets/images/godrop-logo.png';

export const Sidebar = ({ mode, setMode }) => {
    return (
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
    );
};
