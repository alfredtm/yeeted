# yeeted

**Yeeting Go binaries into containers.** A demo Go REST API showing what happens when you stop building images you've already built and start treating the OCI registry as a content-addressed cache.

Same Go service. Three CI workflows. Every push.

## The numbers

| Workflow | Cache miss (real build) | No-op push (cache hit) |
|---|---|---|
| `build-docker.yml` (Dockerfile + BuildKit + GHA cache) | **66s** | 66s |
| `build-ko.yml` (Dockerfile-less, `ko publish`) | **20s** | 20s |
| `build-crane.yml` ([`yeet-cache-action@v2`](https://github.com/alfredtm/yeet-cache-action)) | **22s** | **~16s wall / ~1.5s actual work** |

`yeet-cache-action` is the only one that skips the build entirely on no-op pushes — the registry already has the image, so it just retags. Everything else still pays the full build cost on every commit, whether the source changed or not.

## How it works

The crane workflow ([`.github/workflows/build-crane.yml`](.github/workflows/build-crane.yml)):

1. Hashes the Go source via the GitHub git API — **no checkout needed yet.**
2. Asks `ghcr.io`: *do you already have an image tagged `:src-<hash>`?* — one HTTP HEAD.
3. **Hit** → retag the cached image to `${{ github.sha }}` and `latest`. Done in ~1.5 seconds of real work.
4. **Miss** → checkout, `go build`, `yeet-pack pack` (build the OCI image in memory + push in one round-trip), GitHub-native attestation in the post hook. Done in ~22 seconds.

The `:src-<hash>` tag is a content address. Two builds with identical source files, build flags, and base image digest produce identical bytes — and the registry already has the answer.

## What's in the repo

```
.
├── cmd/server/main.go         # stdlib http server, graceful shutdown
├── internal/
│   ├── handler/               # chi router, /healthz + /items CRUD
│   ├── model/                 # thread-safe in-memory store
│   ├── metrics/               # Prometheus middleware + /metrics
│   ├── telemetry/             # OTel tracing (no-op without OTLP endpoint)
│   └── db/                    # pgx pool (lazy, optional)
├── Dockerfile                 # multi-stage, distroless — used only by build-docker.yml
├── k8s/                       # deployment + service manifests
└── .github/workflows/
    ├── build-docker.yml       # standard docker buildx + GHA cache
    ├── build-ko.yml           # ko publish
    └── build-crane.yml        # yeet-cache-action@v2 + yeet-pack
```

The Go service itself isn't the point — it has realistic dependencies (`chi`, `pgx`, `prometheus`, `otel`) so the build is non-trivial. The point is the **three workflows running side-by-side**, with timing notices on every key step. Open the Actions tab, watch them race.

## Cool, how do I steal this?

The reusable bit is [`alfredtm/yeet-cache-action`](https://github.com/alfredtm/yeet-cache-action) — drop it in front of your existing build pipeline (docker buildx, ko, kaniko, whatever) and it handles the cache check + retag + attestation. See [its README](https://github.com/alfredtm/yeet-cache-action#readme) for the full story.

This repo is the reference example. Use it as a template.

## License

MIT.
