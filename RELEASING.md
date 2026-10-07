# Releasing

Codeshot uses semantic version tags: `vX.Y.Z`. The tag is the source of truth;
the release workflow stamps it into the binary, and `codeshot --version`
prints it with the package channel (`deb` or `tar`). The workflow stamps
`internal/cli.channel` into each binary so `codeshot update` can select the
Debian installation flow. Source builds use Go's build information instead.

## Cutting a release

For example, to release 0.1.0:

1. Update the `## [0.1.0] - Work In Progress` section of `CHANGELOG.md` with
   the release notes and replace `Work In Progress` with the release date.
2. Run `runbook run go/check` and commit the changes on main.
3. Tag the commit and push it:

```bash
git tag v0.1.0
git push origin main v0.1.0
```

`.github/workflows/release.yml` runs the same checks as CI, builds static
Linux amd64 binaries and verifies their version. It publishes a GitHub release
with the matching changelog section as its notes and these assets:

- `codeshot_v0.1.0_linux_amd64.tar.gz`, containing the `codeshot` executable
- `codeshot_0.1.0_amd64.deb`, installing the executable in `/usr/bin` and bash
  and fish completion in their system directories
- `checksums.txt`, containing SHA-256 checksums of both packages

The Debian package depends on git and recommends docker.io. Docker must be
available to run the judge; an existing Docker installation also works.
No Windows package is built.

Publication fails if the changelog section is missing or still marked
`Work In Progress`. After releasing, open a new section for the next version.

Before tagging, confirm the release commit is on main, the working tree is
clean, and `runbook run go/check` passes. Afterwards, confirm the workflow
passed and both packages and their checksums are attached to the release.
