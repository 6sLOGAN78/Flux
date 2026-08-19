import React, { useState, useRef } from 'react';
import {
  Download,
  Copy,
  Check,
  Sparkles,
  QrCode,
  Palette,
  Layers,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Input } from '@/components/ui/Input';
import { cn } from '@/lib/utils';

export interface QRStudioCanvasProps {
  url?: string;
  initialFgColor?: string;
  initialBgColor?: string;
  initialLogoText?: string;
  className?: string;
}

const COLOR_PRESETS = [
  { name: 'Charcoal', hex: '#09090b' },
  { name: 'Emerald', hex: '#059669' },
  { name: 'Blue', hex: '#2563eb' },
  { name: 'Violet', hex: '#7c3aed' },
  { name: 'Rose', hex: '#e11d48' },
];

export function QRStudioCanvas({
  url = 'https://flux.to/v2-launch',
  initialFgColor = '#09090b',
  initialBgColor = '#ffffff',
  initialLogoText = 'F',
  className,
}: QRStudioCanvasProps) {
  const [targetUrl, setTargetUrl] = useState(url);
  const [fgColor, setFgColor] = useState(initialFgColor);
  const [bgColor, setBgColor] = useState(initialBgColor);
  const [logoText, setLogoText] = useState(initialLogoText);
  const [isCopied, setIsCopied] = useState(false);

  // Generate deterministic QR matrix pattern based on url hash
  const generatePattern = (input: string) => {
    let hash = 0;
    for (let i = 0; i < input.length; i++) {
      hash = (hash << 5) - hash + input.charCodeAt(i);
      hash |= 0;
    }

    const size = 21;
    const matrix: boolean[][] = Array.from({ length: size }, () =>
      Array.from({ length: size }, () => false)
    );

    // Corner Finder Patterns (7x7)
    const placeFinder = (rStart: number, cStart: number) => {
      for (let r = 0; r < 7; r++) {
        for (let c = 0; c < 7; c++) {
          const isBorder = r === 0 || r === 6 || c === 0 || c === 6;
          const isCenter = r >= 2 && r <= 4 && c >= 2 && c <= 4;
          matrix[rStart + r][cStart + c] = isBorder || isCenter;
        }
      }
    };

    placeFinder(0, 0); // Top-left
    placeFinder(0, size - 7); // Top-right
    placeFinder(size - 7, 0); // Bottom-left

    // Pseudo-random data bits
    for (let r = 0; r < size; r++) {
      for (let c = 0; c < size; c++) {
        // Skip corner finder areas
        if (
          (r < 8 && c < 8) ||
          (r < 8 && c >= size - 8) ||
          (r >= size - 8 && c < 8)
        ) {
          continue;
        }
        // Skip center logo area
        if (r >= 8 && r <= 12 && c >= 8 && c <= 12) {
          continue;
        }

        const val = Math.abs(Math.sin((r * size + c) * 1.5 + hash)) > 0.45;
        matrix[r][c] = val;
      }
    }

    return matrix;
  };

  const matrix = generatePattern(targetUrl);
  const matrixSize = matrix.length;
  const cellSize = 12;
  const canvasPixelSize = matrixSize * cellSize;

  const handleDownloadSVG = () => {
    const svgContent = `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${canvasPixelSize} ${canvasPixelSize}" width="512" height="512">
  <rect width="100%" height="100%" fill="${bgColor}"/>
  ${matrix
    .map((row, r) =>
      row
        .map((cell, c) =>
          cell
            ? `<rect x="${c * cellSize}" y="${r * cellSize}" width="${cellSize}" height="${cellSize}" fill="${fgColor}" rx="2"/>`
            : ''
        )
        .join('')
    )
    .join('')}
  <circle cx="${canvasPixelSize / 2}" cy="${canvasPixelSize / 2}" r="22" fill="${bgColor}" stroke="${fgColor}" stroke-width="2"/>
  <text x="${canvasPixelSize / 2}" y="${canvasPixelSize / 2 + 5}" font-family="sans-serif" font-weight="bold" font-size="14" fill="${fgColor}" text-anchor="middle">${logoText}</text>
</svg>`;

    const blob = new Blob([svgContent], { type: 'image/svg+xml' });
    const downloadUrl = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = `flux-qr-${Date.now()}.svg`;
    link.click();
    URL.revokeObjectURL(downloadUrl);
  };

  const handleDownloadPNG = () => {
    const canvas = document.createElement('canvas');
    canvas.width = 512;
    canvas.height = 512;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.fillStyle = bgColor;
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    const scale = canvas.width / canvasPixelSize;
    ctx.fillStyle = fgColor;

    matrix.forEach((row, r) => {
      row.forEach((cell, c) => {
        if (cell) {
          ctx.fillRect(
            c * cellSize * scale,
            r * cellSize * scale,
            cellSize * scale,
            cellSize * scale
          );
        }
      });
    });

    // Center Logo
    const center = canvas.width / 2;
    ctx.fillStyle = bgColor;
    ctx.beginPath();
    ctx.arc(center, center, 44, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = fgColor;
    ctx.lineWidth = 4;
    ctx.stroke();

    ctx.fillStyle = fgColor;
    ctx.font = 'bold 28px sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(logoText, center, center);

    const pngUrl = canvas.toDataURL('image/png');
    const link = document.createElement('a');
    link.href = pngUrl;
    link.download = `flux-qr-${Date.now()}.png`;
    link.click();
  };

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col gap-8 lg:flex-row">
        {/* Left Side: QR Canvas Preview */}
        <div className="flex flex-col items-center justify-center rounded-xl border border-zinc-200 bg-zinc-50 p-6 dark:border-zinc-800 dark:bg-zinc-900/40">
          <div
            style={{ backgroundColor: bgColor }}
            className="relative flex items-center justify-center rounded-xl p-4 shadow-md transition-colors"
          >
            <svg
              viewBox={`0 0 ${canvasPixelSize} ${canvasPixelSize}`}
              className="h-60 w-60"
            >
              {matrix.map((row, r) =>
                row.map((cell, c) =>
                  cell ? (
                    <rect
                      key={`${r}-${c}`}
                      x={c * cellSize}
                      y={r * cellSize}
                      width={cellSize - 0.5}
                      height={cellSize - 0.5}
                      fill={fgColor}
                      rx={1.5}
                    />
                  ) : null
                )
              )}
            </svg>

            {/* Center Brand Badge */}
            <div
              style={{ backgroundColor: bgColor, borderColor: fgColor }}
              className="absolute flex h-10 w-10 items-center justify-center rounded-full border-2 shadow-xs transition-colors"
            >
              <span
                style={{ color: fgColor }}
                className="font-mono text-xs font-bold"
              >
                {logoText}
              </span>
            </div>
          </div>

          <div className="mt-4 text-center">
            <span className="font-mono text-xs text-zinc-500 dark:text-zinc-400">
              {targetUrl}
            </span>
          </div>

          {/* Export Action Buttons */}
          <div className="mt-6 flex w-full gap-2">
            <Button
              variant="outline"
              size="sm"
              className="flex-1"
              onClick={handleDownloadPNG}
              leftIcon={<Download className="h-3.5 w-3.5" />}
            >
              Download PNG
            </Button>
            <Button
              variant="primary"
              size="sm"
              className="flex-1"
              onClick={handleDownloadSVG}
              leftIcon={<Download className="h-3.5 w-3.5" />}
            >
              Download SVG
            </Button>
          </div>
        </div>

        {/* Right Side: QR Customization Studio Controls */}
        <div className="flex-1 space-y-6">
          <div>
            <div className="flex items-center gap-2">
              <QrCode className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
              <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
                QR Code Studio
              </h3>
            </div>
            <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
              Generate vector QR codes with customizable brand accents and center logos.
            </p>
          </div>

          <div className="space-y-4">
            <Input
              label="QR Target URL"
              value={targetUrl}
              onChange={(e) => setTargetUrl(e.target.value)}
              placeholder="https://flux.to/my-link"
            />

            {/* Foreground Color Control */}
            <div>
              <label className="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
                Foreground Color
              </label>
              <div className="flex items-center gap-3">
                <input
                  type="color"
                  value={fgColor}
                  onChange={(e) => setFgColor(e.target.value)}
                  className="h-8 w-8 cursor-pointer rounded-md border border-zinc-200 dark:border-zinc-800"
                />
                <div className="flex items-center gap-1.5">
                  {COLOR_PRESETS.map((p) => (
                    <button
                      key={p.name}
                      type="button"
                      onClick={() => setFgColor(p.hex)}
                      style={{ backgroundColor: p.hex }}
                      className="h-6 w-6 rounded-full border border-white/20 shadow-xs transition-transform hover:scale-110"
                      title={p.name}
                    />
                  ))}
                </div>
                <span className="font-mono text-xs text-zinc-400">{fgColor}</span>
              </div>
            </div>

            {/* Background Color Control */}
            <div>
              <label className="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
                Background Color
              </label>
              <div className="flex items-center gap-3">
                <input
                  type="color"
                  value={bgColor}
                  onChange={(e) => setBgColor(e.target.value)}
                  className="h-8 w-8 cursor-pointer rounded-md border border-zinc-200 dark:border-zinc-800"
                />
                <button
                  type="button"
                  onClick={() => setBgColor('#ffffff')}
                  className="rounded-md border border-zinc-200 bg-white px-2 py-1 text-xs font-medium text-zinc-700 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-300"
                >
                  White
                </button>
                <button
                  type="button"
                  onClick={() => setBgColor('#f4f4f5')}
                  className="rounded-md border border-zinc-200 bg-zinc-100 px-2 py-1 text-xs font-medium text-zinc-700 dark:border-zinc-800 dark:bg-zinc-800 dark:text-zinc-300"
                >
                  Zinc-100
                </button>
                <span className="font-mono text-xs text-zinc-400">{bgColor}</span>
              </div>
            </div>

            {/* Center Logo Character */}
            <div className="max-w-xs">
              <Input
                label="Center Brand Monogram"
                value={logoText}
                maxLength={3}
                onChange={(e) => setLogoText(e.target.value.toUpperCase())}
                placeholder="F"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default QRStudioCanvas;
