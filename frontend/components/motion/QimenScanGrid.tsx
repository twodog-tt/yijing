export default function QimenScanGrid({ compact = false }: { compact?: boolean }) {
  return (
    <div
      className={`qimen-scan-grid ${compact ? "qimen-scan-grid--compact" : ""}`}
      aria-hidden
    >
      <div className="qimen-scan-grid__scan-h" />
      <div className="qimen-scan-grid__scan-v" />
      <div className="qimen-scan-grid__dot" />
      <div className="qimen-scan-grid__cells">
        {Array.from({ length: 9 }).map((_, index) => (
          <div key={index} className="qimen-scan-grid__cell" />
        ))}
      </div>
    </div>
  );
}
