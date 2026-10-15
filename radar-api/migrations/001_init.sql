CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS probes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    flag_icon TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS endpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    http_enabled BOOLEAN NOT NULL DEFAULT true,
    icmp_enabled BOOLEAN NOT NULL DEFAULT true,
    probe_id INTEGER REFERENCES probes(id) ON DELETE SET NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE endpoints ADD COLUMN logo_icon TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    probe_id INTEGER NOT NULL REFERENCES probes(id) ON DELETE RESTRICT,
    token_hash TEXT NOT NULL UNIQUE,
    last_seen_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE agents ADD COLUMN public_ip TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS samples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    endpoint_id INTEGER NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    protocol TEXT NOT NULL CHECK (protocol IN ('http', 'icmp')),
    observed_at INTEGER NOT NULL,
    latency_ms REAL,
    ok BOOLEAN NOT NULL,
    CONSTRAINT samples_latency_check CHECK (
        (ok AND latency_ms IS NOT NULL AND latency_ms >= 0)
        OR (NOT ok AND latency_ms IS NULL)
    ),
    UNIQUE (agent_id, endpoint_id, protocol, observed_at)
);

CREATE INDEX IF NOT EXISTS samples_grid_idx
    ON samples (protocol, observed_at, endpoint_id, agent_id);

INSERT INTO probes (code, name, flag_icon) VALUES
    ('probe1', 'Irancell', '/flags/probe1.svg')
ON CONFLICT (code) DO UPDATE
SET name = excluded.name, flag_icon = excluded.flag_icon;

UPDATE endpoints SET probe_id = (SELECT id FROM probes WHERE code = 'probe1');
UPDATE agents SET probe_id = (SELECT id FROM probes WHERE code = 'probe1');
DELETE FROM probes WHERE code IN ('probe2', 'probe3');

WITH seed(name, host, http_enabled, icmp_enabled, logo_icon) AS (VALUES
    ('Google', 'google.com', true, true, '/logos/google.svg'),
    ('GitHub', 'github.com', true, true, '/logos/github.svg'),
    ('1.1.1.1', '1.1.1.1', true, true, '/logos/cloudflare.svg'),
    ('8.8.8.8', '8.8.8.8', true, true, '/logos/google-dns.svg'),
    ('YouTube', 'youtube.com', true, true, '/logos/youtube.svg'),
    ('ChatGPT', 'chatgpt.com', true, true, '/logos/chatgpt.svg'),
    ('Claude', 'claude.ai', true, true, '/logos/claude.svg'),
    ('DeepSeek', 'deepseek.com', true, true, '/logos/deepseek.svg'),
    ('Microsoft', 'microsoft.com', true, true, '/logos/microsoft.svg'),
    ('Apple', 'apple.com', true, true, '/logos/apple.svg'),
    ('Play Store', 'play.google.com', true, true, '/logos/play-store.svg'),
    ('App Store', 'apps.apple.com', true, true, '/logos/app-store.svg'),
    ('Docker', 'docker.com', true, true, '/logos/docker.svg'),
    ('Spotify', 'spotify.com', true, true, '/logos/spotify.svg'),
    ('Grok', 'grok.com', true, true, '/logos/grok.svg'),
    ('Arena', 'arena.ai', true, true, '/logos/arena.svg'),
    ('Ondpline', 'ondpline.com', true, true, '/logos/ondpline.svg'),
    ('GitLab', 'gitlab.com', true, true, '/logos/gitlab.svg'),
    ('Digikala', 'digikala.ir', true, true, '/logos/digikala.svg'),
    ('Snapp', 'snapp.ir', true, true, '/logos/snapp.svg'),
    ('Tapsi', 'tapsi.ir', true, true, '/logos/tapsi.svg'),
    ('Divar', 'divar.ir', true, true, '/logos/divar.svg'),
    ('Bale', 'bale.ai', true, true, '/logos/bale.svg'),
    ('Devin AI', 'devin.ai', true, true, '/logos/devin.svg'),
    ('Neshan', 'neshan.org', true, true, '/logos/neshan.svg'),
    ('Wikipedia', 'wikipedia.org', true, true, '/logos/wikipedia.svg'),
    ('WhatsApp', 'whatsapp.com', true, true, '/logos/whatsapp.svg'),
    ('Zoomit', 'zoomit.ir', true, true, '/logos/zoomit.svg'),
    ('Digiato', 'digiato.ir', true, true, '/logos/digiato.svg'),
    ('Let''s Encrypt', 'letsencrypt.org', true, true, '/logos/letsencrypt.svg'),
    ('Steam', 'steampowered.com', true, true, '/logos/steam.svg'),
    ('Torob', 'torob.ir', true, true, '/logos/torob.svg'),
    ('PlayStation', 'playstation.com', true, true, '/logos/playstation.svg'),
    ('Xbox', 'xbox.com', true, true, '/logos/xbox.svg'),
    ('Gmail', 'gmail.com', true, true, '/logos/gmail.svg'),
    ('Figma', 'figma.com', true, true, '/logos/figma.svg'),
    ('Windows Update', 'windowsupdate.microsoft.com', true, true, '/logos/windows-update.svg'),
    ('Bing', 'bing.com', true, true, '/logos/bing.svg'),
    ('Aparat', 'aparat.ir', true, true, '/logos/aparat.svg'),
    ('PyPI', 'pypi.org', true, true, '/logos/pypi.svg'),
    ('npm', 'registry.npmjs.org', true, true, '/logos/npm.svg'),
    ('WordPress', 'wordpress.org', true, true, '/logos/wordpress.svg'),
    ('Go modules', 'proxy.golang.org', true, true, '/logos/golang.svg'),
    ('Crates', 'crates.io', true, true, '/logos/crates.svg'),
    ('VS Code', 'update.code.visualstudio.com', true, true, '/logos/vscode.svg'),
    ('NuGet', 'api.nuget.org', true, true, '/logos/nuget.svg'),
    ('Instagram', 'instagram.com', true, true, '/logos/instagram.svg'),
    ('Telegram', 'telegram.org', true, true, '/logos/telegram.svg')
)
INSERT INTO endpoints (name, host, http_enabled, icmp_enabled, probe_id, active, logo_icon)
SELECT seed.name, seed.host, seed.http_enabled, seed.icmp_enabled, p.id, true, seed.logo_icon
FROM seed
JOIN probes p ON p.code = 'probe1'
WHERE NOT EXISTS (SELECT 1 FROM endpoints e WHERE e.host = seed.host);

WITH seed(name, host, http_enabled, icmp_enabled, logo_icon) AS (VALUES
    ('Google', 'google.com', true, true, '/logos/google.svg'),
    ('GitHub', 'github.com', true, true, '/logos/github.svg'),
    ('1.1.1.1', '1.1.1.1', true, true, '/logos/cloudflare.svg'),
    ('8.8.8.8', '8.8.8.8', true, true, '/logos/google-dns.svg'),
    ('YouTube', 'youtube.com', true, true, '/logos/youtube.svg'),
    ('ChatGPT', 'chatgpt.com', true, true, '/logos/chatgpt.svg'),
    ('Claude', 'claude.ai', true, true, '/logos/claude.svg'),
    ('DeepSeek', 'deepseek.com', true, true, '/logos/deepseek.svg'),
    ('Microsoft', 'microsoft.com', true, true, '/logos/microsoft.svg'),
    ('Apple', 'apple.com', true, true, '/logos/apple.svg'),
    ('Play Store', 'play.google.com', true, true, '/logos/play-store.svg'),
    ('App Store', 'apps.apple.com', true, true, '/logos/app-store.svg'),
    ('Docker', 'docker.com', true, true, '/logos/docker.svg'),
    ('Spotify', 'spotify.com', true, true, '/logos/spotify.svg'),
    ('Grok', 'grok.com', true, true, '/logos/grok.svg'),
    ('Arena', 'arena.ai', true, true, '/logos/arena.svg'),
    ('Ondpline', 'ondpline.com', true, true, '/logos/ondpline.svg'),
    ('GitLab', 'gitlab.com', true, true, '/logos/gitlab.svg'),
    ('Digikala', 'digikala.ir', true, true, '/logos/digikala.svg'),
    ('Snapp', 'snapp.ir', true, true, '/logos/snapp.svg'),
    ('Tapsi', 'tapsi.ir', true, true, '/logos/tapsi.svg'),
    ('Divar', 'divar.ir', true, true, '/logos/divar.svg'),
    ('Bale', 'bale.ai', true, true, '/logos/bale.svg'),
    ('Devin AI', 'devin.ai', true, true, '/logos/devin.svg'),
    ('Neshan', 'neshan.org', true, true, '/logos/neshan.svg'),
    ('Wikipedia', 'wikipedia.org', true, true, '/logos/wikipedia.svg'),
    ('WhatsApp', 'whatsapp.com', true, true, '/logos/whatsapp.svg'),
    ('Zoomit', 'zoomit.ir', true, true, '/logos/zoomit.svg'),
    ('Digiato', 'digiato.ir', true, true, '/logos/digiato.svg'),
    ('Let''s Encrypt', 'letsencrypt.org', true, true, '/logos/letsencrypt.svg'),
    ('Steam', 'steampowered.com', true, true, '/logos/steam.svg'),
    ('Torob', 'torob.ir', true, true, '/logos/torob.svg'),
    ('PlayStation', 'playstation.com', true, true, '/logos/playstation.svg'),
    ('Xbox', 'xbox.com', true, true, '/logos/xbox.svg'),
    ('Gmail', 'gmail.com', true, true, '/logos/gmail.svg'),
    ('Figma', 'figma.com', true, true, '/logos/figma.svg'),
    ('Windows Update', 'windowsupdate.microsoft.com', true, true, '/logos/windows-update.svg'),
    ('Bing', 'bing.com', true, true, '/logos/bing.svg'),
    ('Aparat', 'aparat.ir', true, true, '/logos/aparat.svg'),
    ('PyPI', 'pypi.org', true, true, '/logos/pypi.svg'),
    ('npm', 'registry.npmjs.org', true, true, '/logos/npm.svg'),
    ('WordPress', 'wordpress.org', true, true, '/logos/wordpress.svg'),
    ('Go modules', 'proxy.golang.org', true, true, '/logos/golang.svg'),
    ('Crates', 'crates.io', true, true, '/logos/crates.svg'),
    ('VS Code', 'update.code.visualstudio.com', true, true, '/logos/vscode.svg'),
    ('NuGet', 'api.nuget.org', true, true, '/logos/nuget.svg'),
    ('Instagram', 'instagram.com', true, true, '/logos/instagram.svg'),
    ('Telegram', 'telegram.org', true, true, '/logos/telegram.svg')
)
UPDATE endpoints SET
    name = seed.name,
    http_enabled = seed.http_enabled,
    icmp_enabled = seed.icmp_enabled,
    probe_id = p.id,
    active = true,
    logo_icon = seed.logo_icon
FROM seed
JOIN probes p ON p.code = 'probe1'
WHERE endpoints.host = seed.host;

UPDATE endpoints SET active = false
WHERE host NOT IN (
    'google.com', 'github.com', '1.1.1.1', '8.8.8.8', 'youtube.com',
    'chatgpt.com', 'claude.ai', 'deepseek.com', 'microsoft.com', 'apple.com',
    'play.google.com', 'apps.apple.com', 'docker.com', 'spotify.com',
    'grok.com', 'arena.ai', 'ondpline.com',
    'gitlab.com', 'digikala.ir', 'snapp.ir', 'tapsi.ir', 'divar.ir',
    'bale.ai', 'devin.ai', 'neshan.org', 'wikipedia.org',
    'whatsapp.com', 'zoomit.ir', 'digiato.ir', 'letsencrypt.org', 'steampowered.com',
    'torob.ir', 'playstation.com', 'xbox.com', 'gmail.com', 'figma.com',
    'windowsupdate.microsoft.com', 'bing.com', 'aparat.ir', 'pypi.org',
    'registry.npmjs.org', 'wordpress.org', 'proxy.golang.org', 'crates.io',
    'update.code.visualstudio.com', 'api.nuget.org', 'instagram.com', 'telegram.org'
);
