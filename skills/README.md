# Goa Agent Skills

This directory contains reusable Agent Skill packages for application repositories
that use Goa. They are not contributor instructions for changing Goa itself.

## Available Skills

- [`goa-service-designer/`](goa-service-designer/): design-first workflow for
  creating and evolving Goa services, including DSL changes, generated code,
  HTTP/gRPC/JSON-RPC mappings, errors, interceptors, and downstream consumers.

## Install in one command

Run this in your application repository:

```bash
npx skills add goadesign/goa --skill goa-service-designer
```

The [Skills CLI](https://github.com/vercel-labs/skills) requires Node.js and npm.
It downloads the complete skill, including its reference files, and lets you
choose your coding tool. Installation is project-local by default.

Select tools explicitly for a non-interactive installation:

```bash
npx --yes skills add goadesign/goa --skill goa-service-designer \
  -a codex -a cursor -a claude-code --yes
```

Add `--global` for a personal installation. Add `--copy` when your environment
requires copies instead of symlinks. The installer selects only
`goa-service-designer`; contributor and release skills are not installed.

## Give your agent a task

```text
Use the goa-service-designer skill to add a catalog service with HTTP and
gRPC endpoints. Generate the code, implement product lookup, and test it.
```

The skill guides the agent to inspect the application, change the design first,
regenerate contracts, implement outside `gen/`, update affected consumers, and
verify the result. It covers service contracts, transports, validation, errors,
and interceptors. It is intended for Goa application services, not for editing
Goa's compiler or runtime.

For Goa-AI applications, also give the coding agent the generated
`AGENTS_QUICKSTART.md`. That guide reflects the application's agent design;
this reusable skill focuses on Goa services.

## Install without Node.js

Copy the entire `goa-service-designer/` directory, including `SKILL.md` and its
reference files, into your coding tool's project skill directory. For tools
supporting the shared Agent Skills directory, the result is:

```text
.agents/skills/goa-service-designer/SKILL.md
```

For a tool using its own skill directory, such as `.claude/skills/`, copy the
same complete directory there. Consult your tool's skill documentation for its
supported locations. Do not copy only `SKILL.md`; its references are part of
the skill.

See the [coding-agent workflow](https://goa.design/docs/ai-development/) for
context selection, generated-code ownership, a complete service/agent example,
and verification.
