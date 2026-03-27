Here’s a clean, “just run it” way to fire up **aider** inside Docker against a specific local repo folder, feed it an instructions file, and only ping you when it’s finished or stuck.

# 1) Pick the Docker image & mount your repo

Aider ships official images. The **core** image is small; the **full** image includes extras (browser GUI, Playwright helpers). Both run fine headless. ([aider.chat][1])

```bash
# from the root of YOUR git repo
git config user.email "you@example.com"
git config user.name "Your Name"

# choose one (core or full):
docker pull paulgauthier/aider
# or: docker pull paulgauthier/aider-full

# .env holds your keys (see section 2)
# instructions.md holds your task(s)
# Mount your repo to /app (aider expects to run at the repo root)
docker run --rm -it \
  --name aider-job \
  --user $(id -u):$(id -g) \
  --workdir /app \
  --env-file .env \
  -v "$(pwd)":/app \
  paulgauthier/aider \
  --model gpt-4o \
  --message-file instructions.md \
  --yes-always \
  --auto-commits \
  --notifications
```

**Why these flags?**

* `--message-file` lets you pass a file of natural-language instructions and then **aider exits when done** (non-interactive). ([aider.chat][2])
* `--yes-always` auto-accepts prompts so it can run unattended. ([aider.chat][3])
* `--auto-commits` commits edits automatically. ([aider.chat][4])
* `--notifications` makes aider alert you when it’s waiting/finished; you can also supply a custom notifier (Slack/Discord/etc.). ([aider.chat][5])

> Heads-up: run the container **from your repo root** so aider sees the git repo and files; the image docs assume your code is mounted at `/app`. ([aider.chat][1])

# 2) API keys & model selection (env)

Aider accepts keys via **env vars**, `.env`, command-line, or YAML config. Easiest in Docker is a **.env file** next to your repo and pass `--env-file .env`. ([aider.chat][6])

Minimal `.env` examples (pick what you use):

```dotenv
# OpenAI
OPENAI_API_KEY=sk-...

# Anthropic
ANTHROPIC_API_KEY=sk-ant-...

# Other providers via generic --api-key mechanism (maps to PROVIDER_API_KEY)
# Example for Gemini, Groq, OpenRouter, Bedrock, etc.:
GEMINI_API_KEY=...
GROQ_API_KEY=...
OPENROUTER_API_KEY=...
# For Azure OpenAI, set the usual OpenAI-compatible vars too if needed:
OPENAI_API_BASE=https://your-azure-endpoint.openai.azure.com/
OPENAI_API_TYPE=azure
OPENAI_API_VERSION=2024-02-01
OPENAI_API_DEPLOYMENT_ID=your-deployment
```

* Keys via env or `.env` are first-class; OpenAI/Anthropic can also be given as dedicated flags (`--openai-api-key`, `--anthropic-api-key`). Others work via `--api-key provider=<KEY>` or simply setting `PROVIDER_API_KEY` env vars as above. ([aider.chat][7])
* `.env` load order: home dir → repo root → CWD → `--env-file` (later wins). In Docker we explicitly use `--env-file .env`. ([aider.chat][6])
* The **model** is set with `--model`. You can also set aliases or advanced settings in `.aider.conf.yml` if you like. ([aider.chat][3])

# 3) Your instruction file(s)

You’ve got two good patterns; pick one:

### A) One-shot “do this and exit”

Use `--message-file` to point at a single instruction file (Markdown or text). Aider applies edits, auto-commits, exits. ([aider.chat][2])

```markdown
# instructions.md (example)
Refactor the auth middleware for better error handling and add unit tests.
Target files:
- server/auth/middleware.ts
- server/auth/__tests__/middleware.test.ts

Constraints:
- Keep public API stable.
- Aim for >85% coverage in the new tests.

Then run eslint and fix issues if trivial.
```

Optionally pre-scope files with `--file` (editable) or `--read` (context only):

```bash
docker run --rm -it --name aider-job \
  --user $(id -u):$(id -g) --workdir /app --env-file .env -v "$(pwd)":/app \
  paulgauthier/aider \
  --model gpt-4o \
  --file server/auth/middleware.ts \
  --read README.md \
  --message-file instructions.md \
  --yes-always --auto-commits --notifications
```

`--file/--read` helps keep the context tight. ([aider.chat][3])

### B) A mini-script of “/commands” + a final instruction

If you want to **pre-load** actions (add files, set mode, pick model), put `/commands` into a text file and pass `--load`, and still use `--message-file` for the task message:

```
# cmds.txt
/model gpt-4o
/add server/auth/middleware.ts
/add server/auth/__tests__/middleware.test.ts
/architect   # (optional) switches edit format for architecture-level changes
```

```bash
docker run --rm -it --name aider-job \
  --user $(id -u):$(id -g) --workdir /app --env-file .env -v "$(pwd)":/app \
  paulgauthier/aider \
  --load cmds.txt \
  --message-file instructions.md \
  --yes-always --auto-commits --notifications
```

`--load` executes slash-commands from a file on launch; `--message-file` is the natural-language task. ([aider.chat][3])

# 4) “Only notify me if done or stuck”

* **Done / waiting**: `--notifications` triggers a desktop/terminal notification when aider finishes a response and is waiting for input (useful even with `--message-file`). For remote alerts, pair `--notifications-command` with something like **Apprise** to send Slack/Discord/Pushbullet pings. ([aider.chat][5])
* **Stuck / long-running**: Set a **timeout** (`--timeout SECONDS`) so the process errors out instead of hanging; your wrapper can treat a non-zero exit as “stuck.” You can also disable streaming (`--no-stream`) if you want the log quiet until completion. (Options ref lists `--timeout`, `--stream`.) ([aider.chat][3])

Example Slack ping on finish:

```bash
docker run --rm -it --name aider-job \
  --user $(id -u):$(id -g) --workdir /app --env-file .env -v "$(pwd)":/app \
  paulgauthier/aider \
  --model gpt-4o \
  --message-file instructions.md \
  --yes-always --auto-commits \
  --timeout 900 \
  --notifications-command "apprise -b 'aider finished (repo: $(basename $(pwd)))' 'slack://$SLACK_WEBHOOK_TOKEN'"
```

([aider.chat][5])

# 5) Optional: YAML config for defaults (nice for CI/cron)

Drop `.aider.conf.yml` in your repo root to bake in defaults (model, auto-commits, notifications, lint/test commands, etc.). ([aider.chat][4])

```yaml
# .aider.conf.yml (sample bits)
model: gpt-4o
auto-commits: true
yes-always: true
notifications: true
test-cmd: "npm test -- --watch=false"
auto-test: false          # set true to run tests after changes
lint-cmd:
  - "javascript: eslint . --fix"
auto-lint: true
```

> Note: Only **OpenAI & Anthropic** keys may live in this YAML; other providers belong in `.env`. ([aider.chat][4])

# 6) Docker Compose (nice for multiple projects)

```yaml
# docker-compose.yml
services:
  aider:
    image: paulgauthier/aider
    working_dir: /app
    user: "${UID:-1000}:${GID:-1000}"
    env_file: .env
    volumes:
      - ./:/app
    command: >
      --model gpt-4o
      --message-file instructions.md
      --yes-always
      --auto-commits
      --notifications
```

Run with:

```bash
docker compose run --rm aider
```

# 7) Day-2 tips

* If you rely on `/run` to execute tests, remember they will execute **inside the container**, not your host env—adjust tools/paths accordingly. ([aider.chat][1])
* For scripted bulk ops, you can iterate files and call aider with `--message` per file (or one `instructions.md` per task). ([aider.chat][2])
* If you need purely non-interactive logs, add `--no-stream` and pipe `docker logs aider-job` when it exits. Options ref lists `--stream/--no-stream`. ([aider.chat][2])

---

## Quick checklist (env & commands)

**.env**

* `OPENAI_API_KEY=` / `ANTHROPIC_API_KEY=` (and/or other `*_API_KEY`s) ([aider.chat][7])
* (Optional) provider-specific settings like `OPENAI_API_BASE` for Azure. ([aider.chat][3])

**.aider.conf.yml (optional)**
Defaults for `model`, `auto-commits`, `yes-always`, `notifications`, `test-cmd`, `lint-cmd`. ([aider.chat][4])

**Run**

```bash
docker run --rm -it \
  --user $(id -u):$(id -g) \
  --workdir /app \
  --env-file .env \
  -v "$(pwd)":/app \
  paulgauthier/aider \
  --model gpt-4o \
  --message-file instructions.md \
  --yes-always --auto-commits \
  --notifications --timeout 900
```

([aider.chat][1])

If you want, I can tailor a tiny `cmds.txt` + `instructions.md` for your **cowork** workflow (scan issue → branch → implement → PR) so it’s fully “triggerable” from your Go runner.

[1]: https://aider.chat/docs/install/docker.html "Aider with docker | aider"
[2]: https://aider.chat/docs/scripting.html "Scripting aider | aider"
[3]: https://aider.chat/docs/config/options.html "Options reference | aider"
[4]: https://aider.chat/docs/config/aider_conf.html "YAML config file | aider"
[5]: https://aider.chat/docs/usage/notifications.html "Notifications | aider"
[6]: https://aider.chat/docs/config/dotenv.html?utm_source=chatgpt.com "Config with .env"
[7]: https://aider.chat/docs/config/api-keys.html?utm_source=chatgpt.com "API Keys"
