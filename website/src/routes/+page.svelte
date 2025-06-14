<script lang="ts">
	let installCopied = false;
	let webhookCopied = false;

	function copyToClipboard(text: string, type: 'install' | 'webhook') {
		navigator.clipboard.writeText(text);
		if (type === 'install') {
			installCopied = true;
			setTimeout(() => installCopied = false, 2000);
		} else {
			webhookCopied = true;
			setTimeout(() => webhookCopied = false, 2000);
		}
	}

	const installCommand = `go install github.com/QRY91/doggowoof/cmd/doggowoof@latest
ln -sf $(go env GOPATH)/bin/doggowoof $(go env GOPATH)/bin/doggo
doggo init
doggo daemon start`;

	const webhookExample = `curl -X POST http://localhost:8080/webhook/github \\
  -H "Content-Type: application/json" \\
  -H "X-GitHub-Event: workflow_run" \\
  -d '{"action":"completed","workflow_run":{"name":"Tests","conclusion":"failure"},"repository":{"name":"your-repo"}}'`;
</script>

<svelte:head>
	<title>DOGGOWOOF 🚨🐕🚨 - The Local-First Alert Triage System That Actually Barks!</title>
	<meta name="description" content="Your personal guard dog for development alerts. Smart, private, and LOUD when it matters. Local-first alert system for developers." />
</svelte:head>

<main>
	<!-- Hero Section -->
	<section class="hero">
		<div class="hero-content">
			<h1 class="hero-title">
				<span class="title-main">DOGGOWOOF</span>
				<span class="title-emojis">🚨🐕🚨</span>
			</h1>
			<p class="hero-subtitle">THE LOCAL-FIRST ALERT TRIAGE SYSTEM THAT ACTUALLY BARKS!</p>
			<p class="hero-description">
				Your personal guard dog for development alerts. Smart, private, and <strong>LOUD</strong> when it matters.
			</p>
			<div class="hero-quote">
				🗣️ "WOOF! WOOF! HEY HUMAN, THIS ONE ACTUALLY MATTERS!" 🗣️
			</div>
			<div class="hero-actions">
				<a href="#get-started" class="btn btn-primary">GET YOUR GUARD DOG 🐕‍🦺</a>
				<a href="https://github.com/QRY91/doggowoof" class="btn btn-secondary">VIEW ON GITHUB</a>
			</div>
		</div>
	</section>

	<!-- Vision Section -->
	<section class="vision">
		<div class="container">
			<h2>🔥 STOP THE NOTIFICATION MADNESS!</h2>
			<p>
				DOGGOWOOF transforms information overload into actionable insights with the enthusiasm of a golden retriever
				and the precision of a border collie.
			</p>
			<div class="philosophy">
				<strong>PHILOSOPHY:</strong> Local-first, privacy-focused, ZERO BS monitoring with BIG ENERGY when alerts matter!
			</div>
		</div>
	</section>

	<!-- Features Grid -->
	<section class="features">
		<div class="container">
			<h2>🎯 WHAT YOUR GUARD DOG DOES</h2>
			<div class="features-grid">
				<div class="feature">
					<div class="feature-icon">🚨</div>
					<h3>Smart Alert Filtering</h3>
					<p>Only barks when something actually matters. No more notification fatigue!</p>
				</div>
				<div class="feature">
					<div class="feature-icon">🏠</div>
					<h3>100% Local-First</h3>
					<p>Your data stays on YOUR machine. No telemetry, no cloud dependencies!</p>
				</div>
				<div class="feature">
					<div class="feature-icon">⚡</div>
					<h3>GitHub CI Integration</h3>
					<p>Instantly know when your builds fail. Working right now!</p>
				</div>
				<div class="feature">
					<div class="feature-icon">🧠</div>
					<h3>Pattern Learning</h3>
					<p>Learns what you care about over time. Gets smarter with every alert!</p>
				</div>
				<div class="feature">
					<div class="feature-icon">📊</div>
					<h3>SQLite Storage</h3>
					<p>All your alert history stored locally in SQLite. Fast and reliable!</p>
				</div>
				<div class="feature">
					<div class="feature-icon">🔧</div>
					<h3>Webhook Ready</h3>
					<p>Easy integration with any service that can send HTTP requests!</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Status Section -->
	<section class="status">
		<div class="container">
			<h2>📊 MVP ACHIEVED! 🎉</h2>
			<div class="status-grid">
				<div class="status-working">
					<h3>✅ WORKING NOW</h3>
					<ul>
						<li>Go CLI with complete commands</li>
						<li>Python daemon with webhook receiver</li>
						<li>GitHub CI failure detection</li>
						<li>Desktop notification system</li>
						<li>SQLite storage and history</li>
						<li>Local-first operation</li>
					</ul>
				</div>
				<div class="status-coming">
					<h3>🔜 COMING SOON</h3>
					<ul>
						<li>Advanced pattern learning</li>
						<li>Smart filtering improvements</li>
						<li>Uroboro integration</li>
						<li>Email/IMAP monitoring</li>
						<li>Discord alert routing</li>
					</ul>
				</div>
			</div>
		</div>
	</section>

	<!-- Installation Section -->
	<section id="get-started" class="installation">
		<div class="container">
			<h2>🚀 GET THAT TAIL WAGGING!</h2>
			<div class="install-steps">
				<div class="step">
					<h3>1. Install Your Guard Dog</h3>
					<div class="code-block">
						<pre><code>{installCommand}</code></pre>
						<button
							class="copy-btn"
							on:click={() => copyToClipboard(installCommand, 'install')}
						>
							{installCopied ? '✅ Copied!' : '📋 Copy'}
						</button>
					</div>
				</div>
				<div class="step">
					<h3>2. Test with GitHub Webhook</h3>
					<div class="code-block">
						<pre><code>{webhookExample}</code></pre>
						<button
							class="copy-btn"
							on:click={() => copyToClipboard(webhookExample, 'webhook')}
						>
							{webhookCopied ? '✅ Copied!' : '📋 Copy'}
						</button>
					</div>
				</div>
				<div class="step">
					<h3>3. Pet Your Good Doggo!</h3>
					<div class="code-block">
						<pre><code>doggo pet</code></pre>
					</div>
					<p class="step-note">Secret command! Your guard dog deserves some love! 🐕❤️</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Footer -->
	<footer>
		<div class="container">
			<div class="footer-content">
				<div class="footer-main">
					<h3>🚨 WHY DOGGOWOOF? 🚨</h3>
					<p>Because your alerts deserve a guard dog that:</p>
					<ul>
						<li>✅ <strong>STAYS AWAKE</strong> while you code</li>
						<li>✅ <strong>BARKS LOUD</strong> when something's actually wrong</li>
						<li>✅ <strong>STAYS QUIET</strong> when it's just noise</li>
						<li>✅ <strong>LEARNS YOUR PATTERNS</strong> like a loyal companion</li>
						<li>✅ <strong>PROTECTS YOUR PRIVACY</strong> (no data leaves home!)</li>
					</ul>
				</div>
				<div class="footer-links">
					<a href="https://github.com/QRY91/doggowoof">GitHub</a>
					<a href="https://github.com/QRY91/doggowoof/issues">Issues</a>
					<a href="https://github.com/QRY91/doggowoof/blob/main/README.md">Documentation</a>
				</div>
			</div>
			<div class="footer-bottom">
				<p><strong>DOGGOWOOF: BECAUSE YOUR ALERTS DESERVE A VERY GOOD BOY WATCHING OVER THEM!</strong> 🚨🐕‍🦺🚨</p>
			</div>
		</div>
	</footer>
</main>

<style>
	:global(body) {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
		line-height: 1.6;
		color: #2d2016;
		background: linear-gradient(135deg, #d4a574 0%, #8b6914 100%);
	}

	main {
		min-height: 100vh;
	}

	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 2rem;
	}

	/* Hero Section */
	.hero {
		padding: 4rem 2rem;
		text-align: center;
		color: white;
		background: linear-gradient(135deg, #ff6b6b, #4ecdc4);
		position: relative;
		overflow: hidden;
	}

	.hero::before {
		content: '';
		position: absolute;
		top: -50%;
		left: -50%;
		width: 200%;
		height: 200%;
		background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E") repeat;
		animation: float 20s infinite linear;
		z-index: 0;
	}

	@keyframes float {
		0% { transform: translate(-50%, -50%) rotate(0deg); }
		100% { transform: translate(-50%, -50%) rotate(360deg); }
	}

	.hero-content {
		position: relative;
		z-index: 1;
	}

	.hero-title {
		font-family: 'Bangers', 'Fredoka One', cursive;
		font-size: clamp(3rem, 10vw, 6rem);
		margin: 0;
		font-weight: 400;
		text-shadow:
			0 0 10px rgba(255,255,255,0.8),
			0 0 20px rgba(255,71,87,0.6),
			0 0 30px rgba(255,71,87,0.4),
			4px 4px 0px rgba(0,0,0,0.8),
			8px 8px 0px rgba(255,71,87,0.3);
		transform: rotate(-2deg);
		animation: doggo-bounce 2s ease-in-out infinite alternate;
	}

	@keyframes doggo-bounce {
		0% { transform: rotate(-2deg) translateY(0px); }
		100% { transform: rotate(-1deg) translateY(-10px); }
	}

	.title-main {
		display: block;
		letter-spacing: 0.1em;
	}

	.title-emojis {
		display: block;
		font-size: 0.6em;
		margin-top: 0.5rem;
	}

	.hero-subtitle {
		font-family: 'Fredoka One', cursive;
		font-size: clamp(1rem, 3vw, 1.5rem);
		margin: 1rem 0;
		font-weight: 400;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		text-shadow: 2px 2px 4px rgba(0,0,0,0.5);
	}

	.hero-description {
		font-size: clamp(1.1rem, 2.5vw, 1.3rem);
		margin: 2rem 0;
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
	}

	.hero-quote {
		font-family: 'Caveat', cursive;
		background: rgba(255,255,255,0.2);
		padding: 1.5rem;
		border-radius: 15px;
		font-size: 1.4rem;
		font-weight: 700;
		margin: 2rem auto;
		max-width: 500px;
		backdrop-filter: blur(10px);
		border: 3px solid rgba(255,255,255,0.5);
		transform: rotate(1deg);
		text-shadow: 1px 1px 2px rgba(0,0,0,0.3);
	}

	.hero-actions {
		display: flex;
		gap: 1rem;
		justify-content: center;
		flex-wrap: wrap;
		margin-top: 3rem;
	}

	.btn {
		padding: 1rem 2rem;
		border-radius: 50px;
		text-decoration: none;
		font-weight: 700;
		font-size: 1.1rem;
		transition: all 0.3s ease;
		display: inline-block;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.btn-primary {
		background: #ff4757;
		color: white;
		box-shadow: 0 4px 15px rgba(255, 71, 87, 0.4);
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(255, 71, 87, 0.6);
	}

	.btn-secondary {
		background: rgba(255,255,255,0.2);
		color: white;
		border: 2px solid white;
		backdrop-filter: blur(10px);
	}

	.btn-secondary:hover {
		background: white;
		color: #4ecdc4;
		transform: translateY(-2px);
	}

	/* Sections */
	section {
		padding: 4rem 0;
	}

	.vision {
		background: #f5e6d3;
		text-align: center;
	}

	.vision h2 {
		font-family: 'Bungee', cursive;
		font-size: 2.5rem;
		color: #ff4757;
		margin-bottom: 1rem;
		text-shadow: 3px 3px 0px rgba(0,0,0,0.3);
		transform: rotate(-1deg);
	}

	.vision p {
		font-size: 1.2rem;
		max-width: 800px;
		margin: 0 auto 2rem;
	}

	.philosophy {
		background: linear-gradient(135deg, #8b6914, #d4a574);
		color: white;
		padding: 2rem;
		border-radius: 15px;
		font-size: 1.1rem;
		max-width: 600px;
		margin: 0 auto;
		box-shadow: 0 4px 15px rgba(139, 105, 20, 0.3);
	}

	.features {
		background: #e8dcc6;
	}

	.features h2 {
		font-family: 'Fredoka One', cursive;
		text-align: center;
		font-size: 2.5rem;
		color: #6b4e00;
		margin-bottom: 3rem;
		text-shadow: 2px 2px 0px rgba(212,165,116,0.7);
	}

	.features-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 2rem;
	}

	.feature {
		background: #f7efdc;
		padding: 2rem;
		border-radius: 15px;
		text-align: center;
		box-shadow: 0 4px 15px rgba(139, 105, 20, 0.2);
		transition: transform 0.3s ease;
		border: 1px solid #d4a574;
	}

	.feature:hover {
		transform: translateY(-5px);
	}

	.feature-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
	}

	.feature h3 {
		color: #8b6914;
		margin-bottom: 1rem;
	}

	.status {
		background: #f0e4d2;
	}

	.status h2 {
		font-family: 'Bangers', cursive;
		text-align: center;
		font-size: 2.5rem;
		color: #4ecdc4;
		margin-bottom: 3rem;
		text-shadow: 3px 3px 0px rgba(0,0,0,0.2);
		transform: rotate(1deg);
	}

	.status-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: 3rem;
	}

	.status-working, .status-coming {
		padding: 2rem;
		border-radius: 15px;
	}

	.status-working {
		background: linear-gradient(135deg, #2ed573, #7bed9f);
		color: white;
	}

	.status-coming {
		background: linear-gradient(135deg, #ffa726, #ffcc02);
		color: #333;
	}

	.status-working h3, .status-coming h3 {
		font-size: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.status-working ul, .status-coming ul {
		list-style: none;
		padding: 0;
	}

	.status-working li, .status-coming li {
		padding: 0.5rem 0;
		font-size: 1.1rem;
	}

	.installation {
		background: #ede0c8;
	}

	.installation h2 {
		font-family: 'Bungee', cursive;
		text-align: center;
		font-size: 2.5rem;
		color: #ff4757;
		margin-bottom: 3rem;
		text-shadow: 3px 3px 0px rgba(0,0,0,0.3);
		animation: wiggle 3s ease-in-out infinite;
	}

	@keyframes wiggle {
		0%, 100% { transform: rotate(-1deg); }
		50% { transform: rotate(1deg); }
	}

	.install-steps {
		max-width: 800px;
		margin: 0 auto;
	}

	.step {
		background: #f9f2e7;
		padding: 2rem;
		border-radius: 15px;
		margin-bottom: 2rem;
		box-shadow: 0 4px 15px rgba(139, 105, 20, 0.15);
		border: 1px solid #d4a574;
	}

	.step h3 {
		color: #4ecdc4;
		margin-bottom: 1rem;
		font-size: 1.5rem;
	}

	.code-block {
		position: relative;
		background: #1a1a1a;
		color: #00ff00;
		padding: 1.5rem;
		border-radius: 10px;
		overflow-x: auto;
		font-family: 'Courier New', monospace;
	}

	.code-block pre {
		margin: 0;
		white-space: pre-wrap;
	}

	.copy-btn {
		position: absolute;
		top: 1rem;
		right: 1rem;
		background: #4ecdc4;
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 5px;
		cursor: pointer;
		font-size: 0.9rem;
		transition: background-color 0.3s ease;
	}

	.copy-btn:hover {
		background: #45b7b8;
	}

	.step-note {
		margin-top: 1rem;
		font-style: italic;
		color: #666;
	}

	footer {
		background: linear-gradient(135deg, #5d4e37, #3e2723);
		color: #f5e6d3;
		padding: 3rem 0 1rem;
	}

	.footer-content {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 3rem;
		margin-bottom: 2rem;
	}

	.footer-main h3 {
		font-family: 'Fredoka One', cursive;
		color: #4ecdc4;
		margin-bottom: 1rem;
		text-shadow: 2px 2px 0px rgba(0,0,0,0.3);
	}

	.footer-main ul {
		list-style: none;
		padding: 0;
	}

	.footer-main li {
		padding: 0.3rem 0;
	}

	.footer-links {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.footer-links a {
		color: #4ecdc4;
		text-decoration: none;
		font-weight: 600;
		transition: color 0.3s ease;
	}

	.footer-links a:hover {
		color: #7bed9f;
	}

	.footer-bottom {
		text-align: center;
		padding-top: 2rem;
		border-top: 1px solid #8b6914;
		font-weight: 600;
		font-size: 1.1rem;
	}

	@media (max-width: 768px) {
		.hero-actions {
			flex-direction: column;
			align-items: center;
		}

		.features-grid {
			grid-template-columns: 1fr;
		}

		.status-grid {
			grid-template-columns: 1fr;
		}

		.footer-content {
			grid-template-columns: 1fr;
			gap: 2rem;
		}

		.container {
			padding: 0 1rem;
		}
	}
</style>
