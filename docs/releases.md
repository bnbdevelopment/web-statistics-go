# Release process

`.github/workflows/release.yaml` runs for each push to `main`. It replaces the two former image workflows and preserves the existing Google Artifact Registry location.

## Component selection

| Changed path | Versioned and built artifacts |
| --- | --- |
| `backend/**` | Backend image and Helm chart |
| `frontend/**` | Frontend image and Helm chart |
| `charts/web-statistics/**` | Helm chart only |
| Any combination | Union of the rows above; each artifact is released once |
| Other paths | No application release |

The chart is downstream of both images because its default values pin their immutable tags. A chart change does not cause either image to rebuild.

The script uses Conventional Commit markers across the pushed commits:

- `BREAKING CHANGE:` or `type!:` increments the major version.
- `feat:` increments the minor version.
- Every other application change increments the patch version.

Persistent versions live in `backend/VERSION`, `frontend/package.json` plus its lock file, and `charts/web-statistics/Chart.yaml`. The workflow commits only affected version files and chart image tags back to `main`. Its `[skip cd]` commit marker prevents a second release run. Push conflicts are retried against the latest `main` version up to three times.

Published tags and GitHub Releases use component scoped names, such as `backend-v1.4.2`, `frontend-v2.0.0`, and `chart-v0.8.1`. GitHub generates the notes from commits.

## Repository setup

Configure these settings before enabling the workflow:

1. Add `SERVICE_ACCOUNT_KEY` as a repository Actions secret. The Google service account needs Artifact Registry Writer on `europe-west3-docker.pkg.dev/bnbdevelopment/webstats`.
2. Set **Settings → Actions → General → Workflow permissions** to read and write. The workflow needs `contents: write` for its version commit, tags, and GitHub Releases.
3. If `main` has branch protection, allow GitHub Actions to push the release commit or exempt the workflow identity from the matching ruleset.
4. Ensure the Artifact Registry Docker repository named `webstats` already exists. Helm OCI artifacts and Docker images share this repository.

Image artifacts are published as:

```text
europe-west3-docker.pkg.dev/bnbdevelopment/webstats/backend:<version>
europe-west3-docker.pkg.dev/bnbdevelopment/webstats/frontend:<version>
```

The Helm chart is published as:

```text
oci://europe-west3-docker.pkg.dev/bnbdevelopment/webstats/web-statistics
```

The image jobs use BuildKit cache, generate provenance and SBOM attestations, and publish both an immutable SemVer tag and `latest`. The chart packages only after all affected image builds succeed. GitHub Releases are created only after every selected artifact has been published successfully.
