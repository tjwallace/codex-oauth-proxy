# Releases

Release Please manages version/changelog PRs using its default notes generator.
Merging a release PR creates an unprefixed `X.Y.Z` tag and publishes:

- Multi-platform images to `ghcr.io/tjwallace/codex-oauth-proxy:X.Y.Z`.
- Go binary archives and checksums to this fork's GitHub Release via GoReleaser.
- `@tjwallace/codex-oauth-proxy` to npm, after the Go assets and installer are verified.

## Setup

- Install a GitHub App with Contents, Issues, and Pull requests write permissions.
  Set Actions secrets `BOT_CLIENT_ID` and `BOT_APP_PRIVATE_KEY` so generated PRs
  and tags trigger downstream workflows.
- Give this repository Actions access to the GHCR package and make it public for
  unauthenticated pulls. Publication uses `GITHUB_TOKEN`; no registry secret is needed.
- Bootstrap the scoped npm package once from the first release tag, then configure
  its trusted publisher for `tjwallace/codex-oauth-proxy`, workflow `release.yml`,
  environment `npm`, allowing `npm publish`. Create the matching GitHub environment.
- Leave GitHub immutable releases disabled so GoReleaser can attach assets after
  Release Please creates the release. Retry failures with **Re-run failed jobs**.

## References

- [Release Please Action and App setup](https://github.com/googleapis/release-please-action#readme)
- [GoReleaser GitHub Actions integration](https://goreleaser.com/ci/actions/)
- [npm trusted publishing](https://docs.npmjs.com/trusted-publishers/)
- [Image tags and usage](README.md#container-image)
