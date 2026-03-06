#!/usr/bin/env node

/**
 * iTaK Shield - npm binary installer
 * Copies the correct platform binary to bin/ on postinstall.
 */

const fs = require('fs');
const path = require('path');
const os = require('os');

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

const srcPath = path.join(__dirname, 'dist', binaryName);
const binDir = path.join(__dirname, 'bin');
const destName = platform === 'win32' ? 'itak-shield.exe' : 'itak-shield';
const destPath = path.join(binDir, destName);

// Ensure bin/ directory exists
if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
}

// Copy binary
try {
    fs.copyFileSync(srcPath, destPath);
    // Make executable on unix
    if (platform !== 'win32') {
        fs.chmodSync(destPath, 0o755);
    }
    console.log(`[itak-shield] Installed ${binaryName} for ${platform}-${arch}`);
} catch (err) {
    console.error(`[itak-shield] Failed to install binary: ${err.message}`);
    process.exit(1);
}
