# hadolint ignore=DL3007
FROM chainguard/wolfi-base:latest

# Build depends on a built binary
ARG HOPS_EXECUTABLE=bin/hops

# Copy in the hops executable
COPY ${HOPS_EXECUTABLE} /bin/hops

# Add the hops executable to the PATH
ENV PATH="/bin:${PATH}"
