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

    // ── Automation Platforms ───────────────────
    n8n: {
        name: 'n8n',
        url: '',
        keyHint: 'n8n webhook or API credentials',
        icon: '/icons/n8n.svg',
        category: 'automation',
        needsCustomUrl: true,
        instructions: [
            'Set the target URL to your n8n webhook or HTTP Request node endpoint.',
            'Route your automation traffic through <code id="instrUrl1"></code>'
        ]
    },
    make: {
        name: 'Make (Integromat)',
        url: 'https://hook.make.com',
        keyHint: 'Make webhook authentication (if any)',
        icon: '/icons/make.svg',
        category: 'automation',
        needsCustomUrl: true,
        instructions: [
            'Set the target URL to your Make.com Webhook URL.',
            'Send requests to <code id="instrUrl1"></code>'
        ]
    },
    zapier: {
        name: 'Zapier',
        url: 'https://hooks.zapier.com',
        keyHint: 'Zapier authentication (if any)',
        icon: '/icons/zapier.svg',
        category: 'automation',
        needsCustomUrl: true,
        instructions: [
            'Set the target URL to your Zapier Catch Hook URL.',
            'Send requests to <code id="instrUrl1"></code>'
        ]
    },
    activepieces: {
        name: 'Activepieces',
        url: '',
        keyHint: 'Activepieces API key or token',
        icon: '/icons/activepieces.svg',
        category: 'automation',
        needsCustomUrl: true,
        instructions: [
            'Set the target to your Activepieces instance URL.',
            'Point your integrations to <code id="instrUrl1"></code>'
        ]
    },
    nodered: {
        name: 'Node-RED',
        url: 'http://localhost:1880',
        keyHint: 'Node-RED HTTP In node auth',
        icon: '/icons/nodered.svg',
        category: 'automation',
        needsCustomUrl: true,
        instructions: [
            'Set the target to your Node-RED HTTP In endpoint.',
            'Route requests through <code id="instrUrl1"></code>'
        ]
    },
    pipedream: {
        name: 'Pipedream',
        url: 'https://eo.pipedream.net',
        keyHint: 'Pipedream credentials',
        icon: '/icons/pipedream.svg',
        category: 'automation',
        needsCustomUrl: true,
        instructions: [
            'Set the target URL to your Pipedream workflow endpoint.',
            'Send data to <code id="instrUrl1"></code>'
        ]
    },
    gumloop: {
        name: 'Gumloop',
        url: 'https://api.gumloop.com',
        keyHint: 'Gumloop API Key',
        icon: '/icons/gumloop.svg',
        category: 'automation',
        needsCustomUrl: true,
        instructions: [
            'Set the target URL to the Gumloop API endpoint.',
            'Send requests through <code id="instrUrl1"></code>'
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

const CATEGORIES = {
    foundation: { label: 'Foundation Models', desc: 'Direct from the model developers' },
    infra: { label: 'API Gateways', desc: 'Unified access to multiple models' },
    specialized: { label: 'Specialized', desc: 'Search, speed, and niche providers' },
    local: { label: 'Local / Self-Hosted', desc: 'Run models on your own hardware' },
    agents: { label: 'Agent Frameworks', desc: 'AI agents that call LLM providers' },
    automation: { label: 'Automation Platforms', desc: 'n8n, Make, Zapier, and workflow engines' },
    custom: { label: 'Other', desc: 'Any OpenAI-compatible endpoint' }
};

// ─── State ───────────────────────────────────
var currentStep = 1;
var selectedProvider = null;
var selectedMode = null;
var proxyRunning = false;
var pollInterval = null;
var startTime = null;
var activeTab = 'overview';
var defaultRandomPort = Math.floor(Math.random() * (65535 - 10000 + 1)) + 10000;

// ─── Featured provider display order ─────────
var FEATURED_ORDER = ['custom', 'itakagent', 'openclaw', 'agentzero'];

// ─── PWA ─────────────────────────────────────
var deferredPWAPrompt = null;

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
    var categoryOrder = ['foundation', 'infra', 'specialized', 'local', 'automation'];

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
    if (step === 3 && !selectedMode) return;
    if (step === 4 && !selectedProvider) return;

    // Populate review if going to step 5
    if (step === 5) populateReview();

    currentStep = step;

    // Hide all panels, show target
    document.querySelectorAll('.wiz-panel').forEach(function (p) { p.classList.remove('active'); });
    var panel = document.getElementById('step' + step);
    if (panel) panel.classList.add('active');

    // Update step indicators
    document.querySelectorAll('.wiz-dot').forEach(function (dot) {
        var s = parseInt(dot.dataset.step);
        dot.classList.remove('active', 'completed');
        if (s === step) dot.classList.add('active');
        else if (s < step) dot.classList.add('completed');
    });

    document.querySelectorAll('.wiz-line').forEach(function (line) {
        var l = parseInt(line.dataset.line);
        line.classList.toggle('completed', l < step);
    });
}

// ─── Mode Selection ──────────────────────────

function selectMode(mode) {
    selectedMode = mode;
    document.getElementById('modeIndividual').classList.toggle('selected', mode === 'individual');
    document.getElementById('modeCompany').classList.toggle('selected', mode === 'company');
    document.getElementById('step2Next').disabled = false;
}

function setMode(mode) {
    selectedMode = mode;
    localStorage.setItem('itak_mode', mode);
    applyMode(mode);
}

function applyMode(mode) {
    var shell = document.getElementById('appShell');
    if (!shell) return;
    shell.classList.remove('mode-individual', 'mode-company');
    shell.classList.add('mode-' + mode);

    // Update settings toggle buttons
    var indBtn = document.getElementById('settingsModeIndividual');
    var comBtn = document.getElementById('settingsModeCompany');
    if (indBtn) indBtn.classList.toggle('active', mode === 'individual');
    if (comBtn) comBtn.classList.toggle('active', mode === 'company');

    // If currently on a company-only tab in individual mode, switch to overview
    if (mode === 'individual' && (activeTab === 'analytics' || activeTab === 'team')) {
        switchTab('overview');
    }
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
    document.getElementById('step3Next').disabled = false;
}

// ─── Review Step ─────────────────────────────

function populateReview() {
    var provider = PROVIDERS[selectedProvider];
    var targetUrl = getTargetUrl();
    var port = document.getElementById('proxyPort').value || defaultRandomPort;
    var apiKey = document.getElementById('apiKey').value;
    var verbose = document.getElementById('verboseMode').checked;

    document.getElementById('reviewMode').textContent = selectedMode === 'company' ? 'Company / Team' : 'Individual';
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

// ─── Port Availability Check ─────────────────

function checkPortAvailability() {
    var port = parseInt(document.getElementById('proxyPort').value);
    var statusEl = document.getElementById('portStatus');
    if (!port || port < 1024 || port > 65535) {
        if (statusEl) {
            statusEl.style.display = '';
            statusEl.style.color = 'var(--warning)';
            statusEl.textContent = 'Port must be between 1024 and 65535.';
        }
        return;
    }
    if (statusEl) {
        statusEl.style.display = '';
        statusEl.style.color = 'var(--text-muted)';
        statusEl.textContent = 'Checking port ' + port + '...';
    }
    fetch('/api/port/check?port=' + port)
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (!statusEl) return;
            statusEl.style.display = '';
            if (data.ok) {
                statusEl.style.color = 'var(--success)';
                statusEl.textContent = 'Port ' + port + ' is available.';
            } else {
                statusEl.style.color = 'var(--danger)';
                statusEl.textContent = 'Port ' + port + ' is in use. Click to randomize: ';
                var link = document.createElement('a');
                link.href = '#';
                link.textContent = 'pick another';
                link.style.color = 'var(--accent)';
                link.onclick = function (e) {
                    e.preventDefault();
                    defaultRandomPort = Math.floor(Math.random() * (65535 - 10000 + 1)) + 10000;
                    document.getElementById('proxyPort').value = defaultRandomPort;
                    checkPortAvailability();
                };
                statusEl.appendChild(link);
            }
        })
        .catch(function () {
            if (statusEl) {
                statusEl.style.display = '';
                statusEl.style.color = 'var(--text-muted)';
                statusEl.textContent = '';
            }
        });
}

function startProxy() {
    var btn = document.getElementById('startBtn');
    btn.disabled = true;
    btn.textContent = 'Starting...';

    var targetUrl = getTargetUrl();
    var port = parseInt(document.getElementById('proxyPort').value) || defaultRandomPort;
    var verbose = document.getElementById('verboseMode').checked;

    // Persist usage mode
    if (selectedMode) {
        localStorage.setItem('itak_mode', selectedMode);
    }

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
                var errMsg = data.error || 'Unknown error';
                // If port is in use, offer to randomize.
                if (errMsg.indexOf('already in use') >= 0) {
                    defaultRandomPort = Math.floor(Math.random() * (65535 - 10000 + 1)) + 10000;
                    alert('Port ' + port + ' is in use. A new random port (' + defaultRandomPort + ') has been selected. Click Start again.');
                    document.getElementById('proxyPort').value = defaultRandomPort;
                    checkPortAvailability();
                } else {
                    alert('Failed to start: ' + errMsg);
                }
                btn.disabled = false;
                btn.textContent = 'Start iTaK Shield';
            }
        })
        .catch(function (err) {
            alert('Failed to connect to backend: ' + err.message);
            btn.disabled = false;
            btn.textContent = 'Start iTaK Shield';
        });
}

// ─── Stop Proxy ──────────────────────────────

function stopProxy() {
    fetch('/api/stop', { method: 'POST' }).catch(function () { });
    proxyRunning = false;
    stopPolling();

    var dot = document.getElementById('statusDot');
    var text = document.getElementById('statusText');
    if (dot) dot.className = 'status-dot-sm stopped';
    if (text) text.textContent = 'iTaK Shield is stopped';
}

// ─── Reset Wizard ────────────────────────────

function openWizard() {
    document.getElementById('wizardOverlay').classList.remove('hidden');
    document.getElementById('appShell').style.display = 'none';

    // Reset wizard state
    selectedProvider = null;
    selectedMode = null;
    currentStep = 1;

    document.getElementById('apiKey').value = '';
    document.getElementById('proxyPort').value = defaultRandomPort;
    document.getElementById('verboseMode').checked = true;
    document.getElementById('customUrl').value = '';
    document.getElementById('step2Next').disabled = true;
    document.getElementById('step3Next').disabled = true;
    document.querySelectorAll('.provider-card').forEach(function (c) { c.classList.remove('selected'); });
    document.getElementById('customUrlGroup').classList.remove('visible');
    document.getElementById('modeIndividual').classList.remove('selected');
    document.getElementById('modeCompany').classList.remove('selected');

    var btn = document.getElementById('startBtn');
    btn.disabled = false;
    btn.textContent = 'Start iTaK Shield';

    goToStep(1);
}

// ─── Show Dashboard (transition to app shell) ──

function showDashboard(port, targetUrl) {
    // Hide wizard overlay, show app shell
    document.getElementById('wizardOverlay').classList.add('hidden');
    document.getElementById('appShell').style.display = '';

    var proxyAddr = 'http://127.0.0.1:' + port;
    document.getElementById('dashProxy').textContent = proxyAddr;
    document.getElementById('dashTarget').textContent = targetUrl;

    // Provider-specific instructions
    var provider = PROVIDERS[selectedProvider];
    document.getElementById('dashProviderName').textContent = provider.name;

    var instrBody = document.getElementById('dashInstructionsBody');
    var ol = document.createElement('ol');
    ol.className = 'instructions-list';
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

    // Update status bar
    var dot = document.getElementById('statusDot');
    var text = document.getElementById('statusText');
    if (dot) dot.className = 'status-dot-sm running';
    if (text) text.textContent = 'iTaK Shield is running';

    // Apply usage mode
    var mode = selectedMode || localStorage.getItem('itak_mode') || 'individual';
    applyMode(mode);

    // Populate settings panel
    document.getElementById('settingsProvider').textContent = provider.name;
    document.getElementById('settingsTarget').textContent = targetUrl;
    document.getElementById('settingsProxy').textContent = proxyAddr;

    // Switch to overview tab
    switchTab('overview');

    // Load team users if in company mode
    if (mode === 'company') {
        loadTeamUsers();
    }
}

// ─── Stats Polling ───────────────────────────

function startPolling() {
    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(function () {
        pollStatus();
        pollAnalytics();
    }, 2000);
    pollStatus();
    pollAnalytics();
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
                var dot = document.getElementById('statusDot');
                var text = document.getElementById('statusText');
                if (dot) dot.className = 'status-dot-sm stopped';
                if (text) text.textContent = 'iTaK Shield has stopped';
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

// ─── Enterprise Analytics (removed analyticsSection hide/show) ──

function pollAnalytics() {
    if (!proxyRunning) return;

    fetch('/api/analytics')
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (!data.active) return;
            updateAnalytics(data);
        })
        .catch(function () { });
}

function updateAnalytics(data) {
    var features = data.features || {};
    var badgeContainer = document.getElementById('featureBadges');
    var badgeNames = {
        auth: { label: 'Auth', icon: '🔑' },
        cache: { label: 'Cache', icon: '💾' },
        retry: { label: 'Retry', icon: '🔄' },
        spend: { label: 'Spend', icon: '💰' },
        dlp: { label: 'DLP', icon: '🛡️' }
    };

    var badgeHtml = '';
    for (var key in badgeNames) {
        var active = features[key] ? 'active' : '';
        badgeHtml += '<span class="feature-badge ' + active + '">' +
            badgeNames[key].icon + ' ' + badgeNames[key].label + '</span>';
    }
    badgeContainer.innerHTML = badgeHtml;

    // Cache stats
    var cacheCard = document.getElementById('cacheCard');
    if (data.cache && features.cache) {
        cacheCard.style.display = '';
        var total = (data.cache.hits || 0) + (data.cache.misses || 0);
        var hitRate = total > 0 ? Math.round((data.cache.hits / total) * 100) : 0;
        document.getElementById('cacheHitRate').textContent = hitRate + '%';
        document.getElementById('cacheBar').style.width = hitRate + '%';
        document.getElementById('cacheHits').textContent = data.cache.hits || 0;
        document.getElementById('cacheMisses').textContent = data.cache.misses || 0;
        document.getElementById('cacheEntries').textContent = data.cache.entries || 0;
        document.getElementById('cacheMax').textContent = data.cache.max_entries || 0;
    } else {
        cacheCard.style.display = 'none';
    }

    // Spend stats
    var spendCard = document.getElementById('spendCard');
    if (data.spend && features.spend) {
        spendCard.style.display = '';
        document.getElementById('spendTotal').textContent = '$' + (data.spend.total_usd || 0).toFixed(4);
        document.getElementById('spendInput').textContent = formatTokens(data.spend.total_input || 0);
        document.getElementById('spendOutput').textContent = formatTokens(data.spend.total_output || 0);

        var byUserContainer = document.getElementById('spendByUser');
        if (data.spend.by_user && Object.keys(data.spend.by_user).length > 0) {
            var rows = '<div class="spend-table-header"><span>User</span><span>Tokens</span><span>Cost</span></div>';
            for (var user in data.spend.by_user) {
                var s = data.spend.by_user[user];
                var totalTokens = (s.input_tokens || 0) + (s.output_tokens || 0);
                rows += '<div class="spend-table-row">' +
                    '<span>' + user + '</span>' +
                    '<span>' + formatTokens(totalTokens) + '</span>' +
                    '<span>$' + (s.estimated_usd || 0).toFixed(4) + '</span>' +
                    '</div>';
            }
            byUserContainer.innerHTML = rows;
        } else {
            byUserContainer.innerHTML = '';
        }
    } else {
        spendCard.style.display = 'none';
    }

    // User activity
    var usersCard = document.getElementById('usersCard');
    if (data.auth_users && features.auth) {
        usersCard.style.display = '';
        var table = document.getElementById('userActivityTable');
        var sorted = Object.entries(data.auth_users).sort(function (a, b) { return b[1] - a[1]; });
        var html = '<div class="user-table-header"><span>User</span><span>Requests</span></div>';
        sorted.forEach(function (entry) {
            html += '<div class="user-table-row"><span>' + entry[0] + '</span><span>' + entry[1] + '</span></div>';
        });
        table.innerHTML = html;
    } else {
        usersCard.style.display = 'none';
    }
}

function formatTokens(n) {
    if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
    if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
    return n.toString();
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

// ─── Team Management ─────────────────────────

function loadTeamUsers() {
    fetch('/api/users')
        .then(function (resp) { return resp.json(); })
        .then(function (users) {
            var container = document.getElementById('teamUserList');
            if (!users || users.length === 0) {
                container.innerHTML = '<div class="log-empty">No users yet. Add a user above to get started.</div>';
                return;
            }
            container.innerHTML = '';
            users.forEach(function (user) {
                container.appendChild(renderUserCard(user));
            });
        })
        .catch(function () { });
}

function createUser() {
    var name = document.getElementById('newUserName').value.trim();
    var email = document.getElementById('newUserEmail').value.trim();
    var group = document.getElementById('newUserGroup').value.trim() || 'default';
    var rateLimit = parseInt(document.getElementById('newUserRate').value) || 0;

    if (!name) {
        alert('User name is required.');
        return;
    }

    fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name, email: email, group: group, rate_limit: rateLimit })
    })
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.ok) {
                // Clear form
                document.getElementById('newUserName').value = '';
                document.getElementById('newUserEmail').value = '';
                document.getElementById('newUserGroup').value = 'default';
                document.getElementById('newUserRate').value = '0';
                loadTeamUsers();
            } else {
                alert('Error: ' + (data.error || 'Failed to create user'));
            }
        })
        .catch(function (err) { alert('Network error: ' + err.message); });
}

function deleteUser(userId) {
    if (!confirm('Delete this user and all their tokens?')) return;

    fetch('/api/users/' + userId, { method: 'DELETE' })
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.ok) {
                loadTeamUsers();
            } else {
                alert('Error: ' + (data.error || 'Failed to delete user'));
            }
        })
        .catch(function (err) { alert('Network error: ' + err.message); });
}

function generateToken(userId) {
    var labelInput = document.getElementById('tokenLabel_' + userId);
    var expiresInput = document.getElementById('tokenExpires_' + userId);
    var label = labelInput ? labelInput.value.trim() : 'api-key';
    var expiresIn = expiresInput ? parseInt(expiresInput.value) : 0;

    if (!label) label = 'api-key';

    var body = { user_id: userId, label: label };
    if (expiresIn > 0) body.expires_in = expiresIn;

    fetch('/api/tokens', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
    })
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.ok && data.token) {
                // Show the generated token key in a special reveal box
                var reveal = document.getElementById('tokenReveal_' + userId);
                if (reveal) {
                    reveal.style.display = '';
                    reveal.querySelector('.token-key-value').textContent = data.token.key;
                }
                // Clear the form
                if (labelInput) labelInput.value = '';
                if (expiresInput) expiresInput.value = '';
                // Refresh user list to show the new token
                loadTeamUsers();
            } else {
                alert('Error: ' + (data.error || 'Failed to generate token'));
            }
        })
        .catch(function (err) { alert('Network error: ' + err.message); });
}

function revokeToken(userId, tokenKey) {
    if (!confirm('Revoke this token? It will immediately stop working.')) return;

    fetch('/api/tokens/revoke', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: userId, token_key: tokenKey })
    })
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.ok) {
                loadTeamUsers();
            } else {
                alert('Error: ' + (data.error || 'Failed to revoke token'));
            }
        })
        .catch(function (err) { alert('Network error: ' + err.message); });
}

function renderUserCard(user) {
    var card = document.createElement('div');
    card.className = 'team-user-card';

    // Header
    var header = document.createElement('div');
    header.className = 'team-user-header';

    var info = document.createElement('div');
    info.className = 'team-user-info';
    info.innerHTML =
        '<span class="team-user-name">' + user.name + '</span>' +
        (user.email ? '<span class="team-user-email">' + user.email + '</span>' : '') +
        '<span class="team-user-meta">' + user.group + (user.rate_limit > 0 ? ' | ' + user.rate_limit + ' req/min' : ' | unlimited') + '</span>';

    var actions = document.createElement('div');
    actions.className = 'team-user-actions';
    actions.innerHTML = '<button class="btn-icon btn-icon-danger" onclick="deleteUser(\'' + user.id + '\')" title="Delete user">&#x2716;</button>';

    header.appendChild(info);
    header.appendChild(actions);
    card.appendChild(header);

    // Existing tokens
    var tokensDiv = document.createElement('div');
    tokensDiv.className = 'team-tokens';

    if (user.tokens && user.tokens.length > 0) {
        user.tokens.forEach(function (token) {
            if (token.revoked) return;
            var row = document.createElement('div');
            row.className = 'team-token-row';

            var keyPreview = token.key.substring(0, 8) + '...' + token.key.substring(token.key.length - 4);
            var expiry = token.expires_at ? new Date(token.expires_at).toLocaleDateString() : 'Never';

            row.innerHTML =
                '<div class="team-token-info">' +
                '<span class="team-token-label">' + token.label + '</span>' +
                '<code class="team-token-preview">' + keyPreview + '</code>' +
                '<span class="team-token-expiry">Expires: ' + expiry + '</span>' +
                '</div>' +
                '<button class="btn-icon btn-icon-warning" onclick="revokeToken(\'' + user.id + '\', \'' + token.key + '\')" title="Revoke">Revoke</button>';

            tokensDiv.appendChild(row);
        });
    }

    card.appendChild(tokensDiv);

    // Generate token form
    var genDiv = document.createElement('div');
    genDiv.className = 'team-gen-token';
    genDiv.innerHTML =
        '<div class="team-gen-row">' +
        '<input type="text" class="form-input form-input-sm" id="tokenLabel_' + user.id + '" placeholder="Label (e.g. prod-key)">' +
        '<input type="number" class="form-input form-input-sm" id="tokenExpires_' + user.id + '" placeholder="Expires in (hours)" min="0">' +
        '<button class="btn btn-sm btn-primary" onclick="generateToken(\'' + user.id + '\')">Generate Token</button>' +
        '</div>' +
        '<div class="team-token-reveal" id="tokenReveal_' + user.id + '" style="display: none;">' +
        '<span class="team-token-reveal-label">New token (copy now, shown once):</span>' +
        '<div class="team-token-reveal-key">' +
        '<code class="token-key-value"></code>' +
        '<button class="copy-btn" onclick="copyTokenText(this)">Copy</button>' +
        '</div>' +
        '</div>';

    card.appendChild(genDiv);

    return card;
}

function copyTokenText(btn) {
    var code = btn.previousElementSibling;
    if (!code) return;
    navigator.clipboard.writeText(code.textContent).then(function () {
        var orig = btn.textContent;
        btn.textContent = 'Copied!';
        btn.style.color = 'var(--success)';
        btn.style.borderColor = 'var(--success)';
        setTimeout(function () {
            btn.textContent = orig;
            btn.style.color = '';
            btn.style.borderColor = '';
        }, 2000);
    });
}

// ─── Help Tooltips ───────────────────────────

var activeTooltip = null;

function initHelpIcons() {
    var icons = document.querySelectorAll('.help-icon');
    icons.forEach(function (icon) {
        icon.addEventListener('mouseenter', function () {
            showHelpTooltip(icon);
        });
        icon.addEventListener('mouseleave', function () {
            hideHelpTooltip();
        });
        icon.addEventListener('click', function (e) {
            e.stopPropagation();
            if (activeTooltip && activeTooltip.parentElement === icon) {
                hideHelpTooltip();
            } else {
                showHelpTooltip(icon);
            }
        });
    });

    document.addEventListener('click', function () {
        hideHelpTooltip();
    });
}

function showHelpTooltip(icon) {
    hideHelpTooltip();
    var text = icon.getAttribute('data-help');
    if (!text) return;

    var tooltip = document.createElement('div');
    tooltip.className = 'help-tooltip';
    tooltip.textContent = text;
    activeTooltip = tooltip;

    document.body.appendChild(tooltip);

    // Position below the icon
    var rect = icon.getBoundingClientRect();
    var tW = tooltip.offsetWidth;
    var tH = tooltip.offsetHeight;

    var left = rect.left + (rect.width / 2) - (tW / 2);
    var top = rect.bottom + 6;

    // Keep within viewport
    if (left < 8) left = 8;
    if (left + tW > window.innerWidth - 8) left = window.innerWidth - tW - 8;
    if (top + tH > window.innerHeight - 8) {
        top = rect.top - tH - 6;
    }

    tooltip.style.left = left + 'px';
    tooltip.style.top = top + 'px';
}

function hideHelpTooltip() {
    if (activeTooltip) {
        activeTooltip.remove();
        activeTooltip = null;
    }
}

// ─── Tunnel / Remote Access ──────────────────

var tunnelConnected = false;
var tunnelStartTime = null;
var tunnelPollInterval = null;

function connectTunnel() {
    var addr = document.getElementById('tunnelRelayAddr').value.trim();
    if (!addr) {
        alert('Enter the relay server address (e.g. your-vps.com:9443)');
        return;
    }

    var btn = document.getElementById('tunnelConnectBtn');
    btn.disabled = true;
    btn.textContent = 'Connecting...';

    // Update status to connecting
    updateTunnelUI('connecting');

    fetch('/api/tunnel/connect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ relay: addr })
    })
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.ok) {
                tunnelConnected = true;
                tunnelStartTime = Date.now();
                showTunnelConnected(data.public_url || ('http://' + addr), addr);
                startTunnelPolling();
            } else {
                alert('Failed to connect: ' + (data.error || 'Unknown error'));
                updateTunnelUI('disconnected');
                btn.disabled = false;
                btn.textContent = 'Connect';
            }
        })
        .catch(function (err) {
            alert('Network error: ' + err.message);
            updateTunnelUI('disconnected');
            btn.disabled = false;
            btn.textContent = 'Connect';
        });
}

function disconnectTunnel() {
    fetch('/api/tunnel/disconnect', { method: 'POST' }).catch(function () { });
    tunnelConnected = false;
    stopTunnelPolling();
    showTunnelDisconnected();
}

function showTunnelConnected(publicUrl, relay) {
    document.getElementById('tunnelDisconnected').style.display = 'none';
    document.getElementById('tunnelConnected').style.display = '';
    document.getElementById('tunnelPublicUrl').textContent = publicUrl;
    document.getElementById('tunnelRelayDisplay').textContent = relay;
    updateTunnelUI('connected');
}

function showTunnelDisconnected() {
    document.getElementById('tunnelConnected').style.display = 'none';
    document.getElementById('tunnelDisconnected').style.display = '';
    var btn = document.getElementById('tunnelConnectBtn');
    btn.disabled = false;
    btn.textContent = 'Connect';
    updateTunnelUI('disconnected');
}

function updateTunnelUI(state) {
    var dot = document.getElementById('tunnelDot');
    var label = document.getElementById('tunnelStatusLabel');

    dot.className = 'tunnel-status-dot ' + state;
    if (state === 'connected') {
        label.textContent = 'Connected';
        label.style.color = 'var(--success)';
    } else if (state === 'connecting') {
        label.textContent = 'Connecting...';
        label.style.color = 'var(--warning)';
    } else {
        label.textContent = 'Disconnected';
        label.style.color = '';
    }
}

function startTunnelPolling() {
    if (tunnelPollInterval) clearInterval(tunnelPollInterval);
    tunnelPollInterval = setInterval(function () {
        pollTunnelStatus();
    }, 3000);
}

function stopTunnelPolling() {
    if (tunnelPollInterval) {
        clearInterval(tunnelPollInterval);
        tunnelPollInterval = null;
    }
}

function pollTunnelStatus() {
    if (!tunnelConnected) return;

    // Update uptime
    if (tunnelStartTime) {
        var elapsed = Math.floor((Date.now() - tunnelStartTime) / 1000);
        document.getElementById('tunnelUptime').textContent = formatUptime(elapsed);
    }

    fetch('/api/tunnel/status')
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (!data.connected) {
                tunnelConnected = false;
                stopTunnelPolling();
                showTunnelDisconnected();
            }
        })
        .catch(function () { });
}

// ─── Copy Inline Code Helper ─────────────────

function copyInline(btn) {
    var code = btn.previousElementSibling;
    if (!code) return;
    navigator.clipboard.writeText(code.textContent).then(function () {
        var orig = btn.textContent;
        btn.textContent = 'Copied!';
        btn.style.color = 'var(--success)';
        btn.style.borderColor = 'var(--success)';
        setTimeout(function () {
            btn.textContent = orig;
            btn.style.color = '';
            btn.style.borderColor = '';
        }, 1500);
    });
}

// ─── Tab Switching ───────────────────────────

function switchTab(tab) {
    activeTab = tab;

    // Update sidebar nav
    document.querySelectorAll('.nav-item').forEach(function (item) {
        item.classList.toggle('active', item.dataset.tab === tab);
    });

    // Update panels
    document.querySelectorAll('.tab-panel').forEach(function (panel) {
        panel.classList.toggle('active', panel.dataset.tab === tab);
    });

    // Close mobile drawer after selection
    var sidebar = document.getElementById('sidebar');
    var backdrop = document.getElementById('sidebarBackdrop');
    if (sidebar) sidebar.classList.remove('drawer-open');
    if (backdrop) backdrop.classList.remove('visible');
}

// ─── Mobile Drawer Toggle ───────────────────

function toggleMobileDrawer() {
    var sidebar = document.getElementById('sidebar');
    var backdrop = document.getElementById('sidebarBackdrop');
    if (!sidebar) return;

    var isOpen = sidebar.classList.toggle('drawer-open');
    if (backdrop) {
        if (isOpen) {
            backdrop.classList.add('visible');
        } else {
            backdrop.classList.remove('visible');
        }
    }
}

// ─── Sidebar Toggle ──────────────────────────

function toggleSidebar() {
    var shell = document.getElementById('appShell');
    if (!shell) return;
    var collapsed = shell.classList.toggle('sidebar-collapsed');
    localStorage.setItem('itak_sidebar', collapsed ? 'collapsed' : 'expanded');
}

// ─── Sidebar Resize ──────────────────────────

function initSidebarResize() {
    var handle = document.getElementById('sidebarResize');
    var sidebar = document.getElementById('sidebar');
    if (!handle || !sidebar) return;

    var startX, startW;

    handle.addEventListener('mousedown', function (e) {
        e.preventDefault();
        startX = e.clientX;
        startW = sidebar.offsetWidth;
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';

        function onMove(ev) {
            var newW = Math.max(180, Math.min(360, startW + (ev.clientX - startX)));
            sidebar.style.width = newW + 'px';
        }

        function onUp() {
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
            document.removeEventListener('mousemove', onMove);
            document.removeEventListener('mouseup', onUp);
            localStorage.setItem('itak_sidebar_width', sidebar.style.width);
        }

        document.addEventListener('mousemove', onMove);
        document.addEventListener('mouseup', onUp);
    });

    // Restore saved width
    var savedW = localStorage.getItem('itak_sidebar_width');
    if (savedW) sidebar.style.width = savedW;
}



// ─── PWA Install ─────────────────────────────

function pwaInstall() {
    if (deferredPWAPrompt) {
        deferredPWAPrompt.prompt();
        deferredPWAPrompt.userChoice.then(function () {
            deferredPWAPrompt = null;
            document.getElementById('pwaBanner').style.display = 'none';
        });
    }
}

function pwaDismiss() {
    document.getElementById('pwaBanner').style.display = 'none';
    sessionStorage.setItem('itak_pwa_dismissed', '1');
}

// ─── Init ────────────────────────────────────

// ─── Theme Toggle (light/dark) ───────────────

function initThemeToggle() {
    var theme = localStorage.getItem('itak_theme') || 'dark';
    applyTheme(theme);
}

function toggleTheme() {
    var current = document.documentElement.getAttribute('data-theme') || 'dark';
    var next = current === 'dark' ? 'light' : 'dark';
    applyTheme(next);
    localStorage.setItem('itak_theme', next);
}

function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    var moon = document.getElementById('themeIconMoon');
    var sun = document.getElementById('themeIconSun');
    if (moon && sun) {
        moon.style.display = theme === 'dark' ? '' : 'none';
        sun.style.display = theme === 'dark' ? 'none' : '';
    }
    // Update the <meta name="theme-color"> for mobile browser chrome.
    var meta = document.querySelector('meta[name="theme-color"]');
    if (meta) {
        meta.setAttribute('content', theme === 'dark' ? '#0d1117' : '#ffffff');
    }
}

document.addEventListener('DOMContentLoaded', function () {
    buildProviderGrid();
    document.getElementById('proxyPort').value = defaultRandomPort;
    checkInitialStatus();
    initHelpIcons();
    initSidebarResize();
    initThemeToggle();

    // Restore sidebar collapsed state
    if (localStorage.getItem('itak_sidebar') === 'collapsed') {
        var shell = document.getElementById('appShell');
        if (shell) shell.classList.add('sidebar-collapsed');
    }

    // Check if tunnel is already active
    fetch('/api/tunnel/status')
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (data.connected) {
                tunnelConnected = true;
                tunnelStartTime = Date.now() - ((data.uptime_seconds || 0) * 1000);
                showTunnelConnected(data.public_url || '-', data.relay || '-');
                startTunnelPolling();
            }
        })
        .catch(function () { });

    // PWA install prompt
    window.addEventListener('beforeinstallprompt', function (e) {
        e.preventDefault();
        deferredPWAPrompt = e;
        if (!sessionStorage.getItem('itak_pwa_dismissed')) {
            document.getElementById('pwaBanner').style.display = '';
        }
    });

    // Register service worker
    if ('serviceWorker' in navigator) {
        navigator.serviceWorker.register('/sw.js').catch(function () { });
    }
});

// ─── Guard / Security ────────────────────────

function testGuard() {
    var text = document.getElementById('guardTestInput').value.trim();
    var source = document.getElementById('guardTestSource').value;

    if (!text) {
        alert('Please enter some text to scan.');
        return;
    }

    fetch('/api/guard/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: text, source: source })
    })
        .then(function (resp) { return resp.json(); })
        .then(function (data) {
            if (!data.ok) {
                alert('Error: ' + (data.error || 'Unknown'));
                return;
            }
            renderGuardResult(data);
        })
        .catch(function (err) {
            alert('Scan failed: ' + err.message);
        });
}

function renderGuardResult(data) {
    var container = document.getElementById('guardResult');
    var card = document.getElementById('guardResultCard');
    var icon = document.getElementById('guardResultIcon');
    var verdict = document.getElementById('guardResultVerdict');
    var severity = document.getElementById('guardResultSeverity');
    var action = document.getElementById('guardResultAction');
    var source = document.getElementById('guardResultSource');
    var time = document.getElementById('guardResultTime');
    var reasonsRow = document.getElementById('guardReasonsRow');
    var reasons = document.getElementById('guardResultReasons');

    container.style.display = '';

    // Clear classes
    card.className = 'guard-result-card';

    if (data.blocked) {
        card.classList.add('blocked');
        icon.textContent = '\u26D4';
        verdict.textContent = 'BLOCKED';
        verdict.style.color = 'var(--danger)';
    } else if (data.severity === 'MEDIUM' || data.severity === 'LOW') {
        card.classList.add('warn');
        icon.textContent = '\u26A0\uFE0F';
        verdict.textContent = 'FLAGGED';
        verdict.style.color = 'var(--warning)';
    } else {
        card.classList.add('safe');
        icon.textContent = '\u2705';
        verdict.textContent = 'SAFE';
        verdict.style.color = 'var(--success)';
    }

    // Severity badge
    severity.textContent = data.severity;
    severity.className = 'info-val guard-severity sev-' + data.severity.toLowerCase();

    action.textContent = data.action;
    source.textContent = data.source;
    time.textContent = (data.scan_ms || 0) + 'ms';

    // Reasons
    if (data.reasons && data.reasons.length > 0) {
        reasonsRow.style.display = '';
        reasons.textContent = data.reasons.join(', ');
    } else {
        reasonsRow.style.display = 'none';
    }
}

function setGuardSensitivity(level) {
    var btns = document.querySelectorAll('#guardSensitivity .mode-btn');
    btns.forEach(function (btn) { btn.classList.remove('active'); });

    var labels = { 2: 'Paranoid', 3: 'Default', 4: 'Relaxed' };
    btns.forEach(function (btn) {
        if (btn.textContent.trim() === labels[level]) {
            btn.classList.add('active');
        }
    });

    var descs = {
        2: 'Paranoid: Blocks MEDIUM severity and above. Maximum protection, may have false positives.',
        3: 'Default: Blocks HIGH severity and above. Good balance of security and usability.',
        4: 'Relaxed: Blocks CRITICAL severity only. Minimum protection, low false positives.'
    };
    var desc = document.getElementById('guardSensitivityDesc');
    if (desc) desc.textContent = descs[level] || '';

    fetch('/api/guard/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sensitivity: level })
    }).catch(function () { });
}

// ─── User Management ────────────────────────────

var _deleteUserID = null;

function loadUserProfiles() {
    fetch('/api/profiles')
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (!data.ok) return;
            var profiles = data.profiles || [];
            renderUserCards(profiles);

            // Update summary stats.
            var el = document.getElementById('usersTotalCount');
            if (el) el.textContent = profiles.length;

            var totalActive = 0;
            var totalSpend = 0;
            profiles.forEach(function (p) {
                totalActive += (p.active_tokens || 0);
                totalSpend += (p.estimated_usd || 0);
            });

            var elTokens = document.getElementById('usersActiveTokens');
            if (elTokens) elTokens.textContent = totalActive;

            var elSpend = document.getElementById('usersTotalSpend');
            if (elSpend) elSpend.textContent = '$' + totalSpend.toFixed(2);
        })
        .catch(function () {
            var container = document.getElementById('usersListContainer');
            if (container) container.innerHTML = '<p class="text-secondary" style="text-align:center;padding:2rem;">Failed to load users.</p>';
        });
}

function renderUserCards(profiles) {
    var container = document.getElementById('usersListContainer');
    if (!container) return;

    if (!profiles || profiles.length === 0) {
        container.innerHTML = '<div class="info-card" style="text-align:center;padding:2rem;">' +
            '<p class="text-secondary">No users yet. Click "+ Add User" to create one.</p>' +
            '<p class="text-secondary" style="font-size:0.75rem;margin-top:0.5rem;">Each user gets their own Shield API key and can bring their own provider.</p>' +
            '</div>';
        return;
    }

    var html = '';
    profiles.forEach(function (p) {
        var typeBadge = p.type === 'business'
            ? '<span style="background:var(--primary);color:#fff;padding:2px 8px;border-radius:4px;font-size:0.65rem;text-transform:uppercase;">Business</span>'
            : '<span style="background:rgba(255,255,255,0.1);color:var(--text-secondary);padding:2px 8px;border-radius:4px;font-size:0.65rem;text-transform:uppercase;">Personal</span>';

        var providerLabel = p.provider || 'Not set';
        var providerColors = {
            'openai': '#10a37f',
            'anthropic': '#d4a574',
            'google': '#4285f4',
            'mistral': '#ff7000',
            'groq': '#f55036',
            'local': '#888'
        };
        var provColor = providerColors[p.provider] || 'var(--text-secondary)';

        var maskedKey = p.upstream_key || 'Not configured';

        html += '<div class="info-card" style="margin-bottom:0.75rem;position:relative;">';
        html += '<div style="display:flex;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;gap:0.5rem;">';
        html += '<div>';
        html += '<div style="display:flex;align-items:center;gap:0.5rem;margin-bottom:0.25rem;">';
        html += '<strong style="font-size:1rem;">' + escapeHtml(p.name) + '</strong>';
        html += typeBadge;
        html += '</div>';
        if (p.email) html += '<div class="text-secondary" style="font-size:0.75rem;">' + escapeHtml(p.email) + '</div>';
        html += '</div>';
        html += '<div style="display:flex;gap:0.5rem;">';
        html += '<button class="btn btn-ghost btn-sm" onclick="promptDeleteUser(\'' + p.id + '\',\'' + escapeHtml(p.name) + '\')" style="color:var(--danger);font-size:0.7rem;">Delete</button>';
        html += '</div>';
        html += '</div>';

        html += '<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:0.5rem;margin-top:0.75rem;">';

        // Provider
        html += '<div style="padding:0.5rem;background:rgba(0,0,0,0.2);border-radius:6px;">';
        html += '<div class="text-secondary" style="font-size:0.65rem;margin-bottom:0.15rem;">Provider</div>';
        html += '<div style="font-size:0.8rem;color:' + provColor + ';">' + providerLabel.charAt(0).toUpperCase() + providerLabel.slice(1) + '</div>';
        html += '</div>';

        // API Key (masked)
        html += '<div style="padding:0.5rem;background:rgba(0,0,0,0.2);border-radius:6px;">';
        html += '<div class="text-secondary" style="font-size:0.65rem;margin-bottom:0.15rem;">API Key</div>';
        html += '<div style="font-size:0.8rem;font-family:monospace;">' + maskedKey + '</div>';
        html += '</div>';

        // Active Keys
        html += '<div style="padding:0.5rem;background:rgba(0,0,0,0.2);border-radius:6px;">';
        html += '<div class="text-secondary" style="font-size:0.65rem;margin-bottom:0.15rem;">Shield Keys</div>';
        html += '<div style="font-size:0.8rem;">' + (p.active_tokens || 0) + ' active / ' + (p.total_tokens || 0) + ' total</div>';
        html += '</div>';

        // Tokens Used
        html += '<div style="padding:0.5rem;background:rgba(0,0,0,0.2);border-radius:6px;">';
        html += '<div class="text-secondary" style="font-size:0.65rem;margin-bottom:0.15rem;">Tokens Used</div>';
        html += '<div style="font-size:0.8rem;">' + formatNumber(p.input_tokens || 0) + ' in / ' + formatNumber(p.output_tokens || 0) + ' out</div>';
        html += '</div>';

        // Spend
        html += '<div style="padding:0.5rem;background:rgba(0,0,0,0.2);border-radius:6px;">';
        html += '<div class="text-secondary" style="font-size:0.65rem;margin-bottom:0.15rem;">Est. Spend</div>';
        html += '<div style="font-size:0.8rem;">$' + (p.estimated_usd || 0).toFixed(4) + '</div>';
        html += '</div>';

        // Rate Limit
        html += '<div style="padding:0.5rem;background:rgba(0,0,0,0.2);border-radius:6px;">';
        html += '<div class="text-secondary" style="font-size:0.65rem;margin-bottom:0.15rem;">Rate Limit</div>';
        html += '<div style="font-size:0.8rem;">' + (p.rate_limit > 0 ? p.rate_limit + ' req/min' : 'Unlimited') + '</div>';
        html += '</div>';

        html += '</div>';
        html += '</div>';
    });

    container.innerHTML = html;
}

function escapeHtml(str) {
    var div = document.createElement('div');
    div.textContent = str || '';
    return div.innerHTML;
}

function formatNumber(n) {
    if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
    if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
    return String(n);
}

// ─── Add User Modal ─────────────────────────────

function openAddUserModal() {
    document.getElementById('addUserModal').style.display = 'flex';
    document.getElementById('addUserName').value = '';
    document.getElementById('addUserEmail').value = '';
    document.getElementById('addUserType').value = 'personal';
    document.getElementById('addUserProvider').value = '';
    document.getElementById('addUserAPIKey').value = '';
    document.getElementById('addUserRateLimit').value = '0';
    document.getElementById('addUserName').focus();
}

function closeAddUserModal() {
    document.getElementById('addUserModal').style.display = 'none';
}

function submitAddUser() {
    var name = document.getElementById('addUserName').value.trim();
    if (!name) {
        alert('Name is required');
        return;
    }

    var body = {
        name: name,
        email: document.getElementById('addUserEmail').value.trim(),
        type: document.getElementById('addUserType').value,
        provider: document.getElementById('addUserProvider').value,
        upstream_key: document.getElementById('addUserAPIKey').value.trim(),
        rate_limit: parseInt(document.getElementById('addUserRateLimit').value) || 0
    };

    fetch('/api/profiles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
    })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (data.ok) {
                closeAddUserModal();
                loadUserProfiles();
                if (data.token && data.token.key) {
                    alert('User created! Their Shield API key is:\n\n' + data.token.key + '\n\nCopy this key now. It will not be shown again.');
                }
            } else {
                alert('Error: ' + (data.error || 'Unknown error'));
            }
        })
        .catch(function () {
            alert('Failed to create user. Check your connection.');
        });
}

// ─── Delete User ────────────────────────────────

function promptDeleteUser(userID, userName) {
    _deleteUserID = userID;
    var text = document.getElementById('deleteConfirmText');
    if (text) text.textContent = 'Are you sure you want to delete "' + userName + '"? This action cannot be undone. All their tokens will be revoked.';
    document.getElementById('deleteConfirmModal').style.display = 'flex';
}

function closeDeleteConfirm() {
    document.getElementById('deleteConfirmModal').style.display = 'none';
    _deleteUserID = null;
}

function confirmDeleteUser() {
    if (!_deleteUserID) return;
    fetch('/api/profiles/' + _deleteUserID, { method: 'DELETE' })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            closeDeleteConfirm();
            if (data.ok) {
                loadUserProfiles();
            } else {
                alert('Error: ' + (data.error || 'Failed to delete'));
            }
        })
        .catch(function () {
            closeDeleteConfirm();
            alert('Failed to delete user.');
        });
}

// Auto-load profiles when switching to Users tab, backups when switching to Settings.
var _origSwitchTab = switchTab;
switchTab = function (tab) {
    _origSwitchTab(tab);
    if (tab === 'users') loadUserProfiles();
    if (tab === 'settings') { loadBackups(); loadPricing(); }
};

// ─── Backup Management ──────────────────────────

function createBackup() {
    fetch('/api/backups/create', { method: 'POST' })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (data.ok) {
                loadBackups();
            } else {
                alert('Backup failed: ' + (data.error || 'Unknown error'));
            }
        })
        .catch(function () { alert('Failed to create backup.'); });
}

function loadBackups() {
    fetch('/api/backups')
        .then(function (r) { return r.json(); })
        .then(function (data) {
            var container = document.getElementById('backupsList');
            if (!container) return;

            var backups = data.backups || [];
            if (backups.length === 0) {
                container.innerHTML = '<p class="text-secondary" style="font-size:0.75rem;">No backups yet. The first auto-backup will run in 12 hours.</p>';
                return;
            }

            var html = '';
            backups.forEach(function (b) {
                var sizeKB = (b.size / 1024).toFixed(1);
                var date = new Date(b.created_at).toLocaleString();
                html += '<div style="display:flex;justify-content:space-between;align-items:center;padding:0.35rem 0;border-bottom:1px solid var(--border);font-size:0.75rem;">';
                html += '<div>';
                html += '<span style="font-family:monospace;">' + b.name + '</span>';
                html += '<span class="text-secondary" style="margin-left:0.5rem;">' + sizeKB + ' KB</span>';
                html += '<span class="text-secondary" style="margin-left:0.5rem;">' + date + '</span>';
                html += '</div>';
                html += '<button class="btn btn-ghost btn-sm" style="font-size:0.65rem;" onclick="restoreBackup(\'' + b.path.replace(/\\/g, '\\\\') + '\')">Restore</button>';
                html += '</div>';
            });
            container.innerHTML = html;
        })
        .catch(function () {
            var container = document.getElementById('backupsList');
            if (container) container.innerHTML = '<p class="text-secondary" style="font-size:0.75rem;">Failed to load backups.</p>';
        });
}

function restoreBackup(path) {
    if (!confirm('Are you sure you want to restore this backup? This will replace all current data.')) return;

    fetch('/api/backups/restore', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: path })
    })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (data.ok) {
                alert('Backup restored successfully. Refreshing...');
                window.location.reload();
            } else {
                alert('Restore failed: ' + (data.error || 'Unknown error'));
            }
        })
        .catch(function () { alert('Failed to restore backup.'); });
}

// ─── Pricing Management ────────────────────────

function loadPricing() {
    fetch('/api/pricing')
        .then(function (r) { return r.json(); })
        .then(function (data) {
            var container = document.getElementById('pricingTable');
            if (!container) return;

            var profiles = data.profiles || [];
            if (profiles.length === 0) {
                container.innerHTML = '<p class="text-secondary" style="font-size:0.75rem;">No pricing profiles configured.</p>';
                return;
            }

            // Group by provider.
            var groups = {};
            profiles.forEach(function (p) {
                var key = p.provider || 'other';
                if (!groups[key]) groups[key] = [];
                groups[key].push(p);
            });

            var html = '<table style="width:100%;border-collapse:collapse;font-size:0.8rem;">';
            html += '<thead><tr style="border-bottom:1px solid var(--border);text-align:left;">';
            html += '<th style="padding:0.35rem 0.5rem;">Provider</th>';
            html += '<th style="padding:0.35rem 0.5rem;">Model</th>';
            html += '<th style="padding:0.35rem 0.5rem;text-align:right;">Input $/1M</th>';
            html += '<th style="padding:0.35rem 0.5rem;text-align:right;">Output $/1M</th>';
            html += '</tr></thead><tbody>';

            var providerColors = {
                openai: '#10a37f', anthropic: '#d4a574', google: '#4285f4',
                xai: '#1da1f2', deepseek: '#4d6bfe', mistral: '#ff6b35',
                cohere: '#39594d', nvidia: '#76b900', qwen: '#6f42c1',
                kimi: '#ff9500', zhipu: '#2352d8', meta: '#0668e1',
                openrouter: '#6366f1', groq: '#f55036', together: '#0ea5e9',
                fireworks: '#ff4500', huggingface: '#ffcc00', minimax: '#e040fb',
                perplexity: '#20808d', cerebras: '#ff6b00', deepinfra: '#7c3aed',
                siliconflow: '#00bcd4', manus: '#8b5cf6',
                custom: '#9b59b6', local: '#888'
            };

            Object.keys(groups).sort().forEach(function (provider) {
                var color = providerColors[provider] || '#888';
                groups[provider].forEach(function (p) {
                    html += '<tr style="border-bottom:1px solid var(--border);">';
                    html += '<td style="padding:0.35rem 0.5rem;"><span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:' + color + ';margin-right:0.4rem;"></span>' + provider + '</td>';
                    html += '<td style="padding:0.35rem 0.5rem;">' + p.name;
                    if (p.is_default) html += ' <span style="font-size:0.6rem;color:var(--primary);border:1px solid var(--primary);border-radius:3px;padding:0 3px;">DEFAULT</span>';
                    html += '</td>';
                    html += '<td style="padding:0.35rem 0.5rem;text-align:right;"><input type="number" step="0.01" min="0" value="' + p.input_per_million.toFixed(2) + '" style="width:70px;background:rgba(0,0,0,0.3);border:1px solid var(--border);border-radius:4px;color:var(--text-primary);padding:2px 6px;font-size:0.75rem;text-align:right;" onchange="savePricingRate(\'' + p.id + '\',\'' + p.name + '\',\'' + p.provider + '\',' + (p.is_default ? 'true' : 'false') + ',parseFloat(this.value),' + p.output_per_million + ')"></td>';
                    html += '<td style="padding:0.35rem 0.5rem;text-align:right;"><input type="number" step="0.01" min="0" value="' + p.output_per_million.toFixed(2) + '" style="width:70px;background:rgba(0,0,0,0.3);border:1px solid var(--border);border-radius:4px;color:var(--text-primary);padding:2px 6px;font-size:0.75rem;text-align:right;" onchange="savePricingRate(\'' + p.id + '\',\'' + p.name + '\',\'' + p.provider + '\',' + (p.is_default ? 'true' : 'false') + ',' + p.input_per_million + ',parseFloat(this.value))"></td>';
                    html += '</tr>';
                });
            });

            html += '</tbody></table>';
            container.innerHTML = html;
        })
        .catch(function () {
            var container = document.getElementById('pricingTable');
            if (container) container.innerHTML = '<p class="text-secondary" style="font-size:0.75rem;">Failed to load pricing.</p>';
        });
}

function savePricingRate(id, name, provider, isDefault, inputRate, outputRate) {
    fetch('/api/pricing', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            id: id,
            name: name,
            provider: provider,
            input_per_million: inputRate,
            output_per_million: outputRate,
            is_default: isDefault
        })
    })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (!data.ok) {
                console.error('Failed to save pricing:', data.error);
            }
        })
        .catch(function (err) { console.error('Pricing save error:', err); });
}
