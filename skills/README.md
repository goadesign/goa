# Goa Agent Skills

This directory contains reusable Agent Skill packages for application repositories
that use Goa. They are not contributor instructions for changing Goa itself.

## Available Skills

- [`goa-service-designer/`](goa-service-designer/): design-first workflow for
  creating and evolving Goa services, including DSL changes, generated code,
  HTTP/gRPC mappings, errors, interceptors, and downstream consumers.

## How To Use

Copy the whole skill directory into the place your agent expects skills. Keep the
directory name and `SKILL.md` file together.

### Claude Code

Project-local install:

```bash
mkdir -p .claude/skills
cp -R path/to/goa/skills/goa-service-designer .claude/skills/
```

Personal install:

```bash
mkdir -p ~/.claude/skills
cp -R path/to/goa/skills/goa-service-designer ~/.claude/skills/
```

Claude Code discovers `.claude/skills/<skill-name>/SKILL.md` automatically. You
can let the model invoke the skill from its description, or invoke it directly as
`/goa-service-designer`.

### Cursor

For Cursor project skills, copy the directory to:

```text
.cursor/skills/goa-service-designer/SKILL.md
```

### Codex

For repository-scoped Codex skills, copy the directory to:

```bash
mkdir -p .agents/skills
cp -R path/to/goa/skills/goa-service-designer .agents/skills/
```

Codex discovers `.agents/skills/<skill-name>/SKILL.md` automatically from the
current working directory up to the repository root. You can let Codex invoke the
skill from its description, or invoke it explicitly from the CLI/IDE with
`$goa-service-designer`.

### Claude.ai Or Claude API

Package `goa-service-designer/` as a custom skill using the workflow described in
Anthropic's Agent Skills docs. The ZIP should contain the skill directory, with
`SKILL.md` inside that directory.

### Generic Setup

Install the whole `goa-service-designer/` directory as a skill. If your tool only
supports persistent instructions, reference `goa-service-designer/SKILL.md` from
those instructions.
