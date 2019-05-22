# Contributing to Goa

Thank you for your interest in contributing to the Goa project! We appreciate
contributions via submitting Github issues and/or pull requests.

Below are some guidelines to follow when contributing to this project:

* Before opening an issue in Github, check [open issues](https://github.com/goadesign/goa/issues)
  and [pull requests](https://github.com/goadesign/goa/pulls) for existing
  issues and fixes.
* If your issue has not been addressed, [open a Github issue](https://github.com/goadesign/goa/issues/new)
  and follow the checklist presented in the issue description section. A simple
  Goa design that reproduces your issue helps immensely.
* If you know how to fix your bug, we highly encourage PR contributions. See
  [How Can I Get Started section](#how-can-i-get-started?) on how to submit a PR.
* For feature requests and submitting major changes, [open an issue](https://github.com/goadesign/goa/issues/new)
  or hop on to our slack channel (see https://goa.design to join) to discuss
  the feature first.
* Keep conversations friendly! Constructive criticism goes a long way.
* Have fun contributing!

## How Can I Get Started?

1) Visit https://goa.design for more information on Goa and the getting started
guide.
2) To get your hands dirty, fork the Goa repo and issue PRs from the fork.
**PRO Tip:** Add a [git remote](https://git-scm.com/docs/git-remote.html) to
your forked repo in the Goa source code (in $GOPATH/src/goa.design/goa when
installed using `go get`) to avoid messing with import paths while testing
your fix.
3) [Open issues](https://github.com/goadesign/goa/issues) labeled as `good first
issue` are ideal to understand the source code and make minor contributions.
Issues labeled `help wanted` are bugs/features that are not currently being
worked on and contributing to them are most welcome.
4) Link the issue that the PR intends to solve in the PR description. If an issue
does not exist, adding a description in the PR that describes the issue and the
fix is recommended.
5) Making changes to Goa can sometimes break [goa plugins](https://github.com/goadesign/plugins)
or change the generated [goa examples](https://github.com/goadesign/examples).
Run `make test-plugins` and `make test-examples` to see the failures. To fix
such failures, create a branch in [plugins](https://github.com/goadesign/plugins)
and/or [examples](https://github.com/goadesign/examples) repo with the same
name as the branch in [goa](https://github.com/goadesign/goa) repo and fix the
failures. Re-run the above make commands to verify your fix. Don't forget to
issue PRs for the plugin and example changes if any! Linking the plugins and
examples PR to the main Goa PR makes it easier to understand the changes.
6) Ensure the CI build passes when you issue a PR to Goa.
7) Join our slack channel (see https://goa.design to join) and participate in the
conversations.
