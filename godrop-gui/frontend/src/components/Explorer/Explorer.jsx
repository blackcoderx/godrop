import { FileCard } from './FileCard';

export const Explorer = ({ currentPath, files, selectedFiles, onUp, onNavigate, onToggleSelect }) => {
    return (
        <main className="explorer">
            <div className="top-bar">
                <button className="btn-icon" onClick={onUp}>↑</button>
                <div className="breadcrumb">
                    My PC / {currentPath.split(/[/\\]/).slice(-2).map((p, i) => (
                        <span key={i}>{p}{i === 0 ? ' / ' : ''}</span>
                    ))}
                </div>
            </div>

            <div className="file-grid">
                {files.map((file) => (
                    <FileCard
                        key={file.name}
                        file={file}
                        isSelected={selectedFiles.includes(file.img)}
                        onToggleSelect={() => onToggleSelect(file)}
                        onNavigate={() => onNavigate(file)}
                    />
                ))}
            </div>
        </main>
    );
};
