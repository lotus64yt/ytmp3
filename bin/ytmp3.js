#!/usr/bin/env node
const os = require('os');
const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');
const https = require('https');
const tar = require('tar');

const VERSION = require('../package.json').version;
const REPO = 'lotus64yt/ytmp3';

function getPlatform() {
    const p = os.platform();
    if (p === 'win32') return 'windows';
    return p;
}

function getArch() {
    const a = os.arch();
    if (a === 'x64') return 'amd64';
    return a;
}

const platform = getPlatform();
const arch = getArch();
const ext = platform === 'windows' ? '.exe' : '';
const binName = `ytmp3${ext}`;

const cacheDir = path.join(os.homedir(), '.ytmp3-cli');
if (!fs.existsSync(cacheDir)) {
    fs.mkdirSync(cacheDir, { recursive: true });
}

const binPath = path.join(cacheDir, binName);

function downloadAndExtract(url) {
    return new Promise((resolve, reject) => {
        https.get(url, (res) => {
            if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
                return downloadAndExtract(res.headers.location).then(resolve).catch(reject);
            }
            if (res.statusCode !== 200) {
                return reject(new Error(`HTTP Status Code: ${res.statusCode}`));
            }

            res.pipe(tar.x({ cwd: cacheDir }))
                .on('finish', resolve)
                .on('error', reject);
        }).on('error', reject);
    });
}

async function run() {
    if (!fs.existsSync(binPath)) {
        console.log(`[ytmp3-cli] Downloading binary for ${platform}-${arch}...`);
        const url = `https://github.com/${REPO}/releases/download/v${VERSION}/ytmp3_${VERSION}_${platform}_${arch}.tar.gz`;
        try {
            await downloadAndExtract(url);
            if (platform !== 'windows') {
                fs.chmodSync(binPath, 0o755);
            }
            console.log(`[ytmp3-cli] Download complete!`);
        } catch (e) {
            console.error('[ytmp3-cli] Failed to download binary:', e.message);
            process.exit(1);
        }
    }

    const args = process.argv.slice(2);
    const result = spawnSync(binPath, args, { stdio: 'inherit' });
    process.exit(result.status || 0);
}

run();