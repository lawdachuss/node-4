const REPOS = [
	'node-1', 'node-2', 'node-3', 'node-4', 'node-5',
	'node-6', 'node-7', 'node-8', 'node-9', 'node-10',
];
const OWNER = 'lawdachuss';
const WORKFLOW_NAME = 'Secure RDP';

export default {
	async scheduled(event, env, ctx) {
		const token = env.GITHUB_TOKEN;
		if (!token) {
			console.error('GITHUB_TOKEN not set');
			return;
		}

		const headers = {
			Authorization: `Bearer ${token}`,
			'User-Agent': 'workflow-watchman/1.0',
			Accept: 'application/vnd.github.v3+json',
		};

		for (const repo of REPOS) {
			try {
				const needsRestart = await checkRepo(repo, headers);
				if (needsRestart) {
					await triggerWorkflow(repo, headers);
					console.log(`[${repo}] triggered new workflow`);
				} else {
					console.log(`[${repo}] OK — no action needed`);
				}
			} catch (err) {
				console.error(`[${repo}] error: ${err.message}`);
			}
		}
	},

	async fetch(request, env, ctx) {
		const url = new URL(request.url);
		if (url.pathname === '/__health') {
			return new Response('OK', { status: 200 });
		}
		return new Response('Not found', { status: 404 });
	},
};

async function checkRepo(repo, headers) {
	// Get the latest workflow run for the workflow
	const runsUrl = `https://api.github.com/repos/${OWNER}/${repo}/actions/workflows/${WORKFLOW_NAME}/runs?per_page=5`;
	const resp = await fetch(runsUrl, { headers });
	if (!resp.ok) {
		throw new Error(`runs API: ${resp.status} ${await resp.text()}`);
	}

	const data = await resp.json();
	const runs = data.workflow_runs || [];
	if (runs.length === 0) {
		return true; // never ran — trigger one
	}

	const latest = runs[0];

	// If there's an in-progress or queued run, don't double-trigger
	if (latest.status === 'in_progress' || latest.status === 'queued' || latest.status === 'pending') {
		return false;
	}

	// If already completed successfully, skip
	if (latest.conclusion === 'success') {
		return false;
	}

	// If user cancelled, skip — don't restart cancelled workflows
	if (latest.conclusion === 'cancelled') {
		return false;
	}

	// Failure, timed_out, stale, start_up_failure etc. → restart
	// But only if the last run is old enough (avoid rapid re-trigger loops)
	const runAge = Date.now() - new Date(latest.run_started_at).getTime();
	const minAge = 10 * 60 * 1000; // 10 minutes

	if (runAge < minAge) {
		console.log(`[${repo}] last run failed ${Math.round(runAge/1000)}s ago — waiting before retry`);
		return false;
	}

	return true;
}

async function triggerWorkflow(repo, headers) {
	const url = `https://api.github.com/repos/${OWNER}/${repo}/actions/workflows/${WORKFLOW_NAME}/dispatches`;
	const body = { ref: 'main' };

	const resp = await fetch(url, {
		method: 'POST',
		headers: {
			...headers,
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(body),
	});

	if (!resp.ok) {
		const text = await resp.text();
		throw new Error(`dispatch API: ${resp.status} ${text}`);
	}
}
