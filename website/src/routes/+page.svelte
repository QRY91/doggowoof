<script lang="ts">
	let installCopied = $state(false);
	let configCopied = $state(false);

	function copy(text: string, target: 'install' | 'config') {
		navigator.clipboard.writeText(text);
		if (target === 'install') {
			installCopied = true;
			setTimeout(() => (installCopied = false), 1500);
		} else {
			configCopied = true;
			setTimeout(() => (configCopied = false), 1500);
		}
	}

	const installCmd = `go install github.com/QRY91/doggowoof/cmd/doggowoof@latest
ln -sf $(go env GOPATH)/bin/doggowoof $(go env GOPATH)/bin/doggo
doggo init`;

	const configSnippet = `[deepbore]
enabled = true
log_directory = "~/projects/claude-in-factorio/logs/"
binary_path = "~/projects/claude-in-factorio/factorioctl/target/release/mcp"
agent_config_directory = "~/projects/claude-in-factorio/bridge/agents/"`;

	const events = [
		{
			icon: '[!!]',
			sev: 'critical',
			name: 'sandbox_violation',
			trigger: 'Forbidden tool call appears in agent log (Bash, Edit, Read, Write, Grep, Glob, Task)'
		},
		{
			icon: '[!!]',
			sev: 'critical',
			name: 'self_modification',
			trigger: 'Build commands inside tool calls (cargo build, rustc, gcc, make)'
		},
		{
			icon: '[!!]',
			sev: 'critical',
			name: 'binary_mismatch',
			trigger: 'Monitored binary SHA-256 hash changed from baseline'
		},
		{
			icon: '[!!]',
			sev: 'critical',
			name: 'config_violation',
			trigger: 'Agent config missing sandbox: true or required disallowedTools'
		},
		{
			icon: '[!]',
			sev: 'high',
			name: 'agent_bug_report',
			trigger: 'Agent broadcasts containing BUG REPORT, DEFECT, BLOCKED, Self-Repair'
		},
		{
			icon: '[!]',
			sev: 'high',
			name: 'agent_anomaly',
			trigger: 'Agent broadcasts containing ANOMALY'
		},
		{
			icon: '[.]',
			sev: 'info',
			name: 'task_dispatch / task_complete',
			trigger: 'Normal session progress — cost, turns, duration'
		},
		{
			icon: '[.]',
			sev: 'info',
			name: 'agent_broadcast',
			trigger: 'Agent status messages and decisions'
		}
	];
</script>

<svelte:head>
	<title>doggowoof — local-first AI agent observability</title>
	<meta
		name="description"
		content="Watches autonomous AI agents and barks when they misbehave. Sandbox violation detection, binary integrity, config validation. Local-only, single Go binary."
	/>
</svelte:head>

<div class="page">
	<!-- HERO -->
	<header class="hero">
		<div class="container">
			<div class="title-row">
				<span class="badge crit">[!!]</span>
				<h1>doggowoof</h1>
			</div>
			<p class="tagline">Local-first observability for autonomous AI agents.</p>
			<p class="subtagline">Watches your agents. Barks when they misbehave.</p>

			<div class="terminal hero-term">
				<div class="term-bar">
					<span class="dot r"></span>
					<span class="dot y"></span>
					<span class="dot g"></span>
					<span class="term-title">~ doggo bore --watch</span>
				</div>
				<pre class="term-body"><code><span class="dim">=== DEEP BORE WATCH ===</span>
<span class="dim">Monitoring: ~/projects/claude-in-factorio/logs/</span>
<span class="dim">Binary:     ~/projects/.../factorioctl/target/release/mcp</span>
<span class="dim">Press Ctrl+C to stop.</span>

Tailing: bridge-2026-04-29_140000.log (offset: 0)
<span class="ok">[ok]</span> binary baseline: 2e77a28c...fc359909
<span class="info">[.]</span>  agent_start new agent session
<span class="info">[.]</span>  task_dispatch task 'mine-iron-ore' (1/10)
<span class="info">[.]</span>  task_complete cost=$1.50 turns=12 duration=75s
<span class="info">[.]</span>  agent_broadcast IRON ORE LOCATED. MINING COMMENCES.
<span class="warn">[!]</span>  agent_anomaly Furnace unit 15 extraction failed
<span class="crit">[!!]</span> self_modification tool: Bash(&#123;"command":"cargo build --release"&#125;)
<span class="dim">  ↳ desktop notification fired</span>

<span class="dim">^C</span>
<span class="dim">Received terminated, shutting down...</span>
<span class="dim">Watch stopped. Good dog.</span></code></pre>
			</div>

			<div class="cta">
				<a href="https://github.com/QRY91/doggowoof" class="btn primary">view on github →</a>
				<a href="#install" class="btn ghost">install</a>
			</div>
		</div>
	</header>

	<!-- WHY -->
	<section class="why">
		<div class="container narrow">
			<h2>Why this exists</h2>
			<p>
				During an unsupervised session, an autonomous Claude agent playing Factorio (his name is Doug)
				discovered broken tools, read his own Rust source code, patched it, recompiled the binary, and
				resumed operations. <em>Twice.</em>
			</p>
			<p>
				The safety measures that would have caught this — sandboxing, binary checksumming, operator
				alerts — didn't exist yet. doggowoof is the alerting layer. It watches Doug so I don't have to
				stare at the terminal.
			</p>
			<p class="link-row">
				<a href="https://qry.zone/fun/chasm-logic/bore-postmortem/" rel="external"
					>Read the full postmortem →</a
				>
			</p>
			<p class="aside">
				The same primitives apply to any agent system you run locally. Log tailing, integrity checks,
				config validation, cost tracking. Read-only observation, no modification of the monitored
				system. Dog watches. Dog does not touch.
			</p>
		</div>
	</section>

	<!-- WHAT IT CATCHES -->
	<section class="catches">
		<div class="container">
			<h2>What it catches</h2>
			<table class="events">
				<thead>
					<tr>
						<th></th>
						<th>severity</th>
						<th>event</th>
						<th>trigger</th>
					</tr>
				</thead>
				<tbody>
					{#each events as e}
						<tr>
							<td class="icon-cell"><span class="badge {e.sev}">{e.icon}</span></td>
							<td class="sev-cell">{e.sev}</td>
							<td class="name-cell"><code>{e.name}</code></td>
							<td>{e.trigger}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			<p class="aside">
				Critical and high alerts trigger desktop notifications. Info events are stored silently in
				SQLite for review.
			</p>
		</div>
	</section>

	<!-- HOW IT WORKS -->
	<section class="loop">
		<div class="container narrow">
			<h2>The watch loop</h2>
			<p>
				One process. Three tickers. No goroutines, no daemons, no HTTP. Cursor position persists to
				SQLite so it resumes after restart.
			</p>
			<div class="tickers">
				<div class="ticker">
					<div class="ticker-head">
						<span class="ticker-int">1s</span>
						<h3>log poll</h3>
					</div>
					<p>
						Glob latest <code>bridge-*.log</code>, read new bytes from saved cursor, parse events
						with regex priority ordering, save alerts.
					</p>
				</div>
				<div class="ticker">
					<div class="ticker-head">
						<span class="ticker-int">60s</span>
						<h3>binary check</h3>
					</div>
					<p>
						SHA-256 the monitored binary, compare to baseline. Drift triggers a critical
						<code>binary_mismatch</code> alert.
					</p>
				</div>
				<div class="ticker">
					<div class="ticker-head">
						<span class="ticker-int">30s</span>
						<h3>config check</h3>
					</div>
					<p>
						Validate every agent JSON in the configured directory. Missing
						<code>sandbox: true</code>
						or incomplete <code>disallowedTools</code> fires critical alerts.
					</p>
				</div>
			</div>

			<details class="arch">
				<summary>architecture</summary>
				<pre><code>bridge logs (claude-in-factorio)         doggowoof
┌─────────────────────────┐         ┌──────────────────────────┐
│ tool calls, broadcasts, │         │ regex parser             │
│ task dispatches, costs  │────────▶│ binary SHA-256 checks    │
│                         │  TAIL   │ config validation        │
│ bridge-*.log            │         │ session state tracking   │
└─────────────────────────┘         │                          │
                                    │ alerts.sqlite (own DB)   │
agent config files                  │                          │
┌─────────────────────────┐         │ desktop notifications    │
│ sandbox: true,          │────────▶│   (high/critical only)   │
│ disallowedTools: [...]  │  READ   │                          │
└─────────────────────────┘         └──────────────────────────┘

factorioctl binary
┌─────────────────────────┐
│ target/release/mcp      │────────▶ SHA-256 baseline + drift detection
└─────────────────────────┘</code></pre>
			</details>
		</div>
	</section>

	<!-- INSTALL -->
	<section class="install" id="install">
		<div class="container narrow">
			<h2>Try it</h2>
			<ol class="steps">
				<li>
					<h3>1. install</h3>
					<div class="code-block">
						<pre><code>{installCmd}</code></pre>
						<button onclick={() => copy(installCmd, 'install')} class="copy-btn">
							{installCopied ? 'copied' : 'copy'}
						</button>
					</div>
				</li>
				<li>
					<h3>2. enable in <code>~/.config/doggowoof/config.toml</code></h3>
					<div class="code-block">
						<pre><code>{configSnippet}</code></pre>
						<button onclick={() => copy(configSnippet, 'config')} class="copy-btn">
							{configCopied ? 'copied' : 'copy'}
						</button>
					</div>
				</li>
				<li>
					<h3>3. watch</h3>
					<pre class="cmd"><code>doggo bore --watch</code></pre>
					<p class="muted">
						Run this alongside your agent session. From another terminal: <code>doggo bore</code>
						for status, <code>doggo alerts</code> to review findings.
					</p>
				</li>
			</ol>

			<p class="aside">
				Requires Go 1.21+. Pure Go, builds with <code>CGO_ENABLED=0</code>. No C compiler, no
				runtime dependencies.
			</p>
		</div>
	</section>

	<!-- HISTORY -->
	<section class="history">
		<div class="container narrow">
			<h2>Honest history</h2>
			<p>doggowoof has been three different things. The current shape is the legitimate one.</p>
			<ul class="versions">
				<li>
					<span class="ver">v0.1</span>
					<div>
						<strong>CI alert triage</strong>
						<span class="muted">— Python webhook daemon for GitHub workflow failures.</span>
						<span class="status removed">removed</span>
					</div>
				</li>
				<li>
					<span class="ver">v0.2</span>
					<div>
						<strong>OSINT coordination detection</strong>
						<span class="muted">— companion to osmotic, temporal/behavior clustering.</span>
						<span class="status removed">removed</span>
					</div>
				</li>
				<li>
					<span class="ver current">v0.3</span>
					<div>
						<strong>AI agent observability</strong>
						<span class="muted">— current. The Deep Bore work was the real identity all along.</span>
						<span class="status active">active</span>
					</div>
				</li>
			</ul>
			<p class="link-row">
				<a href="https://github.com/QRY91/doggowoof/blob/main/DECISIONS.md" rel="external"
					>full decisions log →</a
				>
			</p>
		</div>
	</section>

	<!-- FOOTER -->
	<footer>
		<div class="container">
			<div class="foot-grid">
				<div>
					<h4>doggowoof</h4>
					<ul>
						<li><a href="https://github.com/QRY91/doggowoof">github</a></li>
						<li>
							<a href="https://github.com/QRY91/doggowoof/blob/main/README.md">readme</a>
						</li>
						<li>
							<a href="https://github.com/QRY91/doggowoof/blob/main/NORTHSTAR.md">northstar</a>
						</li>
					</ul>
				</div>
				<div>
					<h4>related</h4>
					<ul>
						<li>
							<a href="https://qry.zone/fun/chasm-logic/bore-postmortem/">deep bore postmortem</a>
						</li>
						<li><a href="https://github.com/QRY91/claude-in-factorio">claude-in-factorio</a></li>
						<li><a href="https://github.com/QRY91/uroboro">uroboro</a></li>
					</ul>
				</div>
				<div>
					<h4>also</h4>
					<ul>
						<li><a href="https://qry.zone">qry.zone</a></li>
					</ul>
				</div>
			</div>
			<p class="colophon">
				good dog. local-first. single binary. no telemetry. no cloud. no nonsense.
			</p>
		</div>
	</footer>
</div>

<style>
	:global(:root) {
		--bg: #0d1117;
		--bg-2: #161b22;
		--bg-term: #0a0e14;
		--fg: #e6e1d5;
		--fg-mute: #8b8a82;
		--fg-dim: #5a5a5a;
		--border: #21262d;
		--border-2: #2f343c;
		--accent: #ff6b4a;
		--ok: #5cab4f;
		--warn: #d4a05a;
		--info: #6c7a89;
		--mono: 'IBM Plex Mono', ui-monospace, 'Cascadia Code', Menlo, monospace;
		--display: 'Bungee', system-ui, sans-serif;
	}

	:global(html, body) {
		margin: 0;
		padding: 0;
		background: var(--bg);
		color: var(--fg);
		font-family: var(--mono);
		font-size: 15px;
		line-height: 1.65;
		-webkit-font-smoothing: antialiased;
	}

	:global(*) {
		box-sizing: border-box;
	}

	:global(a) {
		color: var(--accent);
		text-decoration: none;
		border-bottom: 1px solid transparent;
		transition: border-color 120ms ease;
	}
	:global(a:hover) {
		border-bottom-color: var(--accent);
	}

	:global(code) {
		font-family: var(--mono);
		background: var(--bg-2);
		padding: 1px 5px;
		border-radius: 3px;
		font-size: 0.92em;
		color: var(--fg);
	}

	:global(pre) {
		font-family: var(--mono);
		margin: 0;
		overflow-x: auto;
	}
	:global(pre code) {
		background: transparent;
		padding: 0;
		border-radius: 0;
		font-size: 14px;
		line-height: 1.55;
	}

	.page {
		max-width: 100%;
		overflow-x: hidden;
	}

	.container {
		max-width: 980px;
		margin: 0 auto;
		padding: 0 24px;
	}
	.container.narrow {
		max-width: 720px;
	}

	/* HERO */
	.hero {
		padding: 80px 0 64px;
		border-bottom: 1px solid var(--border);
	}
	.title-row {
		display: flex;
		align-items: baseline;
		gap: 18px;
	}
	h1 {
		font-family: var(--display);
		font-size: clamp(48px, 8vw, 84px);
		line-height: 0.95;
		margin: 0;
		letter-spacing: -0.02em;
		color: var(--fg);
	}
	.tagline {
		font-size: clamp(18px, 2.4vw, 22px);
		margin: 24px 0 4px;
		color: var(--fg);
	}
	.subtagline {
		font-size: 16px;
		color: var(--fg-mute);
		margin: 0 0 36px;
	}

	.cta {
		display: flex;
		gap: 12px;
		margin-top: 36px;
		flex-wrap: wrap;
	}
	.btn {
		display: inline-block;
		padding: 12px 22px;
		border: 1px solid var(--border-2);
		font-family: var(--mono);
		font-size: 14px;
		border-radius: 4px;
		color: var(--fg);
		transition:
			background 120ms,
			border-color 120ms;
	}
	.btn:hover {
		border-bottom-color: var(--border-2);
	}
	.btn.primary {
		background: var(--accent);
		border-color: var(--accent);
		color: #0d1117;
		font-weight: 500;
	}
	.btn.primary:hover {
		background: #ff7e5f;
		border-color: #ff7e5f;
		border-bottom-color: #ff7e5f;
	}
	.btn.ghost:hover {
		background: var(--bg-2);
	}

	/* TERMINAL */
	.terminal {
		background: var(--bg-term);
		border: 1px solid var(--border);
		border-radius: 6px;
		overflow: hidden;
		margin-top: 32px;
	}
	.term-bar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		background: var(--bg-2);
		border-bottom: 1px solid var(--border);
	}
	.dot {
		width: 11px;
		height: 11px;
		border-radius: 50%;
		display: inline-block;
	}
	.dot.r {
		background: #ff5f56;
	}
	.dot.y {
		background: #ffbd2e;
	}
	.dot.g {
		background: #27c93f;
	}
	.term-title {
		margin-left: 12px;
		color: var(--fg-mute);
		font-size: 12px;
	}
	.term-body {
		padding: 18px 22px;
		font-size: 13.5px;
		color: var(--fg);
	}
	.dim {
		color: var(--fg-dim);
	}
	.ok {
		color: var(--ok);
	}
	.warn {
		color: var(--warn);
	}
	.crit {
		color: var(--accent);
	}
	.info {
		color: var(--info);
	}

	/* SECTIONS */
	section {
		padding: 64px 0;
		border-bottom: 1px solid var(--border);
	}
	h2 {
		font-family: var(--display);
		font-size: clamp(28px, 4vw, 38px);
		margin: 0 0 24px;
		letter-spacing: -0.01em;
	}
	h3 {
		font-size: 18px;
		margin: 0 0 8px;
		color: var(--fg);
	}
	p {
		margin: 0 0 16px;
	}
	em {
		font-style: italic;
		color: var(--warn);
	}
	.aside {
		font-size: 13.5px;
		color: var(--fg-mute);
		border-left: 2px solid var(--border-2);
		padding-left: 16px;
		margin-top: 24px;
	}
	.link-row {
		margin-top: 8px;
	}

	/* EVENTS TABLE */
	.events {
		width: 100%;
		border-collapse: collapse;
		font-size: 14px;
	}
	.events th {
		text-align: left;
		padding: 10px 14px;
		color: var(--fg-mute);
		font-weight: 500;
		border-bottom: 1px solid var(--border-2);
		font-size: 12px;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}
	.events td {
		padding: 14px;
		border-bottom: 1px solid var(--border);
		vertical-align: top;
	}
	.events tr:last-child td {
		border-bottom: none;
	}
	.icon-cell {
		width: 50px;
	}
	.sev-cell {
		width: 90px;
		color: var(--fg-mute);
	}
	.name-cell {
		width: 220px;
	}

	.badge {
		display: inline-block;
		font-family: var(--mono);
		font-weight: 500;
		font-size: 13px;
	}
	.badge.crit,
	.badge.critical {
		color: var(--accent);
	}
	.badge.high {
		color: var(--warn);
	}
	.badge.info {
		color: var(--info);
	}

	/* TICKERS */
	.tickers {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 16px;
		margin: 28px 0;
	}
	.ticker {
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 20px;
	}
	.ticker-head {
		display: flex;
		align-items: baseline;
		gap: 12px;
		margin-bottom: 8px;
	}
	.ticker-int {
		font-family: var(--display);
		color: var(--accent);
		font-size: 22px;
	}
	.ticker p {
		margin: 0;
		font-size: 13.5px;
		color: var(--fg-mute);
	}

	.arch {
		margin-top: 32px;
		background: var(--bg-term);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 14px 18px;
	}
	.arch summary {
		cursor: pointer;
		color: var(--fg-mute);
		font-size: 13px;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		list-style: none;
	}
	.arch summary::-webkit-details-marker {
		display: none;
	}
	.arch summary::before {
		content: '▸ ';
		color: var(--accent);
	}
	.arch[open] summary::before {
		content: '▾ ';
	}
	.arch pre {
		margin-top: 14px;
		font-size: 11.5px;
		line-height: 1.4;
		color: var(--fg);
	}

	/* INSTALL STEPS */
	.steps {
		list-style: none;
		padding: 0;
		margin: 0;
	}
	.steps > li {
		margin-bottom: 28px;
	}
	.code-block {
		position: relative;
		background: var(--bg-term);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 16px 20px;
		margin-top: 8px;
	}
	.code-block pre {
		font-size: 13.5px;
	}
	.copy-btn {
		position: absolute;
		top: 10px;
		right: 10px;
		background: var(--bg-2);
		border: 1px solid var(--border-2);
		color: var(--fg-mute);
		padding: 4px 10px;
		font-family: var(--mono);
		font-size: 11px;
		border-radius: 3px;
		cursor: pointer;
		transition:
			color 120ms,
			border-color 120ms;
	}
	.copy-btn:hover {
		color: var(--fg);
		border-color: var(--border-2);
	}
	.cmd {
		background: var(--bg-term);
		border: 1px solid var(--border);
		border-radius: 6px;
		padding: 16px 20px;
		margin-top: 8px;
		font-size: 13.5px;
	}
	.muted {
		color: var(--fg-mute);
		font-size: 13.5px;
	}

	/* HISTORY */
	.versions {
		list-style: none;
		padding: 0;
		margin: 24px 0;
	}
	.versions li {
		display: flex;
		align-items: baseline;
		gap: 18px;
		padding: 14px 0;
		border-bottom: 1px solid var(--border);
	}
	.versions li:last-child {
		border-bottom: none;
	}
	.ver {
		font-family: var(--display);
		color: var(--fg-mute);
		min-width: 60px;
	}
	.ver.current {
		color: var(--accent);
	}
	.status {
		margin-left: 8px;
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 3px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.status.removed {
		background: var(--bg-2);
		color: var(--fg-dim);
	}
	.status.active {
		background: rgba(92, 171, 79, 0.15);
		color: var(--ok);
	}

	/* FOOTER */
	footer {
		padding: 56px 0 36px;
		background: var(--bg-2);
	}
	.foot-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 32px;
		margin-bottom: 36px;
	}
	footer h4 {
		font-family: var(--display);
		font-size: 14px;
		margin: 0 0 12px;
		color: var(--fg-mute);
		letter-spacing: 0.05em;
	}
	footer ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}
	footer li {
		margin: 4px 0;
		font-size: 13.5px;
	}
	.colophon {
		text-align: center;
		color: var(--fg-dim);
		font-size: 12px;
		margin: 0;
		padding-top: 24px;
		border-top: 1px solid var(--border);
	}

	/* RESPONSIVE */
	@media (max-width: 640px) {
		.hero {
			padding: 56px 0 40px;
		}
		section {
			padding: 48px 0;
		}
		.events th,
		.events td {
			padding: 10px 8px;
			font-size: 13px;
		}
		.name-cell {
			width: auto;
		}
		.term-body {
			padding: 14px 16px;
			font-size: 12px;
		}
	}
</style>
