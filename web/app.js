/* ─────────────────────────────────────────────
   iTaK Shield GUI - Client-side Logic
   ───────────────────────────────────────────── */

// ─── Complete Provider Registry ──────────────
// Every provider entry has:
//   name     - Display name
//   url      - Default API base URL
//   keyHint  - Placeholder text for the API key input
//   icon     - Emoji or symbol
//   category - Which section it belongs in
//   instructions - Setup steps shown on the dashboard
//
// The proxy works with ALL of these because they all speak
// OpenAI-compatible (or similar REST) protocols.

const PROVIDERS = {
    // ── Foundation Model Developers ────────────
    openai: {
        name: 'OpenAI',
        url: 'https://api.openai.com',
        keyHint: 'Starts with sk-...',
        icon: '/icons/openai.svg',
        category: 'foundation',
        instructions: [
            'Open your AI tool\'s settings',
            'Find the "API Base URL" or "OpenAI URL" setting',
            'Change it to <code id="instrUrl1"></code>',
            'Keep your API key the same. Save and go.'
        ]
    },
    anthropic: {
        name: 'Anthropic',
        url: 'https://api.anthropic.com',
        keyHint: 'Starts with sk-ant-...',
        icon: '/icons/anthropic.svg',
        category: 'foundation',
        instructions: [
            'Open your AI tool\'s settings',
            'Find the Anthropic API URL setting',
            'Change it to <code id="instrUrl1"></code>',
            'Keep your API key the same. Save and go.'
        ]
    },
    gemini: {
        name: 'Google Gemini',
        url: 'https://generativelanguage.googleapis.com',
        keyHint: 'Starts with AIza...',
        icon: '/icons/gemini.svg',
        category: 'foundation',
        instructions: [
            'Open your AI tool\'s settings',
            'Find the Google API endpoint setting',
            'Change it to <code id="instrUrl1"></code>',
            'Keep your API key the same. Save and go.'
        ]
    },
    xai: {
        name: 'xAI (Grok)',
        url: 'https://api.x.ai',
        keyHint: 'Your xAI API key',
        icon: '/icons/xai.svg',
        category: 'foundation',
        instructions: [
            'Open your AI tool\'s settings',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Enter your xAI API key. Save and go.'
        ]
    },
    deepseek: {
        name: 'DeepSeek',
        url: 'https://api.deepseek.com',
        keyHint: 'Your DeepSeek API key',
        icon: '/icons/deepseek.svg',
        category: 'foundation',
        instructions: [
            'Open your AI tool\'s settings',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Enter your DeepSeek API key. Save and go.'
        ]
    },
    mistral: {
        name: 'Mistral AI',
        url: 'https://api.mistral.ai',
        keyHint: 'Your Mistral API key',
        icon: '/icons/mistral.svg',
        category: 'foundation',
        instructions: [
            'Open your AI tool\'s settings',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Enter your Mistral API key. Save and go.'
        ]
    },
    cohere: {
        name: 'Cohere',
        url: 'https://api.cohere.com',
        keyHint: 'Your Cohere API key',
        icon: '/icons/cohere.svg',
        category: 'foundation',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Enter your Cohere API key. Save and go.'
        ]
    },
    nvidia: {
        name: 'NVIDIA NIM',
        url: 'https://integrate.api.nvidia.com',
        keyHint: 'Your NVIDIA API key',
        icon: '/icons/nvidia.svg',
        category: 'foundation',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Uses OpenAI-compatible format. Enter your NVIDIA API key.'
        ]
    },
    qwen: {
        name: 'Qwen (Alibaba)',
        url: 'https://dashscope.aliyuncs.com/compatible-mode',
        keyHint: 'Your DashScope API key',
        icon: '/icons/qwen.svg',
        category: 'foundation',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Uses OpenAI-compatible mode. Enter your DashScope key.'
        ]
    },
    kimi: {
        name: 'Kimi (Moonshot)',
        url: 'https://api.moonshot.cn',
        keyHint: 'Your Moonshot API key',
        icon: '/icons/kimi.svg',
        category: 'foundation',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Enter your Moonshot API key. Save and go.'
        ]
    },
    zhipu: {
        name: 'Zhipu AI (GLM)',
        url: 'https://open.bigmodel.cn/api/paas',
        keyHint: 'Your Zhipu API key',
        icon: '/icons/zhipu.svg',
        category: 'foundation',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Enter your Zhipu API key. Save and go.'
        ]
    },
    meta: {
        name: 'Meta AI (Llama)',
        url: 'https://api.llama.com',
        keyHint: 'Your Meta Llama API key',
        icon: '/icons/meta.svg',
        category: 'foundation',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Meta Llama API uses OpenAI-compatible format.'
        ]
    },

    // ── API & Infrastructure Providers ─────────
    openrouter: {
        name: 'OpenRouter',
        url: 'https://openrouter.ai/api',
        keyHint: 'Starts with sk-or-...',
        icon: '/icons/openrouter.svg',
        category: 'infra',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'OpenRouter is a unified gateway to 100+ models.',
            'Enter your OpenRouter API key. Save and go.'
        ]
    },
    groq: {
        name: 'Groq',
        url: 'https://api.groq.com/openai',
        keyHint: 'Your Groq API key',
        icon: '/icons/groq.svg',
        category: 'infra',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Groq uses OpenAI-compatible format. Enter your API key.'
        ]
    },
    together: {
        name: 'Together AI',
        url: 'https://api.together.xyz',
        keyHint: 'Your Together API key',
        icon: '/icons/together.svg',
        category: 'infra',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Together AI uses OpenAI-compatible format. Enter your API key.'
        ]
    },
    fireworks: {
        name: 'Fireworks AI',
        url: 'https://api.fireworks.ai/inference',
        keyHint: 'Your Fireworks API key',
        icon: '/icons/fireworks.svg',
        category: 'infra',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Fireworks uses OpenAI-compatible format. Enter your API key.'
        ]
    },
    huggingface: {
        name: 'Hugging Face',
        url: 'https://api-inference.huggingface.co',
        keyHint: 'Starts with hf_...',
        icon: '/icons/huggingface.svg',
        category: 'infra',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Enter your Hugging Face API token.'
        ]
    },
    deepinfra: {
        name: 'DeepInfra',
        url: 'https://api.deepinfra.com/v1/openai',
        keyHint: 'Your DeepInfra API key',
        icon: '/icons/deepinfra.svg',
        category: 'infra',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'DeepInfra uses OpenAI-compatible format. Enter your API key.'
        ]
    },
    siliconflow: {
        name: 'SiliconFlow',
        url: 'https://api.siliconflow.cn',
        keyHint: 'Your SiliconFlow API key',
        icon: '/icons/siliconflow.svg',
        category: 'infra',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'SiliconFlow uses OpenAI-compatible format.'
        ]
    },
    azure: {
        name: 'Azure OpenAI',
        url: '',
        keyHint: 'Your Azure API key',
        icon: '/icons/azure.svg',
        category: 'infra',
        needsCustomUrl: true,
        instructions: [
            'Set the API base URL to your Azure endpoint: <code>https://YOUR-RESOURCE.openai.azure.com</code>',
            'Then point your AI tool at <code id="instrUrl1"></code>'
        ]
    },
    bedrock: {
        name: 'AWS Bedrock',
        url: '',
        keyHint: 'Your AWS credentials',
        icon: '/icons/bedrock.svg',
        category: 'infra',
        needsCustomUrl: true,
        instructions: [
            'Enter your Bedrock endpoint URL',
            'Point your tool at <code id="instrUrl1"></code>'
        ]
    },

    // ── Specialized & Emerging ─────────────────
    perplexity: {
        name: 'Perplexity',
        url: 'https://api.perplexity.ai',
        keyHint: 'Your Perplexity API key',
        icon: '/icons/perplexity.svg',
        category: 'specialized',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Uses OpenAI-compatible format. Enter your API key.'
        ]
    },
    cerebras: {
        name: 'Cerebras',
        url: 'https://api.cerebras.ai',
        keyHint: 'Your Cerebras API key',
        icon: '/icons/cerebras.svg',
        category: 'specialized',
        instructions: [
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Uses OpenAI-compatible format. Enter your API key.'
        ]
    },

    // ── Local & Self-Hosted ────────────────────
    ollama: {
        name: 'Ollama',
        url: 'http://localhost:11434',
        keyHint: 'No key needed for local Ollama',
        icon: '/icons/ollama.svg',
        category: 'local',
        instructions: [
            'Make sure Ollama is running locally.',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed for local Ollama.'
        ]
    },
    lmstudio: {
        name: 'LM Studio',
        url: 'http://localhost:1234/v1',
        keyHint: 'No key needed for local LM Studio',
        icon: '/icons/lmstudio.svg',
        category: 'local',
        instructions: [
            'Start LM Studio and load a model.',
            'Go to the Local Server tab and click "Start Server".',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. LM Studio uses OpenAI-compatible format.'
        ]
    },
    llamacpp: {
        name: 'Llama.cpp',
        url: 'http://localhost:8080',
        keyHint: 'No key needed for llama-server',
        icon: '/icons/llamacpp.svg',
        category: 'local',
        instructions: [
            'Start llama-server with your model: llama-server -m model.gguf',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. Default port is 8080.'
        ]
    },
    localai: {
        name: 'LocalAI',
        url: 'http://localhost:8080/v1',
        keyHint: 'No key needed for LocalAI',
        icon: '/icons/localai.svg',
        category: 'local',
        instructions: [
            'Start LocalAI via Docker or binary.',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. OpenAI-compatible by default.'
        ]
    },
    vllm: {
        name: 'vLLM',
        url: 'http://localhost:8000/v1',
        keyHint: 'No key needed for local vLLM',
        icon: '/icons/vllm.svg',
        category: 'local',
        instructions: [
            'Start vLLM: python -m vllm.entrypoints.openai.api_server --model your-model',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. Fully OpenAI-compatible.'
        ]
    },
    oobabooga: {
        name: 'Text Gen WebUI',
        url: 'http://localhost:5000/v1',
        keyHint: 'No key needed for Oobabooga',
        icon: '/icons/oobabooga.svg',
        category: 'local',
        instructions: [
            'Start text-generation-webui with the --api flag.',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. Uses OpenAI-compatible API on port 5000.'
        ]
    },
    gpt4all: {
        name: 'GPT4All',
        url: 'http://localhost:4891/v1',
        keyHint: 'No key needed for GPT4All',
        icon: '/icons/gpt4all.svg',
        category: 'local',
        instructions: [
            'Open GPT4All desktop app.',
            'Go to Settings > Application > Enable Local Server.',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. Default port is 4891.'
        ]
    },
    jan: {
        name: 'Jan',
        url: 'http://localhost:1337/v1',
        keyHint: 'No key needed for Jan',
        icon: '/icons/jan.svg',
        category: 'local',
        instructions: [
            'Open Jan and start the local API server.',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. OpenAI-compatible on port 1337.'
        ]
    },
    koboldcpp: {
        name: 'Kobold.cpp',
        url: 'http://localhost:5001',
        keyHint: 'No key needed for KoboldCpp',
        icon: '/icons/koboldcpp.svg',
        category: 'local',
        instructions: [
            'Start KoboldCpp with your model loaded.',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. Default port is 5001.'
        ]
    },
    anythingllm: {
        name: 'AnythingLLM',
        url: 'http://localhost:3001/api',
        keyHint: 'Your AnythingLLM workspace API key',
        icon: '/icons/anythingllm.svg',
        category: 'local',
        instructions: [
            'Start AnythingLLM (Docker or desktop app).',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'You may need a workspace API key from AnythingLLM settings.'
        ]
    },
    msty: {
        name: 'Msty',
        url: 'http://localhost:10101',
        keyHint: 'No key needed for Msty',
        icon: '/icons/msty.svg',
        category: 'local',
        instructions: [
            'Open Msty and ensure the local server is running.',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'No API key needed. Default port is 10101.'
        ]
    },
    openwebui: {
        name: 'Open WebUI',
        url: 'http://localhost:3000',
        keyHint: 'Your Open WebUI API key (from Settings)',
        icon: '/icons/openwebui.svg',
        category: 'local',
        instructions: [
            'Start Open WebUI (Docker recommended).',
            'Set the API base URL to <code id="instrUrl1"></code>',
            'Get your API key from Open WebUI Settings > Account.'
        ]
    },

    // ── Agent Frameworks ───────────────────────
    // These are AI agent platforms that call LLM providers.
    // iTaK Shield sits between the agent and its upstream provider.
    openclaw: {
        name: 'OpenClaw',
        url: '',
        keyHint: 'API key for the upstream provider your agent uses',
        icon: '/icons/openclaw.svg',
        category: 'agents',
        featured: true,
        needsCustomUrl: true,
        instructions: [
            'In your OpenClaw agent config, set the API base URL to <code id="instrUrl1"></code>',
            'OpenClaw will route through iTaK Shield to your upstream provider.',
            'Enter the API key for your upstream provider (OpenAI, Anthropic, etc.)'
        ]
    },
    agentzero: {
        name: 'Agent Zero',
        url: '',
        keyHint: 'API key for the upstream provider your agent uses',
        icon: '/icons/agentzero.svg',
        category: 'agents',
        featured: true,
        needsCustomUrl: true,
        instructions: [
            'In Agent Zero\'s settings.json, change the API URL to <code id="instrUrl1"></code>',
            'Agent Zero will route through iTaK Shield to your upstream provider.',
            'Enter the API key for your upstream provider.'
        ]
    },

    // ── Custom / Catch-All ─────────────────────
    custom: {
        name: 'Custom',
        url: '',
        keyHint: 'Your API key for this provider',
        icon: '/icons/custom.svg',
        category: 'custom',
        featured: true,
        needsCustomUrl: true,
        instructions: [
            'Enter any OpenAI-compatible API base URL',
            'Point your tool at <code id="instrUrl1"></code>',
            'Enter the API key for your provider.'
        ]
    },

    // ── iTaK Agent (Coming Soon) ───────────────
    itakagent: {
        name: 'iTaK Agent',
        url: '',
        keyHint: '',
        icon: '/icons/itakagent.svg',
        category: 'agents',
        featured: true,
        comingSoon: true,
        githubUrl: 'https://github.com/David2024patton',
        instructions: []
    }
};

// Category display metadata
const CATEGORIES = {
    foundation: { label: 'Foundation Models', desc: 'Direct from the model developers' },
    infra: { label: 'API Gateways', desc: 'Unified access to multiple models' },
    specialized: { label: 'Specialized', desc: 'Search, speed, and niche providers' },
    local: { label: 'Local / Self-Hosted', desc: 'Run models on your own hardware' },
    agents: { label: 'Agent Frameworks', desc: 'AI agents that call LLM providers' },
    custom: { label: 'Other', desc: 'Any OpenAI-compatible endpoint' }
};

// ─── State ───────────────────────────────────
var currentStep = 1;
var selectedProvider = null;
var proxyRunning = false;
var pollInterval = null;
var startTime = null;

// ─── Featured provider display order ─────────
var FEATURED_ORDER = ['custom', 'itakagent', 'openclaw', 'agentzero'];

// ─── Build Provider Grid on Load ─────────────

function buildProviderGrid() {
    var container = document.getElementById('providerContainer');
    if (!container) return;
    container.innerHTML = '';

    // ── Featured row (sticky at top) ──────────
    var featuredHeader = document.createElement('div');
    featuredHeader.className = 'provider-category featured-category';
    featuredHeader.innerHTML = '<span class="category-label">Featured</span>' +
        '<span class="category-desc">Quick access and partner frameworks</span>';
    container.appendChild(featuredHeader);

    var featuredGrid = document.createElement('div');
    featuredGrid.className = 'provider-grid featured-grid';
    container.appendChild(featuredGrid);

    FEATURED_ORDER.forEach(function (pid) {
        var p = PROVIDERS[pid];
        if (!p) return;
        var card = createProviderCard(pid, p);
        featuredGrid.appendChild(card);
    });

    // ── Category groups (non-featured) ────────
    var categoryOrder = ['foundation', 'infra', 'specialized', 'local'];

    categoryOrder.forEach(function (catKey) {
        var cat = CATEGORIES[catKey];
        var providersInCat = [];

        for (var pid in PROVIDERS) {
            if (PROVIDERS[pid].category === catKey && !PROVIDERS[pid].featured) {
                providersInCat.push({ id: pid, data: PROVIDERS[pid] });
            }
        }

        if (providersInCat.length === 0) return;

        // Category header
        var header = document.createElement('div');
        header.className = 'provider-category';
        header.innerHTML = '<span class="category-label">' + cat.label + '</span>' +
            '<span class="category-desc">' + cat.desc + '</span>';
        container.appendChild(header);

        // Provider cards grid
        var grid = document.createElement('div');
        grid.className = 'provider-grid';
        container.appendChild(grid);

        providersInCat.forEach(function (p) {
            var card = createProviderCard(p.id, p.data);
            grid.appendChild(card);
        });
    });
}

// ─── Create a single provider card ───────────

function createProviderCard(pid, data) {
    var card = document.createElement('div');
    card.className = 'provider-card';
    card.dataset.provider = pid;
    card.dataset.name = data.name.toLowerCase();

    var iconHtml = '<img src="' + data.icon + '" alt="' + data.name + '" class="provider-icon-img" draggable="false">';

    if (data.comingSoon) {
        card.classList.add('coming-soon');
        card.onclick = function () {
            window.open(data.githubUrl, '_blank');
        };
        card.innerHTML =
            '<div class="coming-soon-badge">COMING SOON</div>' +
            '<div class="provider-icon">' + iconHtml + '</div>' +
            '<div class="provider-name">' + data.name + '</div>';
    } else {
        card.onclick = function () { selectProvider(pid); };
        card.innerHTML =
            '<div class="provider-icon">' + iconHtml + '</div>' +
            '<div class="provider-name">' + data.name + '</div>';
    }

    return card;
}

// ─── Search / Filter Providers ───────────────

function filterProviders(query) {
    var q = query.toLowerCase().trim();
    var container = document.getElementById('providerContainer');
    var cards = container.querySelectorAll('.provider-card');
    var categories = container.querySelectorAll('.provider-category');
    var grids = container.querySelectorAll('.provider-grid');

    // Show everything if empty query
    if (!q) {
        cards.forEach(function (c) { c.style.display = ''; });
        categories.forEach(function (c) { c.style.display = ''; });
        grids.forEach(function (g) { g.style.display = ''; });
        return;
    }

    // Hide/show cards based on name match
    cards.forEach(function (card) {
        var name = card.dataset.name || '';
        card.style.display = name.indexOf(q) !== -1 ? '' : 'none';
    });

    // Hide category headers + grids if all their cards are hidden
    grids.forEach(function (grid) {
        var visibleCards = grid.querySelectorAll('.provider-card:not([style*="display: none"])');
        var prevSibling = grid.previousElementSibling;
        if (visibleCards.length === 0) {
            grid.style.display = 'none';
            if (prevSibling && prevSibling.classList.contains('provider-category')) {
                prevSibling.style.display = 'none';
            }
        } else {
            grid.style.display = '';
            if (prevSibling && prevSibling.classList.contains('provider-category')) {
                prevSibling.style.display = '';
            }
        }
    });
}

// ─── Wizard Navigation ──────────────────────

function goToStep(step) {
    // Validate before advancing
    if (step === 3 && !selectedProvider) return;

    // Populate review if going to step 4
    if (step === 4) populateReview();

    currentStep = step;

    // Hide all panels, show target
    document.querySelectorAll('.wizard-panel').forEach(function (p) { p.classList.remove('active'); });
    var panel = document.getElementById('step' + step);
    if (panel) panel.classList.add('active');

    // Update step indicators
    document.querySelectorAll('.step-dot').forEach(function (dot) {
        var s = parseInt(dot.dataset.step);
        dot.classList.remove('active', 'completed');
        if (s === step) dot.classList.add('active');
        else if (s < step) dot.classList.add('completed');
    });

    document.querySelectorAll('.step-line').forEach(function (line) {
        var l = parseInt(line.dataset.line);
        line.classList.toggle('completed', l < step);
    });
}

// ─── Provider Selection ──────────────────────

function selectProvider(provider) {
    selectedProvider = provider;
    var prov = PROVIDERS[provider];

    // Update UI
    document.querySelectorAll('.provider-card').forEach(function (c) { c.classList.remove('selected'); });
    var selected = document.querySelector('[data-provider="' + provider + '"]');
    if (selected) selected.classList.add('selected');

    // Show/hide custom URL field
    var customGroup = document.getElementById('customUrlGroup');
    if (prov.needsCustomUrl || provider === 'custom') {
        customGroup.classList.add('visible');
    } else {
        customGroup.classList.remove('visible');
    }

    // Update API key placeholder
    var keyInput = document.getElementById('apiKey');
    if (keyInput) keyInput.placeholder = prov.keyHint;

    // Enable next button
    document.getElementById('step2Next').disabled = false;
}

// ─── Review Step ─────────────────────────────

function populateReview() {
    var provider = PROVIDERS[selectedProvider];
    var targetUrl = getTargetUrl();
    var port = document.getElementById('proxyPort').value || '8080';
    var apiKey = document.getElementById('apiKey').value;
    var verbose = document.getElementById('verboseMode').checked;

    document.getElementById('reviewProvider').textContent = provider.name;
    document.getElementById('reviewTarget').textContent = targetUrl || 'Not set';
    document.getElementById('reviewProxy').textContent = 'http://127.0.0.1:' + port;
    document.getElementById('reviewKey').textContent = apiKey ? '\u2022\u2022\u2022\u2022' + apiKey.slice(-4) : 'Not set (passed via headers)';
    document.getElementById('reviewVerbose').textContent = verbose ? 'On' : 'Off';
}

// ─── Get Target URL ──────────────────────────

function getTargetUrl() {
    var prov = PROVIDERS[selectedProvider];
    if (prov.needsCustomUrl || selectedProvider === 'custom') {
        return document.getElementById('customUrl').value;
    }
    return prov.url;
}

// ─── Start Proxy ─────────────────────────────

function startProxy() {
    var btn = document.getElementById('startBtn');
    btn.disabled = true;
    btn.textContent = 'Starting...';

    var targetUrl = getTargetUrl();
    var port = parseInt(document.getElementById('proxyPort').value) || 8080;
    var verbose = document.getElementById('verboseMode').checked;

    fetch('/api/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            target: targetUrl,
            port: port,
            verbose: verbose
        })
    })
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.ok) {
                proxyRunning = true;
                startTime = Date.now();
                showDashboard(port, targetUrl);
                startPolling();
            } else {
                alert('Failed to start: ' + (data.error || 'Unknown error'));
                btn.disabled = false;
                btn.textContent = '\uD83D\uDE80 Start iTaK Shield';
            }
        })
        .catch(function (err) {
            alert('Failed to connect to backend: ' + err.message);
            btn.disabled = false;
            btn.textContent = '\uD83D\uDE80 Start iTaK Shield';
        });
}

// ─── Stop Proxy ──────────────────────────────

function stopProxy() {
    fetch('/api/stop', { method: 'POST' }).catch(function () { });
    proxyRunning = false;
    stopPolling();

    var banner = document.getElementById('statusBanner');
    var dot = document.getElementById('statusDot');
    var text = document.getElementById('statusText');
    banner.className = 'status-banner stopped';
    dot.className = 'status-dot stopped';
    text.textContent = 'iTaK Shield is stopped';
}

// ─── Reset Wizard ────────────────────────────

function resetWizard() {
    stopProxy();
    proxyRunning = false;
    selectedProvider = null;
    currentStep = 1;

    document.getElementById('apiKey').value = '';
    document.getElementById('proxyPort').value = '8080';
    document.getElementById('verboseMode').checked = true;
    document.getElementById('customUrl').value = '';
    document.getElementById('step2Next').disabled = true;
    document.querySelectorAll('.provider-card').forEach(function (c) { c.classList.remove('selected'); });
    document.getElementById('customUrlGroup').classList.remove('visible');

    var btn = document.getElementById('startBtn');
    btn.disabled = false;
    btn.textContent = '\uD83D\uDE80 Start iTaK Shield';

    document.getElementById('dashboard').classList.remove('active');
    document.getElementById('stepsIndicator').style.display = '';
    goToStep(1);

    document.querySelectorAll('.wizard-panel').forEach(function (p) { p.style.display = ''; });
}

// ─── Show Dashboard ──────────────────────────

function showDashboard(port, targetUrl) {
    document.querySelectorAll('.wizard-panel').forEach(function (p) {
        p.classList.remove('active');
        p.style.display = 'none';
    });
    document.getElementById('stepsIndicator').style.display = 'none';
    document.getElementById('dashboard').classList.add('active');

    var proxyAddr = 'http://127.0.0.1:' + port;
    document.getElementById('dashProxy').textContent = proxyAddr;
    document.getElementById('dashTarget').textContent = targetUrl;

    // Provider-specific instructions
    var provider = PROVIDERS[selectedProvider];
    document.getElementById('dashProviderName').textContent = provider.name;

    var instrBody = document.getElementById('dashInstructionsBody');
    var ol = document.createElement('ol');
    provider.instructions.forEach(function (step) {
        var li = document.createElement('li');
        li.innerHTML = step;
        ol.appendChild(li);
    });
    instrBody.innerHTML = '';
    instrBody.appendChild(ol);

    // Replace instruction URL placeholders
    document.querySelectorAll('#instrUrl1').forEach(function (el) {
        el.textContent = proxyAddr;
    });

    // Reset stats
    document.getElementById('statRequests').textContent = '0';
    document.getElementById('statRedacted').textContent = '0';
    document.getElementById('statUptime').textContent = '0s';

    document.getElementById('activityLog').innerHTML =
        '<div class="log-empty">No requests yet. Send a request through the proxy to see activity here.</div>';

    var banner = document.getElementById('statusBanner');
    var dot = document.getElementById('statusDot');
    var text = document.getElementById('statusText');
    banner.className = 'status-banner running';
    dot.className = 'status-dot running';
    text.textContent = 'iTaK Shield is running';
}

// ─── Stats Polling ───────────────────────────

function startPolling() {
    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(pollStatus, 2000);
    pollStatus();
}

function stopPolling() {
    if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
    }
}

function pollStatus() {
    if (!proxyRunning) return;

    fetch('/api/status')
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            document.getElementById('statRequests').textContent = data.requests || '0';
            document.getElementById('statRedacted').textContent = data.redacted || '0';

            if (startTime) {
                var elapsed = Math.floor((Date.now() - startTime) / 1000);
                document.getElementById('statUptime').textContent = formatUptime(elapsed);
            }

            if (data.recent_logs && data.recent_logs.length > 0) {
                var logContainer = document.getElementById('activityLog');
                logContainer.innerHTML = '';
                data.recent_logs.forEach(function (entry) {
                    var div = document.createElement('div');
                    div.className = 'log-entry';
                    div.innerHTML =
                        '<span class="log-time">' + entry.time + '</span>' +
                        '<span class="log-type">' + entry.type + '</span> ' +
                        '<span class="log-msg">' + entry.message + '</span>';
                    logContainer.appendChild(div);
                });
            }

            if (!data.running) {
                proxyRunning = false;
                stopPolling();
                var banner = document.getElementById('statusBanner');
                var dot = document.getElementById('statusDot');
                var text = document.getElementById('statusText');
                banner.className = 'status-banner stopped';
                dot.className = 'status-dot stopped';
                text.textContent = 'iTaK Shield has stopped';
            }
        })
        .catch(function () { });
}

function formatUptime(seconds) {
    if (seconds < 60) return seconds + 's';
    if (seconds < 3600) {
        var m = Math.floor(seconds / 60);
        var s = seconds % 60;
        return m + 'm ' + s + 's';
    }
    var h = Math.floor(seconds / 3600);
    var m = Math.floor((seconds % 3600) / 60);
    return h + 'h ' + m + 'm';
}

// ─── Copy Text Helper ────────────────────────

function copyText(elementId) {
    var el = document.getElementById(elementId);
    if (!el) return;

    var text = el.textContent;
    navigator.clipboard.writeText(text).then(function () {
        var btn = el.parentElement.querySelector('.copy-btn');
        if (btn) {
            var original = btn.textContent;
            btn.textContent = 'Copied!';
            btn.style.color = 'var(--success)';
            btn.style.borderColor = 'var(--success)';
            setTimeout(function () {
                btn.textContent = original;
                btn.style.color = '';
                btn.style.borderColor = '';
            }, 1500);
        }
    });
}

// ─── Auto-check if proxy is already running ──

function checkInitialStatus() {
    fetch('/api/status')
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.running) {
                proxyRunning = true;
                startTime = Date.now() - (data.uptime_seconds * 1000);

                selectedProvider = 'custom';
                for (var key in PROVIDERS) {
                    if (PROVIDERS[key].url === data.target) {
                        selectedProvider = key;
                        break;
                    }
                }

                showDashboard(data.port, data.target);
                startPolling();
            }
        })
        .catch(function () { });
}

// ─── Init ────────────────────────────────────

document.addEventListener('DOMContentLoaded', function () {
    buildProviderGrid();
    checkInitialStatus();
});
