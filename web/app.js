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
                alert('Failed to start: ' + (data.error || 'Unknown error'));
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
    if (sidebar) sidebar.classList.remove('mobile-open');
    if (backdrop) backdrop.classList.remove('visible');
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

// ─── Mobile Drawer ───────────────────────────

function toggleMobileDrawer() {
    var sidebar = document.getElementById('sidebar');
    var backdrop = document.getElementById('sidebarBackdrop');
    if (!sidebar) return;

    var isOpen = sidebar.classList.toggle('mobile-open');
    if (backdrop) backdrop.classList.toggle('visible', isOpen);
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

document.addEventListener('DOMContentLoaded', function () {
    buildProviderGrid();
    document.getElementById('proxyPort').value = defaultRandomPort;
    checkInitialStatus();
    initHelpIcons();
    initSidebarResize();

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
