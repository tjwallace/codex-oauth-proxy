# Container releases

Release Please manages version/changelog PRs using its default notes generator.
Merging a release PR creates an unprefixed `X.Y.Z` tag. The **Container image**
workflow then publishes `ghcr.io/tjwallace/codex-oauth-proxy:X.Y.Z` for `linux/amd64`
and `linux/arm64`. Releases do not publish binary archives or npm packages.

## Setup

- Install a GitHub App with Contents, Issues, and Pull requests write permissions.
  Set Actions secrets `BOT_CLIENT_ID` and `BOT_APP_PRIVATE_KEY`; the App token
  allows release PRs and tags to trigger downstream workflows.
- Give this repository Actions access to the GHCR package and make it public for
  unauthenticated pulls. Publication uses `GITHUB_TOKEN`; no registry secret is needed.

## References

- [Release Please Action](https://github.com/googleapis/release-please-action#readme)
- [GitHub App token setup](https://github.com/actions/create-github-app-token#readme)
- [Image tags and usage](README.md#container-image)
