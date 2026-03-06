#!/usr/bin/env node

/**
 * iTaK Shield - npm binary installer
 * Downloads the correct platform binary from GitHub Releases on postinstall.
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');

const VERSION = require('./package.json').version;
const REPO = 'David2024patton/itak-shield';

const PLATFORM_MAP = {
    'win32-x64': 'itak-shield-windows-amd64.exe',
    'darwin-x64': 'itak-shield-darwin-amd64',
    'darwin-arm64': 'itak-shield-darwin-arm64',
    'linux-x64': 'itak-shield-linux-amd64'
};

const platform = os.platform();
const arch = os.arch();
const key = `${platform}-${arch}`;
const binaryName = PLATFORM_MAP[key];

if (!binaryName) {
    console.error(`[itak-shield] Unsupported platform: ${platform}-${arch}`);
    console.error(`[itak-shield] Supported: ${Object.keys(PLATFORM_MAP).join(', ')}`);
    process.exit(1);
}

const binDir = path.join(__dirname, 'bin');
const destName = platform === 'win32' ? 'itak-shield.exe' : 'itak-shield';
const destPath = path.join(binDir, destName);

// Ensure bin/ directory exists
if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
}

const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;

console.log(`[itak-shield] Downloading ${binaryName} for ${platform}-${arch}...`);

function download(downloadUrl, dest, redirectCount) {
    if (redirectCount > 5) {
        console.error('[itak-shield] Too many redirects');
        process.exit(1);
    }

    const mod = downloadUrl.startsWith('https') ? https : require('http');
    mod.get(downloadUrl, { headers: { 'User-Agent': 'itak-shield-npm' } }, (res) => {
        // Follow redirects (GitHub uses 302s)
        if (res.statusCode === 301 || res.statusCode === 302) {
            download(res.headers.location, dest, redirectCount + 1);
            return;
        }

        if (res.statusCode !== 200) {
            console.error(`[itak-shield] Download failed: HTTP ${res.statusCode}`);
            console.error(`[itak-shield] URL: ${downloadUrl}`);
            console.error(`[itak-shield] You can download manually from:`);
            console.error(`[itak-shield]   https://github.com/${REPO}/releases/tag/v${VERSION}`);
            process.exit(1);
        }

        const file = fs.createWriteStream(dest);
        res.pipe(file);
        file.on('finish', () => {
            file.close();
            // Make executable on unix
            if (platform !== 'win32') {
                fs.chmodSync(dest, 0o755);
            }
            const sizeMB = (fs.statSync(dest).size / 1024 / 1024).toFixed(1);
            console.log(`[itak-shield] Installed ${binaryName} (${sizeMB} MB)`);
        });
    }).on('error', (err) => {
        console.error(`[itak-shield] Download error: ${err.message}`);
        console.error(`[itak-shield] You can download manually from:`);
        console.error(`[itak-shield]   https://github.com/${REPO}/releases/tag/v${VERSION}`);
        process.exit(1);
    });
}

download(url, destPath, 0);
