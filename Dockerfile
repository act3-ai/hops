# hadolint ignore=DL3007
FROM chainguard/wolfi-base:latest

# Build depends on a built binary
ARG HOPS_EXECUTABLE=bin/hops

# Copy in the hops executable
COPY ${HOPS_EXECUTABLE} /bin/hops

# Add the hops executable to the PATH
ENV PATH="/bin:${PATH}"

# Add labels
LABEL org.opencontainers.image.title=hops
LABEL org.opencontainers.image.description="Hops is a Homebrew Bottle installer with a focus on performance and mobility."
LABEL org.opencontainers.image.licenses=MIT
LABEL org.opencontainers.image.source=https://github.com/act3-ai/hops
LABEL org.opencontainers.image.documentation=https://github.com/act3-ai/hops
LABEL org.opencontainers.image.url=https://github.com/act3-ai/hops
