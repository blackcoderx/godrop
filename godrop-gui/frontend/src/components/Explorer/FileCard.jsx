export const FileCard = ({ file, isSelected, onToggleSelect, onNavigate }) => {
    const getIcon = () => {
        if (file.isDir) return '📁';
        if (file.name.match(/\.(jpg|jpeg|png|gif)$/i)) return '🖼️';
        if (file.name.endsWith('.pdf')) return '📄';
        return '📦';
    };

    return (
        <div
            className={`file-card ${isSelected ? 'selected' : ''}`}
            onClick={onToggleSelect}
            onDoubleClick={onNavigate}
        >
            <div className="file-icon">{getIcon()}</div>
            <div className="file-name">{file.name}</div>
        </div>
    );
};
