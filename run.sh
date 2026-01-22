#!/usr/bin/env bash
set -euo pipefail

# ------------ Configuration ------------
IMAGE="${XRAI_IMAGE:-ghcr.io/nenorrell/xrai:latest}"

# ------------ Parse args ------------
# All arguments are passed through to xrai
# Example: ./run.sh generate --dsn "postgres://..." --output /output

# ------------ Usage ------------
usage() {
    cat <<EOF
xrai - LLM-optimized database schema introspection for PostgreSQL

Usage:
  ./run.sh generate [flags]

Examples:
  # Generate schema to local directory
  ./run.sh generate --dsn "postgres://user:pass@host:5432/db" --output /output

  # With all options
  ./run.sh generate \\
    --dsn "postgres://user:pass@host:5432/db" \\
    --output /output \\
    --schemas public,app \\
    --stats \\
    --include-views \\
    --include-routines

Environment Variables:
  XRAI_IMAGE    Docker image to use (default: ghcr.io/nenorrell/xrai:latest)
  DSN           PostgreSQL connection string (can be used instead of --dsn flag)

Note:
  When using Docker, the output directory should be /output (mounted from host).
  The DSN should reference the database from the container's perspective.
  For local databases, use host.docker.internal instead of localhost.

EOF
    exit 0
}

# Show usage if no args or help requested
if [[ $# -eq 0 ]] || [[ "${1:-}" == "-h" ]] || [[ "${1:-}" == "--help" ]]; then
    usage
fi

# ------------ Determine output mount ------------
# Look for --output or -o flag to determine mount point
OUTPUT_DIR=""
ARGS=("$@")
for ((i=0; i<${#ARGS[@]}; i++)); do
    if [[ "${ARGS[i]}" == "--output" ]] || [[ "${ARGS[i]}" == "-o" ]]; then
        if [[ $((i+1)) -lt ${#ARGS[@]} ]]; then
            HOST_OUTPUT="${ARGS[$((i+1))]}"
            # Convert to absolute path if relative
            if [[ ! "$HOST_OUTPUT" =~ ^/ ]]; then
                HOST_OUTPUT="$(pwd)/$HOST_OUTPUT"
            fi
            OUTPUT_DIR="$HOST_OUTPUT"
            # Replace the output path with /output for container
            ARGS[$((i+1))]="/output"
        fi
        break
    fi
done

# ------------ Build Docker command ------------
DOCKER_ARGS=(
    "run" "--rm" "-it"
)

# Mount output directory if specified
if [[ -n "$OUTPUT_DIR" ]]; then
    mkdir -p "$OUTPUT_DIR"
    DOCKER_ARGS+=("-v" "${OUTPUT_DIR}:/output")
fi

# Pass through DSN environment variable if set
if [[ -n "${DSN:-}" ]]; then
    DOCKER_ARGS+=("-e" "DSN")
fi

# Add host.docker.internal for accessing host services
DOCKER_ARGS+=("--add-host=host.docker.internal:host-gateway")

# Add image and arguments
DOCKER_ARGS+=("$IMAGE" "${ARGS[@]}")

echo "🚀 Running xrai..."
exec docker "${DOCKER_ARGS[@]}"
